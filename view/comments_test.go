package view

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"noben.org/lbx/domain"
	"noben.org/lbx/lobsters"
)

func newTestComments(t *testing.T) *commentsScreen {
	t.Helper()
	data, err := os.ReadFile("../lobsters/testdata/story.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var story lobsters.StoryWithComments
	if err := json.Unmarshal(data, &story); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	c := newCommentsScreen(nil, story.Story, time.Time{})
	c.tree = domain.BuildTree(story.Comments)
	c.width, c.height = 80, 24
	c.rebuild()
	return c
}

func firstRootWithChildren(c *commentsScreen) *domain.Node {
	for _, r := range c.visible {
		if r.HasChildren() {
			return r
		}
	}
	return nil
}

func TestCommentsRebuild(t *testing.T) {
	c := newTestComments(t)
	if len(c.visible) == 0 {
		t.Fatal("no visible comments")
	}
	if len(c.lines) == 0 {
		t.Fatal("no rendered lines")
	}
	if !strings.Contains(c.lines[0], "pts") {
		t.Errorf("first line missing score: %q", c.lines[0])
	}
}

func TestCommentsCollapse(t *testing.T) {
	c := newTestComments(t)
	root := firstRootWithChildren(c)
	if root == nil {
		t.Fatal("fixture has no root with children")
	}
	descendants := countDescendants(root)
	before := len(c.visible)

	c.setCollapsed(root, true)
	if got := len(c.visible); got != before-descendants {
		t.Errorf("after collapse visible = %d, want %d", got, before-descendants)
	}

	c.setCollapsed(root, false)
	if got := len(c.visible); got != before {
		t.Errorf("after expand visible = %d, want %d", got, before)
	}
}

func TestCommentsExpandCollapseAll(t *testing.T) {
	c := newTestComments(t)
	c.Update(key("c"))
	for _, n := range c.visible {
		if n.HasChildren() && !c.collapsed[n.Comment.ShortID] {
			t.Errorf("c did not collapse %q", n.Comment.ShortID)
		}
	}
	c.Update(key("a"))
	if len(c.collapsed) != 0 {
		t.Errorf("a left %d collapsed", len(c.collapsed))
	}
}

func TestCommentsExpandThread(t *testing.T) {
	c := newTestComments(t)
	c.Update(key("c"))
	root := firstRootWithChildren(c)
	c.selectIndex(indexOf(c, root))
	c.Update(key("l"))
	if c.collapsed[root.Comment.ShortID] {
		t.Fatal("l did not expand")
	}
	c.Update(key("L"))
	for _, ch := range root.Children {
		if c.collapsed[ch.Comment.ShortID] {
			t.Errorf("L left %q collapsed", ch.Comment.ShortID)
		}
	}
}

func TestCommentsSelectMove(t *testing.T) {
	c := newTestComments(t)
	c.Update(key("j"))
	if c.sel != 1 || c.selID != c.visible[1].Comment.ShortID {
		t.Fatalf("sel = %d, id %q", c.sel, c.selID)
	}
	c.Update(key("k"))
	if c.sel != 0 {
		t.Fatalf("sel = %d, want 0", c.sel)
	}
	root := firstRootWithChildren(c)
	c.selectIndex(indexOf(c, root))
	c.Update(key("h"))
	if !c.collapsed[root.Comment.ShortID] {
		t.Errorf("h did not collapse %q", root.Comment.ShortID)
	}
}

func TestCommentsView(t *testing.T) {
	c := newTestComments(t)
	c.Update(key("G"))
	out := c.View(80, 24)
	if out == "" {
		t.Fatal("empty view")
	}
	if !strings.Contains(out, "pts") {
		t.Errorf("view missing authors:\n%s", out)
	}
}

func TestCommentsError(t *testing.T) {
	c := newCommentsScreen(nil, lobsters.Story{}, time.Time{})
	c.Update(storyLoadedMsg{err: errTest})
	if !strings.Contains(c.View(80, 24), "error") {
		t.Errorf("error view = %q", c.View(80, 24))
	}
}

func countDescendants(n *domain.Node) int {
	count := 0
	for _, c := range n.Children {
		count += 1 + countDescendants(c)
	}
	return count
}

func indexOf(c *commentsScreen, n *domain.Node) int {
	for i, v := range c.visible {
		if v == n {
			return i
		}
	}
	return 0
}

func TestCommentsLoadCollapsed(t *testing.T) {
	data, err := os.ReadFile("../lobsters/testdata/story.json")
	if err != nil {
		t.Fatal(err)
	}
	var story lobsters.StoryWithComments
	if err := json.Unmarshal(data, &story); err != nil {
		t.Fatal(err)
	}
	c := newCommentsScreen(nil, story.Story, time.Time{})
	c.width, c.height = 80, 24
	c.Update(storyLoadedMsg{id: story.ShortID, story: &story})
	if len(c.visible) != len(c.tree.Roots) {
		t.Errorf("visible = %d, want only %d top-level comments", len(c.visible), len(c.tree.Roots))
	}
	if root := firstRootWithChildren(c); root != nil {
		c.setCollapsed(root, false)
		if len(c.visible) <= len(c.tree.Roots) {
			t.Error("expanding a root should reveal replies")
		}
	}
}
