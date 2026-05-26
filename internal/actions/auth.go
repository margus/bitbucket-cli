package actions

import (
	"fmt"
	"os"

	"github.com/margus/bitbucket-cli/internal/config"
	"github.com/margus/bitbucket-cli/internal/util"
)

type Auth struct{}

func (Auth) Default() string { return "saveLoginInfo" }
func (Auth) Commands() map[string]string {
	return map[string]string{
		"saveLoginInfo": "save",
		"show":          "show",
	}
}
func (Auth) RequireGit() bool { return false }

func (a Auth) Dispatch(method string, args []string) error {
	switch method {
	case "saveLoginInfo":
		return a.save()
	case "show":
		return a.show()
	default:
		return fmt.Errorf("unknown method: %s", method)
	}
}

func (Auth) save() error {
	util.O("This action requires app password:", "yellow")
	util.O("If you don't have a app password you may create by following this link:", "yellow")
	util.O("https://support.atlassian.com/bitbucket-cloud/docs/app-passwords/", "green")

	username := util.Prompt("Username:", "")
	appPassword := util.Prompt("App password:", "")

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	cfg.Auth = &config.Auth{Username: username, AppPassword: appPassword}
	if err := config.Save(cfg); err != nil {
		util.O("Cannot save file to: "+config.Path(), "red")
		return err
	}
	util.O("Auth info saved.", "green")
	return nil
}

func (Auth) show() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.Auth == nil {
		util.O("You have to configure auth info to use this command.", "red")
		util.O(`Run "bb auth" first.`, "yellow")
		os.Exit(1)
	}
	util.O(map[string]string{
		"username":    cfg.Auth.Username,
		"appPassword": cfg.Auth.AppPassword,
	}, "yellow")
	return nil
}
