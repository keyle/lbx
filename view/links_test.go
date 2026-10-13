package view

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"noben.org/lbx/render"
)

func TestLobstersStoryID(t *testing.T) {
	cases := []struct {
		href string
		id   string
		ok   bool
	}{
		{"https://lobste.rs/s/abc123/title", "abc123", true},
		{"/s/xyz", "xyz", true},
		{"https://example.com/s/notme", "", false},
		{"https://lobste.rs/u/alice", "", false},
	}
	for _, tc := range cases {
		id, ok := lobstersStoryID(tc.href)
		if ok != tc.ok || id != tc.id {
			t.Errorf("lobstersStoryID(%q) = %q,%v want %q,%v", tc.href, id, ok, tc.id, tc.ok)
		}
	}
}

func TestPickerNavigation(t *testing.T) {
	p := &linkPicker{}
	p.toggle([]render.Link{{Href: "a"}, {Href: "b"}, {Href: "c"}})
	if !p.active {
		t.Fatal("picker should be active")
	}
	p.handleKey(key("j"))
	if p.sel != 1 {
		t.Fatalf("sel = %d, want 1", p.sel)
	}
	p.handleKey(key("k"))
	if p.sel != 0 {
		t.Fatalf("sel = %d, want 0", p.sel)
	}
	p.handleKey(key("tab"))
	if p.active {
		t.Error("tab should close the picker")
	}
}

func TestPickerEnterOpensCommentsForStoryLink(t *testing.T) {
	p := &linkPicker{}
	p.toggle([]render.Link{{Text: "thread", Href: "https://lobste.rs/s/abc123/t"}})
	cmd, handled := p.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if !handled {
		t.Fatal("enter not handled")
	}
	if cmd == nil {
		t.Fatal("no command")
	}
	msg := cmd()
	if _, ok := msg.(openCommentsMsg); !ok {
		t.Fatalf("msg = %T, want openCommentsMsg", msg)
	}
}

func TestPickerEnterOpensReaderForExternalLink(t *testing.T) {
	p := &linkPicker{}
	p.toggle([]render.Link{{Text: "site", Href: "https://example.com/x"}})
	cmd, _ := p.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	msg := cmd()
	if _, ok := msg.(openReaderMsg); !ok {
		t.Fatalf("msg = %T, want openReaderMsg", msg)
	}
}
