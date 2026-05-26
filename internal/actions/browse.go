package actions

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/margus/bitbucket-cli/internal/util"
)

type Browse struct{}

func (Browse) Default() string { return "browse" }
func (Browse) Commands() map[string]string {
	return map[string]string{
		"browse": "browse, b",
		"show":   "show, url",
	}
}
func (Browse) RequireGit() bool { return true }

func (b Browse) Dispatch(method string, args []string) error {
	switch method {
	case "browse":
		return b.browse()
	case "show":
		return b.show()
	default:
		return fmt.Errorf("unknown method: %s", method)
	}
}

func (Browse) browse() error {
	repo, err := util.RepoPath()
	if err != nil {
		return err
	}
	url := "https://bitbucket.org/" + repo
	util.O(url, "white")

	var cmd string
	switch runtime.GOOS {
	case "windows":
		cmd = "start"
	case "darwin":
		cmd = "open"
	case "linux":
		cmd = "xdg-open"
	default:
		return fmt.Errorf("cannot get operation system info")
	}
	return exec.Command(cmd, url).Start()
}

func (Browse) show() error {
	repo, err := util.RepoPath()
	if err != nil {
		return err
	}
	util.O("https://bitbucket.org/"+repo, "white")
	return nil
}
