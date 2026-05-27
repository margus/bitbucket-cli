package main

import (
	"strings"
	"testing"

	"github.com/margus/bitbucket-cli/internal/config"
	"github.com/margus/bitbucket-cli/internal/util"
)

// withSandboxedConfig is a lighter sibling of withTestServer for tests
// that don't need an HTTP backend — just an isolated config home.
func withSandboxedConfig(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("BB_AUTH_USERNAME", "")
	t.Setenv("BB_AUTH_APPPASSWORD", "")
	t.Setenv("BB_AUTH_ACCESSTOKEN", "")
	config.Reset()
	t.Cleanup(config.Reset)
	return tmp
}

// Override util.PromptFn for the duration of a single test.
func stubPrompt(t *testing.T, answers ...string) {
	t.Helper()
	saved := util.PromptFn
	i := 0
	util.PromptFn = func(q, def string) string {
		if i >= len(answers) {
			return def
		}
		a := answers[i]
		i++
		return a
	}
	t.Cleanup(func() { util.PromptFn = saved })
}

func TestAuthTokenSavesAccessToken(t *testing.T) {
	withSandboxedConfig(t)
	stubPrompt(t, "ATBB-token-xyz")

	out, err := runCmd(t, "auth", "token")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out, "Access token saved") {
		t.Errorf("missing success message in:\n%s", out)
	}
	auth := config.LoadAuth()
	if auth == nil || auth.AccessToken != "ATBB-token-xyz" {
		t.Errorf("AccessToken not persisted: %+v", auth)
	}
}

func TestAuthTokenRejectsEmpty(t *testing.T) {
	withSandboxedConfig(t)
	stubPrompt(t, "")

	out, err := runCmd(t, "auth", "token")
	if err == nil {
		t.Errorf("expected error on empty token, got nil. Output:\n%s", out)
	}
	if !strings.Contains(out, "Empty token") {
		t.Errorf("missing 'Empty token' message in:\n%s", out)
	}
}

func TestAuthSaveStoresUsernameAndPassword(t *testing.T) {
	withSandboxedConfig(t)
	stubPrompt(t, "alice", "s3cret")

	out, err := runCmd(t, "auth", "save")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out, "Auth info saved") {
		t.Errorf("missing success message: %s", out)
	}
	auth := config.LoadAuth()
	if auth == nil || auth.Username != "alice" || auth.AppPassword != "s3cret" {
		t.Errorf("creds not persisted: %+v", auth)
	}
}

func TestAuthSaveClearsAccessToken(t *testing.T) {
	withSandboxedConfig(t)
	// Pre-seed with an access token.
	if err := config.SaveAuth(&config.Auth{AccessToken: "old"}); err != nil {
		t.Fatalf("setup: %v", err)
	}
	config.Reset()

	stubPrompt(t, "alice", "pw")
	if _, err := runCmd(t, "auth", "save"); err != nil {
		t.Fatalf("execute: %v", err)
	}

	config.Reset()
	if err := config.Init(); err != nil {
		t.Fatalf("re-init: %v", err)
	}
	auth := config.LoadAuth()
	if auth == nil || auth.AccessToken != "" {
		t.Errorf("save did not clear AccessToken: %+v", auth)
	}
	if auth == nil || auth.Username != "alice" {
		t.Errorf("save did not persist Username: %+v", auth)
	}
}

func TestAuthShowWithNoConfig(t *testing.T) {
	withSandboxedConfig(t)
	out, err := runCmd(t, "auth", "show")
	if err == nil {
		t.Errorf("expected error when no auth configured, got nil. Output: %s", out)
	}
	if !strings.Contains(out, "configure auth info") {
		t.Errorf("missing hint message: %s", out)
	}
}

func TestAuthShowReportsMethodAccessToken(t *testing.T) {
	withSandboxedConfig(t)
	if err := config.SaveAuth(&config.Auth{AccessToken: "ATBB-xyz"}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	out, err := runCmd(t, "auth", "show")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out, "access token") {
		t.Errorf("show should report active method 'access token': %s", out)
	}
}

func TestAuthShowReportsMethodAppPassword(t *testing.T) {
	withSandboxedConfig(t)
	if err := config.SaveAuth(&config.Auth{Username: "alice", AppPassword: "pw"}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	out, err := runCmd(t, "auth", "show")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out, "app password") {
		t.Errorf("show should report active method 'app password': %s", out)
	}
}

func TestAuthLogout(t *testing.T) {
	withSandboxedConfig(t)
	if err := config.SaveAuth(&config.Auth{AccessToken: "ATBB-xyz"}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	out, err := runCmd(t, "auth", "logout")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out, "Logged out") {
		t.Errorf("missing logout message: %s", out)
	}

	config.Reset()
	if err := config.Init(); err != nil {
		t.Fatalf("re-init: %v", err)
	}
	if auth := config.LoadAuth(); auth != nil {
		t.Errorf("creds not cleared: %+v", auth)
	}
}
