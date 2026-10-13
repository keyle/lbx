package theme

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingUsesDefaults(t *testing.T) {
	th, err := Load(filepath.Join(t.TempDir(), "missing.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if th.Tag != "39" || th.Score != "214" || th.Error != "196" {
		t.Errorf("Load missing = %+v, want defaults", th)
	}
}

func TestLoadFileOverridesAndKeepsDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "theme.toml")
	if err := os.WriteFile(path, []byte("tag = \"1\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	th, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if th.Tag != "1" {
		t.Errorf("Tag = %q, want 1", th.Tag)
	}
	if th.Score != "214" {
		t.Errorf("Score = %q, want default 214", th.Score)
	}
}

func TestDefaultTOMLRoundTrips(t *testing.T) {
	path := filepath.Join(t.TempDir(), "theme.toml")
	if err := WriteDefault(path); err != nil {
		t.Fatal(err)
	}
	th, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if th != Default() {
		t.Errorf("round-trip = %+v, want %+v", th, Default())
	}
}
