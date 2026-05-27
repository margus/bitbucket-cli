package main

import (
	"github.com/margus/bitbucket-cli/internal/config"
	"github.com/margus/bitbucket-cli/internal/util"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage Bitbucket auth (access token or username + app password)",
}

var authTokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Save a Bitbucket access token (recommended)",
	Long: `Save a Bitbucket workspace/repository/project access token.

Create one in the Bitbucket web UI:
  Workspace settings -> Access tokens -> Create token
  Repository settings -> Access tokens -> Create token

Access tokens are scoped, revocable, and Atlassian's recommended path
for CLI usage. The CLI uses HTTP Bearer auth when a token is set.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		util.O("Paste a Bitbucket access token (workspace, repository, or project scope).", "yellow")
		util.O("Create one at: https://support.atlassian.com/bitbucket-cloud/docs/access-tokens/", "green")
		token := util.Prompt("Token:", "")
		if token == "" {
			util.O("Empty token — aborting.", "red")
			return errSilent
		}
		// Clear any app-password creds so we don't keep stale Basic auth around.
		if err := config.SaveAuth(&config.Auth{AccessToken: token}); err != nil {
			util.O("Cannot save file to: "+config.Path(), "red")
			return err
		}
		util.O("Access token saved.", "green")
		return nil
	},
}

var authSaveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save Bitbucket username and app password (legacy)",
	RunE: func(cmd *cobra.Command, args []string) error {
		util.O("This action requires an app password.", "yellow")
		util.O("Create one at: https://support.atlassian.com/bitbucket-cloud/docs/app-passwords/", "green")
		util.O(`Tip: "bb auth token" is the modern alternative.`, "gray")
		username := util.Prompt("Username:", "")
		appPassword := util.Prompt("App password:", "")
		// Clear any saved access token to keep exactly one auth method active.
		if err := config.SaveAuth(&config.Auth{Username: username, AppPassword: appPassword}); err != nil {
			util.O("Cannot save file to: "+config.Path(), "red")
			return err
		}
		util.O("Auth info saved.", "green")
		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear all saved credentials (both token and app password)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.SaveAuth(&config.Auth{}); err != nil {
			return err
		}
		util.O("Logged out.", "green")
		return nil
	},
}

var authShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show currently saved auth info",
	RunE: func(cmd *cobra.Command, args []string) error {
		a := config.LoadAuth()
		if a == nil {
			util.O("You have to configure auth info to use this command.", "red")
			util.O(`Run "bb auth token" (recommended) or "bb auth save".`, "yellow")
			return errSilent
		}
		out := map[string]string{"method": a.Method()}
		switch a.Method() {
		case "access token":
			out["accessToken"] = a.AccessToken
		case "app password":
			out["username"] = a.Username
			out["appPassword"] = a.AppPassword
		}
		if render(out) {
			return nil
		}
		util.O(out, "yellow")
		return nil
	},
}

func init() {
	authCmd.AddCommand(authTokenCmd, authSaveCmd, authLogoutCmd, authShowCmd)
	rootCmd.AddCommand(authCmd)
}
