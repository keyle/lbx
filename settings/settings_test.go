package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/cfg")
	t.Setenv("XDG_CACHE_HOME", "/tmp/cache")
	p := Resolve()
	if p.Config != filepath.Join("/tmp/cfg", "lbx") {
		t.Errorf("Config = %q", p.Config)
	}
	if p.Cache != filepath.Join("/tmp/cache", "lbx") {
		t.Errorf("Cache = %q", p.Cache)
	}
}

func TestResolveDefault(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")
	t.Setenv("HOME", "/tmp/home")
	p := Resolve()
	if p.Config != filepath.Join("/tmp/home", ".config", "lbx") {
		t.Errorf("Config = %q", p.Config)
	}
	if p.Cache != filepath.Join("/tmp/home", ".cache", "lbx") {
		t.Errorf("Cache = %q", p.Cache)
	}
}

func TestPathsFiles(t *testing.T) {
	p := Paths{Config: filepath.Join("c", "lbx"), Cache: filepath.Join("k", "lbx")}
	if got := p.ConfigFile(); got != filepath.Join(p.Config, "config.toml") {
		t.Errorf("ConfigFile = %q", got)
	}
	if got := p.ThemeFile(); got != filepath.Join(p.Config, "theme.toml") {
		t.Errorf("ThemeFile = %q", got)
	}
	if got := p.FavoritesFile(); got != filepath.Join(p.Config, "favorites.toml") {
		t.Errorf("FavoritesFile = %q", got)
	}
	if got := p.HistoryFile(); got != filepath.Join(p.Cache, "history.json") {
		t.Errorf("HistoryFile = %q", got)
	}
}

func TestLoadMissingUsesDefaults(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "missing.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.WideView != "auto" || cfg.NoHistory || len(cfg.Categories) != 4 {
		t.Errorf("Load missing = %+v, want defaults", cfg)
	}
}

func TestLoadFileOverridesDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("wide_view = \"never\"\nno_history = true\ncategories = [\"active\"]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.WideView != "never" || !cfg.NoHistory {
		t.Errorf("Load = %+v, want wide_view never / no_history true", cfg)
	}
	if len(cfg.Categories) != 1 || cfg.Categories[0] != "active" {
		t.Errorf("Categories = %v", cfg.Categories)
	}
}

func TestOverridesApply(t *testing.T) {
	cfg := Config{Categories: []string{"hottest"}, WideView: "auto"}
	wide := "never"
	noHistory := true
	got := Overrides{WideView: &wide, NoHistory: &noHistory}.Apply(cfg)
	if got.WideView != "never" || !got.NoHistory {
		t.Errorf("Apply = %+v", got)
	}
	if len(got.Categories) != 1 || got.Categories[0] != "hottest" {
		t.Errorf("unset Categories should be untouched, got %v", got.Categories)
	}
}

func TestWriteDefaultConfig(t *testing.T) {
	p := Paths{Config: filepath.Join(t.TempDir(), "lbx")}
	path, err := p.WriteDefaultConfig()
	if err != nil {
		t.Fatal(err)
	}
	if path != p.ConfigFile() {
		t.Errorf("wrote %q, want %q", path, p.ConfigFile())
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.WideView != "auto" || cfg.NoHistory || len(cfg.Categories) != 4 {
		t.Errorf("written config = %+v", cfg)
	}
}
