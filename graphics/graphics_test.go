package graphics

import "testing"

func TestKittySupported(t *testing.T) {
	t.Setenv("KITTY_WINDOW_ID", "")
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("TERM", "xterm-256color")
	if KittySupported() {
		t.Error("plain xterm should not advertise kitty graphics")
	}

	t.Setenv("TERM_PROGRAM", "WezTerm")
	if !KittySupported() {
		t.Error("WezTerm should be detected")
	}

	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("KITTY_WINDOW_ID", "1")
	if !KittySupported() {
		t.Error("KITTY_WINDOW_ID should be detected")
	}
}

func TestPlaceholder(t *testing.T) {
	if got := Placeholder(""); got != "[image: image]" {
		t.Errorf("Placeholder(\"\") = %q", got)
	}
	if got := Placeholder("a cat"); got != "[image: a cat]" {
		t.Errorf("Placeholder = %q", got)
	}
}
