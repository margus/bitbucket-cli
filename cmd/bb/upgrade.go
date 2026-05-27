package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/margus/bitbucket-cli/internal/util"
	"github.com/spf13/cobra"
)

// upgradeReleasesURL is the GitHub API endpoint queried for the latest
// release. Package-level so tests can point it at an httptest.Server.
var upgradeReleasesURL = "https://api.github.com/repos/margus/bitbucket-cli/releases/latest"

// upgradeTargetPath resolves the install path of the running binary.
// Override in tests so we don't try to overwrite the test runner.
var upgradeTargetPath = os.Executable

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Self-update by fetching the latest release for this OS/arch",
	RunE: func(cmd *cobra.Command, args []string) error {
		return upgradeRun()
	},
}

func upgradeRun() error {
	req, err := http.NewRequest("GET", upgradeReleasesURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "bb-cli/"+version)
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

	if release.TagName == "" || release.TagName <= version {
		util.O("You are already on the latest version of bb", "green")
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

	binPath, err := upgradeTargetPath()
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
		// Cross-device rename: fall back to copy.
		in, oerr := os.Open(tmp.Name())
		if oerr != nil {
			return err
		}
		defer in.Close()
		out, oerr := os.OpenFile(binPath, os.O_WRONLY|os.O_TRUNC, 0755)
		if oerr != nil {
			return oerr
		}
		if _, cerr := io.Copy(out, in); cerr != nil {
			out.Close()
			return cerr
		}
		out.Close()
		_ = os.Remove(tmp.Name())
	}

	util.O("bb-cli updated", "green")
	return nil
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
