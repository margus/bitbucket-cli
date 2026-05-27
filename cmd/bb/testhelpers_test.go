package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/margus/bitbucket-cli/internal/api"
	"github.com/margus/bitbucket-cli/internal/config"
)

// captureStdout temporarily redirects os.Stdout, runs fn, returns what
// was written. NO_COLOR is set so assertions can be plain text.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	t.Setenv("NO_COLOR", "1")

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	orig := os.Stdout
	os.Stdout = w

	done := make(chan struct{})
	var buf bytes.Buffer
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	fn()
	w.Close()
	os.Stdout = orig
	<-done
	return buf.String()
}

// testRepo is the repo every API test pretends to be in. Pass via
// --project so cobra's PersistentPreRunE wires it into util.ProjectURL.
const testRepo = "acme/widgets"

// withTestServer stands up an httptest server and redirects apiClient
// to hit it with a known token. Also sandboxes HOME / XDG_CONFIG_HOME
// so the cobra PersistentPreRunE config-init step doesn't read the
// user's real config. Tests must pass `--project acme/widgets` in args
// (or use runRepoCmd) to skip the git probe.
func withTestServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("BB_AUTH_USERNAME", "")
	t.Setenv("BB_AUTH_APPPASSWORD", "")
	t.Setenv("BB_AUTH_ACCESSTOKEN", "")
	config.Reset()
	t.Cleanup(config.Reset)

	savedClient := apiClient
	apiClient = func() *api.Client {
		return api.NewWithAuth(&config.Auth{AccessToken: "test-token"}, srv.URL)
	}
	t.Cleanup(func() { apiClient = savedClient })

	return srv
}

// resetRootCmd should be called at the start of each test that drives
// rootCmd directly. cobra stashes flag state across Execute() calls,
// and SetArgs doesn't reset prior parses — this clears the slate.
func resetRootCmd(t *testing.T) {
	t.Helper()
	rootCmd.SetArgs(nil)
	// Reset persistent-flag values to their zero defaults. Cobra
	// doesn't clear these between Execute() calls, so leaving
	// --output=json from a previous test would corrupt the next.
	flagProject = ""
	flagTitle = ""
	flagDescription = ""
	flagInteractive = false
	flagOutput = ""
	flagDebug = false
}

// runCmd drives rootCmd with the given args and captures stdout. Useful
// shorthand for the very common "execute and inspect" pattern.
func runCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	resetRootCmd(t)
	var execErr error
	out := captureStdout(t, func() {
		rootCmd.SetArgs(args)
		execErr = rootCmd.Execute()
	})
	return out, execErr
}

// runRepoCmd prepends `--project acme/widgets` so commands that need a
// repo skip the git probe. Use with withTestServer.
func runRepoCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	full := append([]string{"--project", testRepo}, args...)
	return runCmd(t, full...)
}
