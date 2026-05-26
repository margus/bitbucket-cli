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

// withTempHome points HOME at a fresh tempdir and clears XDG_CONFIG_HOME
// so Path() resolves into the tempdir. Sets USERPROFILE too so this works
// on Windows where os.UserHomeDir reads USERPROFILE, not HOME.
func withTempHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	t.Setenv("XDG_CONFIG_HOME", "")
	return dir
}

func TestLoadEmpty(t *testing.T) {
	withTempHome(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Auth != nil {
		t.Errorf("Load() Auth = %+v, want nil for missing file", cfg.Auth)
	}
}

func TestSaveLoadRoundtrip(t *testing.T) {
	withTempHome(t)
	in := &Config{Auth: &Auth{Username: "alice", AppPassword: "s3cret"}}
	if err := Save(in); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Auth == nil {
		t.Fatal("Load() returned no Auth after Save()")
	}
	if got.Auth.Username != in.Auth.Username || got.Auth.AppPassword != in.Auth.AppPassword {
		t.Errorf("roundtrip Auth = %+v, want %+v", got.Auth, in.Auth)
	}
}

func TestSaveCreatesDir(t *testing.T) {
	home := withTempHome(t)
	// The bb/ subdir should not exist yet.
	if _, err := os.Stat(filepath.Join(home, ".config", "bb")); !os.IsNotExist(err) {
		t.Fatalf("precondition: dir exists, err=%v", err)
	}
	if err := Save(&Config{Auth: &Auth{Username: "a", AppPassword: "p"}}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if _, err := os.Stat(Path()); err != nil {
		t.Errorf("config file missing after Save: %v", err)
	}
}

func TestExtraKeysPreserved(t *testing.T) {
	home := withTempHome(t)
	// Hand-write a config that has an extra top-level key besides auth.
	if err := os.MkdirAll(filepath.Join(home, ".config", "bb"), 0700); err != nil {
		t.Fatalf("setup: %v", err)
	}
	raw := []byte(`{"auth":{"username":"a","appPassword":"p"},"unknown":{"x":1}}`)
	if err := os.WriteFile(Path(), raw, 0600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() = %v", err)
	}
	if _, ok := cfg.Extra["unknown"]; !ok {
		t.Fatalf("Extra keys missing: %+v", cfg.Extra)
	}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save() = %v", err)
	}

	b, err := os.ReadFile(Path())
	if err != nil {
		t.Fatalf("re-read: %v", err)
	}
	if !strings.Contains(string(b), `"unknown"`) {
		t.Errorf("Save() dropped extra keys: %s", b)
	}
}

func TestSaveFilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits don't apply on Windows")
	}
	withTempHome(t)
	if err := Save(&Config{Auth: &Auth{Username: "a", AppPassword: "p"}}); err != nil {
		t.Fatalf("Save() = %v", err)
	}
	info, err := os.Stat(Path())
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	// Credentials file should not be world-readable.
	if info.Mode().Perm()&0o077 != 0 {
		t.Errorf("config file permissions = %o, want owner-only", info.Mode().Perm())
	}
}
