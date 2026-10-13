package domain

import (
	"encoding/json"
	"os"
	"testing"

	"noben.org/lbx/lobsters"
)

func ptr(s string) *string { return &s }

func TestBuildTreeNesting(t *testing.T) {
	comments := []lobsters.Comment{
		{ShortID: "a", Depth: 0},
		{ShortID: "b", Depth: 1, ParentComment: ptr("a")},
		{ShortID: "c", Depth: 2, ParentComment: ptr("b")},
		{ShortID: "d", Depth: 1, ParentComment: ptr("a")},
		{ShortID: "e", Depth: 0},
	}

	tree := BuildTree(comments)
	if tree.Size() != 5 {
		t.Fatalf("Size = %d, want 5", tree.Size())
	}
	if len(tree.Roots) != 2 {
		t.Fatalf("roots = %d, want 2", len(tree.Roots))
	}
	a := tree.Find("a")
	if a == nil || len(a.Children) != 2 {
		t.Fatalf("a children = %v, want 2", a)
	}
	if a.Children[0].Comment.ShortID != "b" || a.Children[1].Comment.ShortID != "d" {
		t.Errorf("a children order = %q,%q want b,d", a.Children[0].Comment.ShortID, a.Children[1].Comment.ShortID)
	}
	b := tree.Find("b")
	if b == nil || !b.HasChildren() || b.Children[0].Comment.ShortID != "c" {
		t.Fatalf("b should have child c")
	}
	if tree.Find("e") == nil || tree.Find("e").HasChildren() {
		t.Errorf("e should be a childless root")
	}
}

func TestBuildTreeMalformed(t *testing.T) {
	comments := []lobsters.Comment{
		{ShortID: "orphan", Depth: 1, ParentComment: ptr("missing")},
		{ShortID: "blank", Depth: 1, ParentComment: ptr("")},
	}
	tree := BuildTree(comments)
	if len(tree.Roots) != 2 {
		t.Fatalf("roots = %d, want 2 (orphans promote to root)", len(tree.Roots))
	}
}

func TestBuildTreeEmpty(t *testing.T) {
	tree := BuildTree(nil)
	if tree.Size() != 0 || len(tree.Roots) != 0 {
		t.Fatalf("empty tree = %d comments, %d roots", tree.Size(), len(tree.Roots))
	}
}

func TestBuildTreeFixture(t *testing.T) {
	data, err := os.ReadFile("../lobsters/testdata/story.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var story lobsters.StoryWithComments
	if err := json.Unmarshal(data, &story); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}

	tree := BuildTree(story.Comments)
	if tree.Size() != len(story.Comments) {
		t.Fatalf("Size = %d, want %d", tree.Size(), len(story.Comments))
	}

	var count int
	var walk func(n *Node)
	walk = func(n *Node) {
		count++
		for _, c := range n.Children {
			if c.Comment.ParentComment == nil || *c.Comment.ParentComment != n.Comment.ShortID {
				t.Errorf("child %q parent = %v, want %q", c.Comment.ShortID, c.Comment.ParentComment, n.Comment.ShortID)
			}
			walk(c)
		}
	}
	for _, r := range tree.Roots {
		if r.Comment.ParentComment != nil {
			t.Errorf("root %q has parent %v", r.Comment.ShortID, *r.Comment.ParentComment)
		}
		walk(r)
	}
	if count != tree.Size() {
		t.Errorf("walked %d nodes, tree size %d", count, tree.Size())
	}
}
