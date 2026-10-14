package settings

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

const appName = "lbx"

// Paths holds the resolved config and cache directories.
type Paths struct {
	Config string
	Cache  string
}

// Resolve returns XDG-style paths: $XDG_CONFIG_HOME/lbx and
// $XDG_CACHE_HOME/lbx, falling back to ~/.config/lbx and ~/.cache/lbx. The XDG
// variables take precedence on every platform, so macOS uses ~/.config/lbx
// rather than ~/Library.
func Resolve() Paths {
	return Paths{
		Config: filepath.Join(xdgDir("XDG_CONFIG_HOME", ".config"), appName),
		Cache:  filepath.Join(xdgDir("XDG_CACHE_HOME", ".cache"), appName),
	}
}

func xdgDir(env, fallback string) string {
	if v := os.Getenv(env); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return fallback
	}
	return filepath.Join(home, fallback)
}

// ConfigFile is the path to config.toml.
func (p Paths) ConfigFile() string { return filepath.Join(p.Config, "config.toml") }

// ThemeFile is the path to theme.toml.
func (p Paths) ThemeFile() string { return filepath.Join(p.Config, "theme.toml") }

// FavoritesFile is the path to favorites.toml.
func (p Paths) FavoritesFile() string { return filepath.Join(p.Config, "favorites.toml") }

// HistoryFile is the path to history.json in the cache directory.
func (p Paths) HistoryFile() string { return filepath.Join(p.Cache, "history.json") }

// EnsureConfigDir creates the config directory on demand.
func (p Paths) EnsureConfigDir() error { return os.MkdirAll(p.Config, 0o755) }

// EnsureCacheDir creates the cache directory on demand.
func (p Paths) EnsureCacheDir() error { return os.MkdirAll(p.Cache, 0o755) }

// Config is the user's file-based settings.
type Config struct {
	Categories []string `toml:"categories"`
	Tags       []string `toml:"tags"`
	NoHistory  bool     `toml:"no_history"`
}

// Default returns the built-in configuration.
func Default() Config {
	return Config{
		Categories: []string{"hottest", "newest", "active", "favorites"},
	}
}

// Load reads config.toml, layering file values over the defaults. A missing
// file yields the defaults, not an error.
func Load(path string) (Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return Default(), fmt.Errorf("parse %s: %w", path, err)
	}
	return cfg, nil
}

// defaultConfigTOML is written verbatim by default-config so the file keeps
// its explanatory comments.
const defaultConfigTOML = `# lbx configuration

# Story feeds shown, in order. Known names: hottest, newest, active, favorites.
# Any other name is treated as a tag feed (e.g. "go", "rust").
categories = ["hottest", "newest", "active", "favorites"]

# Extra tag-restricted feeds appended to the rotation.
tags = []

# Disable read-history tracking.
no_history = false
`

// WriteDefaultConfig writes config.toml with the default settings, creating
// the config directory. It returns the path written.
func (p Paths) WriteDefaultConfig() (string, error) {
	if err := p.EnsureConfigDir(); err != nil {
		return "", err
	}
	path := p.ConfigFile()
	if err := os.WriteFile(path, []byte(defaultConfigTOML), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// Overrides carries CLI flag values. A nil field means the flag was not set,
// so the file value (or default) is kept.
type Overrides struct {
	Categories []string
	Tags       []string
	NoHistory  *bool
}

// Apply layers overrides on top of cfg, letting set flags win.
func (o Overrides) Apply(cfg Config) Config {
	if o.Categories != nil {
		cfg.Categories = o.Categories
	}
	if o.Tags != nil {
		cfg.Tags = o.Tags
	}
	if o.NoHistory != nil {
		cfg.NoHistory = *o.NoHistory
	}
	return cfg
}
