package main

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/margus/bitbucket-cli/internal/util"
	"github.com/spf13/cobra"
)

var browseCmd = &cobra.Command{
	Use:               "browse",
	Short:             "Open the current Bitbucket repository in your default browser",
	Aliases:           []string{"b"},
	PersistentPreRunE: requireGit(),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := util.RepoPath()
		if err != nil {
			return err
		}
		url := "https://bitbucket.org/" + repo
		util.O(url, "white")
		opener := openerFor(runtime.GOOS)
		if opener == "" {
			return fmt.Errorf("don't know how to open a browser on %q", runtime.GOOS)
		}
		return exec.Command(opener, url).Start()
	},
}

var browseURLCmd = &cobra.Command{
	Use:               "url",
	Aliases:           []string{"show"},
	Short:             "Print the Bitbucket URL of the current repo (without opening a browser)",
	PersistentPreRunE: requireGit(),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := util.RepoPath()
		if err != nil {
			return err
		}
		util.O("https://bitbucket.org/"+repo, "white")
		return nil
	},
}

func openerFor(goos string) string {
	switch goos {
	case "windows":
		return "start"
	case "darwin":
		return "open"
	case "linux":
		return "xdg-open"
	}
	return ""
}

func init() {
	browseCmd.AddCommand(browseURLCmd)
	rootCmd.AddCommand(browseCmd)
}
