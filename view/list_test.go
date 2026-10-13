package view

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"noben.org/lbx/categories"
	"noben.org/lbx/lobsters"
)

func key(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "ctrl+d":
		return tea.KeyMsg{Type: tea.KeyCtrlD}
	case "ctrl+u":
		return tea.KeyMsg{Type: tea.KeyCtrlU}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func testStories() []lobsters.Story {
	now := time.Now()
	return []lobsters.Story{
		{ShortID: "a", Title: "Alpha", URL: "https://example.com/a", Score: 10, CommentCount: 2, SubmitterUser: "alice", Tags: []string{"go"}, CreatedAt: now.Add(-2 * time.Hour)},
		{ShortID: "b", Title: "Beta", URL: "https://www.example.org/b", Score: 5, CommentCount: 0, SubmitterUser: "bob", Tags: []string{"rust", "plt"}, CreatedAt: now.Add(-30 * time.Minute)},
		{ShortID: "c", Title: "Gamma", URL: "https://example.net/c", Score: 1, CommentCount: 1, SubmitterUser: "carol", CreatedAt: now.Add(-72 * time.Hour)},
	}
}

func newTestList() *listScreen {
	l := newListScreen(nil, categories.Default(), nil, nil, false)
	l.key = l.current().Key()
	l.stories = testStories()
	l.height = 20
	l.width = 120
	return l
}

func TestListNavigation(t *testing.T) {
	l := newTestList()

	l.Update(key("j"))
	if l.cursor != 1 {
		t.Fatalf("after j cursor = %d, want 1", l.cursor)
	}
	l.Update(key("k"))
	if l.cursor != 0 {
		t.Fatalf("after k cursor = %d, want 0", l.cursor)
	}
	l.Update(key("G"))
	if l.cursor != 2 {
		t.Fatalf("after G cursor = %d, want 2", l.cursor)
	}
	l.Update(key("g"))
	if l.cursor != 0 {
		t.Fatalf("after g cursor = %d, want 0", l.cursor)
	}
	l.Update(key("k"))
	if l.cursor != 0 {
		t.Fatalf("k at top cursor = %d, want 0", l.cursor)
	}
}

func TestListCategoryRotation(t *testing.T) {
	l := newTestList()
	if l.current().Label() != "hottest" {
		t.Fatalf("initial category = %q", l.current().Label())
	}
	l.Update(key("tab"))
	if l.current().Label() != "newest" {
		t.Fatalf("after tab category = %q, want newest", l.current().Label())
	}
	if l.key != "newest" {
		t.Fatalf("key = %q, want newest", l.key)
	}
}

func TestListRefreshPreservesSelection(t *testing.T) {
	l := newTestList()
	l.Update(key("j"))
	if l.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", l.cursor)
	}
	l.Update(key("r"))
	if l.pendingSelect != "b" {
		t.Fatalf("pendingSelect = %q, want b", l.pendingSelect)
	}

	reloaded := []lobsters.Story{
		{ShortID: "x", Title: "New"},
		{ShortID: "b", Title: "Beta"},
	}
	l.Update(feedLoadedMsg{key: l.key, stories: reloaded})
	if l.cursor != 1 {
		t.Fatalf("after refresh cursor = %d, want 1 (story b)", l.cursor)
	}
	if s, _ := l.selected(); s.ShortID != "b" {
		t.Fatalf("selected = %q, want b", s.ShortID)
	}
}

func TestListRefreshMissingSelection(t *testing.T) {
	l := newTestList()
	l.Update(key("G"))
	l.Update(key("r"))
	l.Update(feedLoadedMsg{key: l.key, stories: []lobsters.Story{{ShortID: "x"}, {ShortID: "y"}}})
	if l.cursor != 1 {
		t.Fatalf("cursor = %d, want clamped 1", l.cursor)
	}
}

func TestListError(t *testing.T) {
	l := newTestList()
	l.Update(feedLoadedMsg{key: l.key, err: errTest})
	if l.err == nil {
		t.Fatal("err = nil, want error")
	}
	if !strings.Contains(l.View(80, 20), "error") {
		t.Errorf("error view = %q", l.View(80, 20))
	}
}

func TestRenderRow(t *testing.T) {
	l := newTestList()
	row := l.renderRow(1, 120)
	for _, want := range []string{"Beta", "example.org", "rust", "plt", "5 pts", "0 comments", "bob"} {
		if !strings.Contains(row, want) {
			t.Errorf("row missing %q:\n%s", want, row)
		}
	}
	if strings.Contains(row, "www.") {
		t.Errorf("row should strip www.: %s", row)
	}
}

func TestRenderRowTruncates(t *testing.T) {
	l := newTestList()
	row := l.renderRow(0, 10)
	for _, line := range strings.Split(row, "\n") {
		if w := lipgloss.Width(line); w > 10 {
			t.Errorf("line width %d exceeds 10: %q", w, line)
		}
	}
}

func TestListSearchFlow(t *testing.T) {
	l := newTestList()
	l.Update(key("/"))
	if !l.searching {
		t.Fatal("/ did not start search")
	}
	l.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("go")})
	if l.query != "go" {
		t.Fatalf("query = %q, want go", l.query)
	}
	_, cmd := l.handleSearchKey(tea.KeyMsg{Type: tea.KeyEnter})
	if l.searching {
		t.Fatal("enter did not end search")
	}
	if cmd == nil {
		t.Fatal("enter produced no browser command")
	}

	l.Update(key("/"))
	l.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if l.searching {
		t.Fatal("esc did not cancel search")
	}
}

type testErr struct{}

func (testErr) Error() string { return "boom" }

var errTest = testErr{}
