package view

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"noben.org/lbx/lobsters"
)

func sizeMsg(w, h int) tea.WindowSizeMsg { return tea.WindowSizeMsg{Width: w, Height: h} }

func TestWideAutoLayout(t *testing.T) {
	m := New(Options{Client: lobsters.New()})

	wide, _ := m.Update(sizeMsg(200, 40))
	if !wide.(Model).wide {
		t.Error("expected wide mode at 200 columns")
	}

	narrow, _ := m.Update(sizeMsg(100, 40))
	if narrow.(Model).wide {
		t.Error("expected single pane at 100 columns")
	}
}

func TestWideToggle(t *testing.T) {
	m := New(Options{Client: lobsters.New()})
	sized, _ := m.Update(sizeMsg(200, 40))

	off, _ := sized.(Model).Update(key("z"))
	if off.(Model).wide {
		t.Error("z should turn wide view off")
	}
	on, _ := off.(Model).Update(key("z"))
	if !on.(Model).wide {
		t.Error("z should turn wide view back on")
	}
}

func TestWideViewRenders(t *testing.T) {
	m := New(Options{Client: lobsters.New()})
	sized, _ := m.Update(sizeMsg(200, 40))
	out := sized.(Model).View()
	if out == "" {
		t.Fatal("wide view rendered nothing")
	}
}

func TestListQuitKey(t *testing.T) {
	l := newTestList()
	_, cmd := l.Update(key("q"))
	if cmd == nil {
		t.Fatal("q produced no command")
	}
	if _, ok := cmd().(quitMsg); !ok {
		t.Fatalf("q returned %T, want quitMsg", cmd())
	}
}

func TestCommentsBackKey(t *testing.T) {
	c := newTestComments(t)
	_, cmd := c.Update(key("q"))
	if cmd == nil {
		t.Fatal("q produced no command")
	}
	if _, ok := cmd().(popMsg); !ok {
		t.Fatalf("q returned %T, want popMsg", cmd())
	}
}

func TestSearchCapturesGlobalKeys(t *testing.T) {
	m := New(Options{Client: lobsters.New()})
	sized, _ := m.Update(sizeMsg(120, 40))
	searching, _ := sized.(Model).Update(key("/"))
	typed, _ := searching.(Model).Update(key("i"))
	if typed.(Model).showHelp {
		t.Error("i should not open help while searching")
	}
	list := typed.(Model).stack[0].(*listScreen)
	if list.query != "i" {
		t.Errorf("query = %q, want i", list.query)
	}
}

func TestHelpToggle(t *testing.T) {
	m := New(Options{Client: lobsters.New()})
	sized, _ := m.Update(sizeMsg(120, 40))

	shown, _ := sized.(Model).Update(key("i"))
	if !shown.(Model).showHelp {
		t.Fatal("i did not open help")
	}
	if out := shown.(Model).View(); !strings.Contains(out, "keys") {
		t.Errorf("help view missing title: %q", out)
	}
	hidden, _ := shown.(Model).Update(key("esc"))
	if hidden.(Model).showHelp {
		t.Error("esc did not close help")
	}
}

func TestWideNavigationDoesNotFetch(t *testing.T) {
	m := New(Options{Client: lobsters.New()})
	got, _ := m.Update(sizeMsg(200, 40))
	m = got.(Model)
	l := m.stack[0].(*listScreen)
	l.stories = []lobsters.Story{{ShortID: "a"}, {ShortID: "b"}}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if cmd != nil {
		if msg := cmd(); msg != nil {
			t.Fatalf("moving the cursor produced %T; expected no work", msg)
		}
	}
	if len(m.stack) != 1 {
		t.Fatal("j must not open comments")
	}
}
