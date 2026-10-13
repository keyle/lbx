package domain

import "noben.org/lbx/lobsters"

// Node is a comment together with its replies, in thread order.
type Node struct {
	Comment  lobsters.Comment
	Children []*Node
}

// HasChildren reports whether the comment has any replies.
func (n *Node) HasChildren() bool { return len(n.Children) > 0 }

// Tree is an ordered comment tree built from lobste.rs's flat comment array.
type Tree struct {
	Roots []*Node
	index map[string]*Node
}

// BuildTree links the flat comments into a tree using parent_comment. Comments
// whose parent is missing or empty are treated as roots, so malformed threads
// degrade to a flat list rather than being dropped. Input order is preserved.
func BuildTree(comments []lobsters.Comment) *Tree {
	t := &Tree{index: make(map[string]*Node, len(comments))}
	for i := range comments {
		t.index[comments[i].ShortID] = &Node{Comment: comments[i]}
	}
	for i := range comments {
		c := comments[i]
		n := t.index[c.ShortID]
		if c.ParentComment == nil || *c.ParentComment == "" {
			t.Roots = append(t.Roots, n)
			continue
		}
		parent, ok := t.index[*c.ParentComment]
		if !ok {
			t.Roots = append(t.Roots, n)
			continue
		}
		parent.Children = append(parent.Children, n)
	}
	return t
}

// Find returns the node with the given comment short id, or nil.
func (t *Tree) Find(shortID string) *Node {
	return t.index[shortID]
}

// Size returns the number of comments in the tree.
func (t *Tree) Size() int { return len(t.index) }
