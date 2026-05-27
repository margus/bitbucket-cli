package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/margus/bitbucket-cli/internal/api"
	"github.com/margus/bitbucket-cli/internal/config"
	"github.com/margus/bitbucket-cli/internal/util"
	"github.com/spf13/cobra"
)

// apiClient is the constructor every cobra handler calls. The default
// loads auth from viper and prints a friendly hint + exits on missing
// credentials. Tests substitute this with an httptest-backed client.
var apiClient = func() *api.Client {
	c, err := api.NewFromConfig()
	if err == nil {
		return c
	}
	if errors.Is(err, api.ErrNoAuth) {
		util.O("You have to configure auth info to use this command.", "red")
		util.O(`Run "bb auth token" (recommended) or "bb auth save" first.`, "yellow")
	} else {
		util.Errorln("%v", err)
	}
	os.Exit(1)
	return nil // unreachable
}

// Build metadata. Populated via -ldflags at release time (goreleaser does
// this automatically). Dev builds keep the placeholder values.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// versionString formats the three build-metadata vars for the cobra
// --version output. Re-computed via a function so tests that mutate
// `version` see their change without re-initialising rootCmd.
func versionString() string {
	return fmt.Sprintf("%s\n  commit: %s\n  built:  %s", version, commit, date)
}

// Persistent flag values shared with subcommands. Captured on root so
// every subcommand sees them; written into util / shared state in
// PersistentPreRun.
var (
	flagProject     string
	flagTitle       string
	flagDescription string
	flagInteractive bool
	flagDebug       bool
)

var rootCmd = &cobra.Command{
	Use:           "bb",
	Short:         "Bitbucket Cloud REST API CLI",
	Long:          "bb is a small, fast CLI for the Bitbucket Cloud REST API: pull requests, pipelines, branches, deployment environments.",
	Version:       versionString(),
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Bridge persistent flags into the legacy helpers that read them
		// (util.ProjectURL is consulted by RepoPath).
		util.ProjectURL = flagProject

		if err := validateOutput(); err != nil {
			return err
		}

		// Wire --debug into the api package's logger hook.
		if flagDebug {
			api.Debug = func(method, url string, status int, body []byte) {
				fmt.Fprintf(os.Stderr, "[DEBUG] %s %s -> %d (%d bytes)\n", method, url, status, len(body))
			}
		} else {
			api.Debug = nil
		}

		// Initialise viper (loads config file + env vars). Missing file
		// is OK; auth-requiring subcommands fail explicitly later.
		if err := config.Init(); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagProject, "project", "", `repository to work with (e.g. "owner/repo" or "https://bitbucket.org/owner/repo")`)
	rootCmd.PersistentFlags().StringVar(&flagTitle, "title", "", "PR title (used with pr create)")
	rootCmd.PersistentFlags().StringVar(&flagDescription, "description", "", "PR description (used with pr create)")
	rootCmd.PersistentFlags().BoolVarP(&flagInteractive, "interactive", "i", false, "prompt for PR title/description on pr create")
	rootCmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "", "output format: json, yaml, or empty for human-readable (default)")
	rootCmd.PersistentFlags().BoolVar(&flagDebug, "debug", false, "log HTTP requests and responses to stderr")

	// Tell cobra to print a slightly cleaner version line.
	rootCmd.SetVersionTemplate("bb {{.Version}}\n")
}

// requireGit returns a PersistentPreRunE that fails if we're not in a
// git repo and --project wasn't supplied. Applied per-subcommand.
func requireGit() func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Always honor the root's PersistentPreRunE first.
		if rootCmd.PersistentPreRunE != nil {
			if err := rootCmd.PersistentPreRunE(cmd, args); err != nil {
				return err
			}
		}
		if flagProject != "" {
			return nil
		}
		if !util.HasGitDir() {
			util.O("ERROR: No git repository found in current directory.", "red")
			util.O(`Use --project "owner/repo" to work with a remote repository.`, "yellow")
			return fmt.Errorf("not a git repository")
		}
		return nil
	}
}
