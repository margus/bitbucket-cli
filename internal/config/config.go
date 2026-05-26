package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Path returns the config file location:
// $XDG_CONFIG_HOME/bb/config.json, defaulting to ~/.config/bb/config.json.
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

// Auth section of the user config.
type Auth struct {
	Username    string `json:"username"`
	AppPassword string `json:"appPassword"`
}

// Config is the on-disk shape. Extra top-level JSON keys are preserved
// on round-trip rather than dropped.
type Config struct {
	Auth  *Auth          `json:"auth,omitempty"`
	Extra map[string]any `json:"-"`
}

// Load reads the config file. A missing file returns an empty Config, not
// an error — callers check whether Auth is populated.
func Load() (*Config, error) {
	b, err := os.ReadFile(Path())
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("reading %s: %w", Path(), err)
	}
	if len(b) == 0 {
		return &Config{}, nil
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", Path(), err)
	}

	cfg := &Config{Extra: map[string]any{}}
	for k, v := range raw {
		if k == "auth" {
			var a Auth
			if err := json.Unmarshal(v, &a); err == nil {
				cfg.Auth = &a
			}
			continue
		}
		var anyVal any
		if err := json.Unmarshal(v, &anyVal); err == nil {
			cfg.Extra[k] = anyVal
		}
	}
	return cfg, nil
}

// Save writes the config back to disk with pretty-printed JSON.
func Save(c *Config) error {
	out := map[string]any{}
	for k, v := range c.Extra {
		out[k] = v
	}
	if c.Auth != nil {
		out["auth"] = c.Auth
	}
	b, err := json.MarshalIndent(out, "", "    ")
	if err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}
	p := Path()
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	if err := os.WriteFile(p, b, 0600); err != nil {
		return fmt.Errorf("writing %s: %w", p, err)
	}
	return nil
}
