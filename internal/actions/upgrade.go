package actions

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/margus/bitbucket-cli/internal/util"
)

type Upgrade struct {
	CurrentVersion string
}

func (Upgrade) Default() string                 { return "index" }
func (Upgrade) Commands() map[string]string     { return map[string]string{"index": "index"} }
func (Upgrade) RequireGit() bool                { return false }

func (u Upgrade) Dispatch(method string, args []string) error {
	switch method {
	case "index":
		return u.index()
	default:
		return fmt.Errorf("unknown method: %s", method)
	}
}

func (u Upgrade) index() error {
	req, err := http.NewRequest("GET", "https://api.github.com/repos/margus/bitbucket-cli/releases/latest", nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "BB-Cli Curl Agent")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("fetching latest release: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var release struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.Unmarshal(body, &release); err != nil {
		return fmt.Errorf("parsing release: %w", err)
	}

	if release.TagName == "" || release.TagName <= u.CurrentVersion {
		util.O("You are already on the latest version of bb-cli", "green")
		return nil
	}

	wantName := fmt.Sprintf("bb-%s-%s", runtime.GOOS, runtime.GOARCH)
	var dlURL string
	for _, a := range release.Assets {
		if a.Name == wantName || a.Name == wantName+".exe" {
			dlURL = a.BrowserDownloadURL
			break
		}
	}
	if dlURL == "" {
		return fmt.Errorf("no release asset found for %s/%s in %s", runtime.GOOS, runtime.GOARCH, release.TagName)
	}

	util.O(fmt.Sprintf("Fetching new version (%s) ...", release.TagName), "green")

	binPath, err := os.Executable()
	if err != nil {
		return err
	}

	dl, err := http.Get(dlURL)
	if err != nil {
		return fmt.Errorf("downloading: %w", err)
	}
	defer dl.Body.Close()

	tmp, err := os.CreateTemp("", "bb-cli-*")
	if err != nil {
		return err
	}
	if _, err := io.Copy(tmp, dl.Body); err != nil {
		tmp.Close()
		_ = os.Remove(tmp.Name())
		return err
	}
	tmp.Close()

	if err := os.Chmod(tmp.Name(), 0755); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	if err := os.Rename(tmp.Name(), binPath); err != nil {
		// Cross-device rename failure: fall back to copy.
		in, err2 := os.Open(tmp.Name())
		if err2 != nil {
			return err
		}
		defer in.Close()
		out, err2 := os.OpenFile(binPath, os.O_WRONLY|os.O_TRUNC, 0755)
		if err2 != nil {
			return err2
		}
		if _, err2 := io.Copy(out, in); err2 != nil {
			out.Close()
			return err2
		}
		out.Close()
		_ = os.Remove(tmp.Name())
	}

	util.O("BB-CLI Updated", "green")
	return nil
}
