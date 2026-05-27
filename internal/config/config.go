// Package config is a thin viper wrapper for the on-disk JSON config and
// BB_* env-var overrides.
//
// File: $XDG_CONFIG_HOME/bb/config.json (or ~/.config/bb/config.json).
// Env vars:
//   BB_AUTH_USERNAME      → auth.username
//   BB_AUTH_APPPASSWORD   → auth.appPassword
//
// Env vars take precedence over the file, matching viper's standard
// precedence (flag > env > config > default).
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Auth holds the saved Bitbucket credentials.
//
// Two mechanisms are supported and the API client picks one at request
// time, preferring AccessToken when set:
//
//   - AccessToken — Bitbucket workspace/repository/project access token,
//     used as a Bearer token. The modern Atlassian-recommended path.
//   - Username + AppPassword — the legacy HTTP-Basic flow, retained for
//     backwards compatibility.
type Auth struct {
	Username    string `mapstructure:"username" json:"username,omitempty"`
	AppPassword string `mapstructure:"appPassword" json:"appPassword,omitempty"`
	AccessToken string `mapstructure:"accessToken" json:"accessToken,omitempty"`
}

// Method returns a short human-readable name for the active auth
// mechanism, or "" if neither is configured.
func (a *Auth) Method() string {
	if a == nil {
		return ""
	}
	if a.AccessToken != "" {
		return "access token"
	}
	if a.Username != "" {
		return "app password"
	}
	return ""
}

// Path returns the canonical on-disk location of the config file.
func Path() string {
	if base := os.Getenv("XDG_CONFIG_HOME"); base != "" {
		return filepath.Join(base, "bb", "config.json")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}
	return filepath.Join(home, ".config", "bb", "config.json")
}

// Init configures viper. Safe to call multiple times. Returns no error
// when the config file is missing — that's the first-run state.
func Init() error {
	v := viper.GetViper()
	v.SetConfigName("config")
	v.SetConfigType("json")
	v.AddConfigPath(filepath.Dir(Path()))

	v.SetEnvPrefix("BB")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		var nf viper.ConfigFileNotFoundError
		// Pathological-but-real: viper returns os.PathError when the
		// directory is missing, ConfigFileNotFoundError when the file
		// is. Treat both as first-run, not error.
		if !os.IsNotExist(err) && !errors.As(err, &nf) {
			return fmt.Errorf("reading config: %w", err)
		}
	}
	return nil
}

// LoadAuth returns the auth section, with env-var overrides applied.
// Returns nil if neither an access token nor a username is configured.
func LoadAuth() *Auth {
	a := &Auth{
		Username:    viper.GetString("auth.username"),
		AppPassword: viper.GetString("auth.appPassword"),
		AccessToken: viper.GetString("auth.accessToken"),
	}
	if a.AccessToken == "" && a.Username == "" {
		return nil
	}
	return a
}

// SaveAuth writes the auth section to disk, creating parent dirs as
// needed. Env-var overrides do NOT round-trip — only file-backed values
// are persisted. Empty fields are explicitly cleared, so switching
// between auth methods doesn't leave stale credentials in the file.
func SaveAuth(a *Auth) error {
	viper.Set("auth.username", a.Username)
	viper.Set("auth.appPassword", a.AppPassword)
	viper.Set("auth.accessToken", a.AccessToken)

	p := Path()
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	// viper.WriteConfigAs doesn't respect the file mode, so write
	// ourselves through a temp file to keep credentials owner-only.
	tmp, err := os.CreateTemp(filepath.Dir(p), "config-*.json")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmp.Close()
	if err := viper.WriteConfigAs(tmp.Name()); err != nil {
		_ = os.Remove(tmp.Name())
		return fmt.Errorf("writing config: %w", err)
	}
	if err := os.Chmod(tmp.Name(), 0600); err != nil {
		_ = os.Remove(tmp.Name())
		return fmt.Errorf("chmod: %w", err)
	}
	return os.Rename(tmp.Name(), p)
}

// Reset clears viper's in-memory state. Tests only.
func Reset() {
	viper.Reset()
}
