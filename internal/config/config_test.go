package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPathXDG(t *testing.T) {
	base := filepath.Join(string(filepath.Separator), "custom", "xdg")
	t.Setenv("XDG_CONFIG_HOME", base)
	want := filepath.Join(base, "bb", "config.json")
	if got := Path(); got != want {
		t.Errorf("Path() with XDG_CONFIG_HOME = %q, want %q", got, want)
	}
}

func TestPathHomeFallback(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("home-dir resolution on Windows uses USERPROFILE, not HOME, and uses backslash separators")
	}
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "/home/tester")
	got := Path()
	if !strings.HasSuffix(got, "/.config/bb/config.json") {
		t.Errorf("Path() without XDG = %q, want suffix /.config/bb/config.json", got)
	}
}

// withTempHome points HOME (and USERPROFILE for Windows) at a fresh
// tempdir, clears XDG_CONFIG_HOME, and resets viper. Returns the dir.
func withTempHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("BB_AUTH_USERNAME", "")
	t.Setenv("BB_AUTH_APPPASSWORD", "")
	Reset()
	t.Cleanup(Reset)
	return dir
}

func TestLoadAuthMissingReturnsNil(t *testing.T) {
	withTempHome(t)
	if err := Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if got := LoadAuth(); got != nil {
		t.Errorf("LoadAuth() with no file/env = %+v, want nil", got)
	}
}

func TestSaveAuthRoundtrip(t *testing.T) {
	withTempHome(t)
	if err := Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	in := &Auth{Username: "alice", AppPassword: "s3cret"}
	if err := SaveAuth(in); err != nil {
		t.Fatalf("SaveAuth: %v", err)
	}

	// Reset viper to simulate a fresh process: only the file should
	// drive the loaded auth.
	Reset()
	if err := Init(); err != nil {
		t.Fatalf("Init (reload): %v", err)
	}
	got := LoadAuth()
	if got == nil {
		t.Fatal("LoadAuth() returned nil after SaveAuth")
	}
	if got.Username != in.Username || got.AppPassword != in.AppPassword {
		t.Errorf("LoadAuth = %+v, want %+v", got, in)
	}
}

func TestSaveAuthCreatesDir(t *testing.T) {
	home := withTempHome(t)
	if _, err := os.Stat(filepath.Join(home, ".config", "bb")); !os.IsNotExist(err) {
		t.Fatalf("precondition: dir already exists (%v)", err)
	}
	if err := Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := SaveAuth(&Auth{Username: "a", AppPassword: "p"}); err != nil {
		t.Fatalf("SaveAuth: %v", err)
	}
	if _, err := os.Stat(Path()); err != nil {
		t.Errorf("config file missing after SaveAuth: %v", err)
	}
}

func TestSaveAuthFilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits don't apply on Windows")
	}
	withTempHome(t)
	if err := Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := SaveAuth(&Auth{Username: "a", AppPassword: "p"}); err != nil {
		t.Fatalf("SaveAuth: %v", err)
	}
	info, err := os.Stat(Path())
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		t.Errorf("config file permissions = %o, want owner-only", info.Mode().Perm())
	}
}

func TestAccessTokenRoundtrip(t *testing.T) {
	withTempHome(t)
	if err := Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	in := &Auth{AccessToken: "ATBB-xyz"}
	if err := SaveAuth(in); err != nil {
		t.Fatalf("SaveAuth: %v", err)
	}

	Reset()
	if err := Init(); err != nil {
		t.Fatalf("Init (reload): %v", err)
	}
	got := LoadAuth()
	if got == nil {
		t.Fatal("LoadAuth = nil after SaveAuth")
	}
	if got.AccessToken != in.AccessToken {
		t.Errorf("AccessToken = %q, want %q", got.AccessToken, in.AccessToken)
	}
	if got.Method() != "access token" {
		t.Errorf("Method() = %q, want %q", got.Method(), "access token")
	}
}

// Saving an Auth with only AccessToken set must clear any previously
// stored username/appPassword. Otherwise switching auth methods leaves
// stale credentials on disk.
func TestSaveAuthClearsOtherMethod(t *testing.T) {
	withTempHome(t)
	if err := Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	// Start with app-password auth.
	if err := SaveAuth(&Auth{Username: "alice", AppPassword: "pw"}); err != nil {
		t.Fatalf("SaveAuth (app pw): %v", err)
	}
	Reset()
	if err := Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if got := LoadAuth(); got == nil || got.Username != "alice" {
		t.Fatalf("setup failed: %+v", got)
	}

	// Switch to access-token auth.
	if err := SaveAuth(&Auth{AccessToken: "new-token"}); err != nil {
		t.Fatalf("SaveAuth (token): %v", err)
	}
	Reset()
	if err := Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	got := LoadAuth()
	if got == nil {
		t.Fatal("LoadAuth = nil after switch")
	}
	if got.AccessToken != "new-token" {
		t.Errorf("AccessToken = %q, want %q", got.AccessToken, "new-token")
	}
	if got.Username != "" || got.AppPassword != "" {
		t.Errorf("stale app-password creds remain after switch: %+v", got)
	}
}

func TestMethodReturnsActiveMechanism(t *testing.T) {
	tests := []struct {
		auth *Auth
		want string
	}{
		{nil, ""},
		{&Auth{}, ""},
		{&Auth{Username: "u"}, "app password"},
		{&Auth{AccessToken: "t"}, "access token"},
		// Token wins when both are set (matches what the API client does).
		{&Auth{Username: "u", AppPassword: "p", AccessToken: "t"}, "access token"},
	}
	for _, tt := range tests {
		if got := tt.auth.Method(); got != tt.want {
			t.Errorf("Method() on %+v = %q, want %q", tt.auth, got, tt.want)
		}
	}
}

func TestEnvVarOverridesFile(t *testing.T) {
	withTempHome(t)
	if err := Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := SaveAuth(&Auth{Username: "from-file", AppPassword: "file-pass"}); err != nil {
		t.Fatalf("SaveAuth: %v", err)
	}

	// Re-init with env vars set to verify env takes precedence.
	Reset()
	t.Setenv("BB_AUTH_USERNAME", "from-env")
	t.Setenv("BB_AUTH_APPPASSWORD", "env-pass")
	if err := Init(); err != nil {
		t.Fatalf("Init (with env): %v", err)
	}
	got := LoadAuth()
	if got == nil {
		t.Fatal("LoadAuth = nil with env vars set")
	}
	if got.Username != "from-env" || got.AppPassword != "env-pass" {
		t.Errorf("env did not override file: got %+v", got)
	}
}
