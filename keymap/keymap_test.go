package keymap

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func runes(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestAction(t *testing.T) {
	m := Map{
		{Name: "next", Keys: []string{"j", "down"}, Help: "j", Desc: "next"},
		{Name: "quit", Keys: []string{"q"}, Help: "q", Desc: "quit"},
	}
	if got := m.Action(runes("j")); got != "next" {
		t.Errorf("Action(j) = %q, want next", got)
	}
	if got := m.Action(runes("q")); got != "quit" {
		t.Errorf("Action(q) = %q, want quit", got)
	}
	if got := m.Action(runes("z")); got != "" {
		t.Errorf("Action(z) = %q, want empty", got)
	}
}

func TestHelpSkipsHidden(t *testing.T) {
	m := Map{
		{Name: "next", Keys: []string{"j"}, Help: "j", Desc: "next"},
		{Name: "hidden", Keys: []string{"x"}},
	}
	got := m.Help()
	if len(got) != 1 || got[0].Keys != "j" || got[0].Desc != "next" {
		t.Fatalf("Help() = %+v", got)
	}
}
