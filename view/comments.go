package view

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"noben.org/lbx/domain"
	"noben.org/lbx/help"
	"noben.org/lbx/keymap"
	"noben.org/lbx/lobsters"
	"noben.org/lbx/render"
	"noben.org/lbx/style"
	"noben.org/lbx/timeago"
)

// commentsKeymap is the comment section's keybindings.
var commentsKeymap = keymap.Map{
	{Name: "down", Keys: []string{"j", "down"}, Help: "j / k", Desc: "scroll down / up"},
	{Name: "up", Keys: []string{"k", "up"}},
	{Name: "halfDown", Keys: []string{"ctrl+d"}, Help: "Ctrl-D/Ctrl-U", Desc: "half-page down / up"},
	{Name: "halfUp", Keys: []string{"ctrl+u"}},
	{Name: "top", Keys: []string{"g", "home"}, Help: "g / G", Desc: "top / bottom"},
	{Name: "bottom", Keys: []string{"G", "end"}},
	{Name: "nextRoot", Keys: []string{"n"}, Help: "n / N", Desc: "next / previous top-level comment"},
	{Name: "prevRoot", Keys: []string{"N"}},
	{Name: "collapse", Keys: []string{"h"}, Help: "h / l", Desc: "collapse / expand thread"},
	{Name: "expand", Keys: []string{"l"}},
	{Name: "toggleAll", Keys: []string{"enter"}, Help: "Enter", Desc: "toggle all threads"},
	{Name: "navigate", Keys: []string{"a"}, Help: "a", Desc: "navigate comments"},
	{Name: "links", Keys: []string{"tab"}, Help: "Tab", Desc: "select link"},
	{Name: "esc", Keys: []string{"esc"}, Help: "Esc", Desc: "leave navigate mode / back"},
	{Name: "back", Keys: []string{"q"}, Help: "q", Desc: "back"},
}

const maxIndent = 10

type commentMode int

const (
	modeRead commentMode = iota
	modeNavigate
)

type commentsScreen struct {
	client *lobsters.Client
	story  lobsters.Story
	tree   *domain.Tree
	since  time.Time

	loading bool
	spin    spinner.Model
	err     error

	collapsed map[string]bool
	mode      commentMode

	scroll int
	sel    int

	width, height int

	visible   []*domain.Node
	lines     []string
	lineID    []string
	lineRoot  []bool
	nodeStart map[string]int

	picker linkPicker
}

func newCommentsScreen(client *lobsters.Client, story lobsters.Story, since time.Time) *commentsScreen {
	return &commentsScreen{
		client:    client,
		story:     story,
		since:     since,
		spin:      newSpinner(),
		collapsed: map[string]bool{},
	}
}

func (c *commentsScreen) Init() tea.Cmd {
	c.loading = true
	client := c.client
	id := c.story.ShortID
	load := func() tea.Msg {
		story, err := client.Story(context.Background(), id)
		return storyLoadedMsg{id: id, story: story, err: err}
	}
	return tea.Batch(load, c.spin.Tick)
}

func (c *commentsScreen) Title() string { return c.story.Title }

func (c *commentsScreen) Help() []help.Binding { return commentsKeymap.Help() }

func (c *commentsScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		c.spin, cmd = c.spin.Update(msg)
		return c, cmd
	case storyLoadedMsg:
		if msg.id != c.story.ShortID {
			return c, nil
		}
		c.loading = false
		c.err = msg.err
		if msg.err == nil && msg.story != nil {
			c.tree = domain.BuildTree(msg.story.Comments)
			// Start with replies hidden; only top-level comments show.
			c.collapsed = map[string]bool{}
			for _, root := range c.tree.Roots {
				if root.HasChildren() {
					c.collapsed[root.Comment.ShortID] = true
				}
			}
			c.rebuild()
		}
		return c, nil
	case tea.KeyMsg:
		return c.handleKey(msg)
	}
	return c, nil
}

func (c *commentsScreen) handleKey(msg tea.KeyMsg) (Screen, tea.Cmd) {
	if cmd, handled := c.picker.handleKey(msg); handled {
		return c, cmd
	}
	switch commentsKeymap.Action(msg) {
	case "back":
		return c, func() tea.Msg { return popMsg{} }
	case "esc":
		if c.mode == modeNavigate {
			c.mode = modeRead
		} else {
			return c, func() tea.Msg { return popMsg{} }
		}
	case "navigate":
		c.mode = modeNavigate
		c.sel = c.currentVisibleIndex()
	case "down":
		if c.mode == modeNavigate {
			c.selectMove(1)
		} else {
			c.scrollBy(1)
		}
	case "up":
		if c.mode == modeNavigate {
			c.selectMove(-1)
		} else {
			c.scrollBy(-1)
		}
	case "halfDown":
		c.scrollBy(c.pageSize() / 2)
	case "halfUp":
		c.scrollBy(-c.pageSize() / 2)
	case "top":
		c.scroll = 0
	case "bottom":
		c.scroll = c.maxScroll()
	case "collapse":
		c.setCollapsed(c.currentNode(), true)
	case "expand":
		c.setCollapsed(c.currentNode(), false)
	case "toggleAll":
		c.toggleAll()
	case "nextRoot":
		c.jumpRoot(1)
	case "prevRoot":
		c.jumpRoot(-1)
	case "links":
		c.picker.toggle(render.Links(c.rawHTML()))
	}
	return c, nil
}

func (c *commentsScreen) rawHTML() string {
	var b strings.Builder
	if c.tree == nil {
		return ""
	}
	var walk func(n *domain.Node)
	walk = func(n *domain.Node) {
		b.WriteString(n.Comment.Comment)
		b.WriteByte('\n')
		for _, child := range n.Children {
			walk(child)
		}
	}
	for _, root := range c.tree.Roots {
		walk(root)
	}
	return b.String()
}

func (c *commentsScreen) rebuild() {
	c.visible = c.visibleNodes()
	c.lines = c.lines[:0]
	c.lineID = c.lineID[:0]
	c.lineRoot = c.lineRoot[:0]
	c.nodeStart = make(map[string]int, len(c.visible))

	for _, n := range c.visible {
		id := n.Comment.ShortID
		indent := strings.Repeat("  ", min(n.Comment.Depth, maxIndent))
		c.nodeStart[id] = len(c.lines)

		header := indent + style.Submitter.Render("@"+n.Comment.CommentingUser) +
			" " + style.Meta.Render(timeago.Since(n.Comment.CreatedAt)) +
			" " + style.Score.Render(fmt.Sprintf("%d", n.Comment.Score))
		if !c.since.IsZero() && n.Comment.CreatedAt.After(c.since) {
			header += " " + style.Comments.Render("new")
		}
		if n.HasChildren() {
			marker := "[-]"
			if c.collapsed[id] {
				marker = "[+]"
			}
			header += " " + style.Meta.Render(marker)
		}
		c.appendLine(header, id, n.Comment.ParentComment == nil)

		body := c.commentBody(n, indent)
		for _, line := range body {
			c.appendLine(indent+line, id, false)
		}
		c.appendLine("", id, false)
	}

	c.scroll = clamp(c.scroll, 0, c.maxScroll())
	if c.sel >= len(c.visible) {
		c.sel = max(0, len(c.visible)-1)
	}
}

func (c *commentsScreen) commentBody(n *domain.Node, indent string) []string {
	if n.Comment.IsDeleted {
		return []string{style.Meta.Render("[deleted]")}
	}
	bodyWidth := c.width - len(indent)
	if bodyWidth < 20 {
		bodyWidth = 20
	}
	body := render.HTML(n.Comment.Comment, bodyWidth)
	if strings.TrimSpace(body) == "" {
		return nil
	}
	return strings.Split(body, "\n")
}

func (c *commentsScreen) appendLine(text, id string, root bool) {
	c.lines = append(c.lines, text)
	c.lineID = append(c.lineID, id)
	c.lineRoot = append(c.lineRoot, root)
}

func (c *commentsScreen) visibleNodes() []*domain.Node {
	if c.tree == nil {
		return nil
	}
	var out []*domain.Node
	var walk func(n *domain.Node)
	walk = func(n *domain.Node) {
		out = append(out, n)
		if c.collapsed[n.Comment.ShortID] {
			return
		}
		for _, child := range n.Children {
			walk(child)
		}
	}
	for _, root := range c.tree.Roots {
		walk(root)
	}
	return out
}

func (c *commentsScreen) currentNode() *domain.Node {
	if len(c.visible) == 0 {
		return nil
	}
	if c.mode == modeNavigate {
		return c.visible[clamp(c.sel, 0, len(c.visible)-1)]
	}
	idx := c.currentVisibleIndex()
	return c.visible[idx]
}

func (c *commentsScreen) currentVisibleIndex() int {
	if len(c.visible) == 0 {
		return 0
	}
	best := 0
	for i, n := range c.visible {
		if c.nodeStart[n.Comment.ShortID] <= c.scroll {
			best = i
		}
	}
	return best
}

func (c *commentsScreen) setCollapsed(n *domain.Node, collapsed bool) {
	if n == nil || !n.HasChildren() {
		return
	}
	id := n.Comment.ShortID
	if collapsed {
		c.collapsed[id] = true
	} else {
		delete(c.collapsed, id)
	}
	c.rebuild()
	c.ensureSelectionVisible()
}

func (c *commentsScreen) toggleAll() {
	anyExpanded := false
	for _, n := range c.visible {
		if n.HasChildren() && !c.collapsed[n.Comment.ShortID] {
			anyExpanded = true
			break
		}
	}
	if c.tree == nil {
		return
	}
	c.collapsed = map[string]bool{}
	if anyExpanded {
		var walk func(n *domain.Node)
		walk = func(n *domain.Node) {
			if n.HasChildren() {
				c.collapsed[n.Comment.ShortID] = true
			}
			for _, child := range n.Children {
				walk(child)
			}
		}
		for _, root := range c.tree.Roots {
			walk(root)
		}
	}
	c.rebuild()
}

func (c *commentsScreen) selectMove(delta int) {
	if len(c.visible) == 0 {
		return
	}
	c.sel = clamp(c.sel+delta, 0, len(c.visible)-1)
	c.ensureSelectionVisible()
}

func (c *commentsScreen) ensureSelectionVisible() {
	if c.mode != modeNavigate || len(c.visible) == 0 {
		return
	}
	n := c.visible[c.sel]
	start := c.nodeStart[n.Comment.ShortID]
	if start < c.scroll {
		c.scroll = start
	}
	if start >= c.scroll+c.pageSize() {
		c.scroll = start - c.pageSize() + 1
	}
	c.scroll = clamp(c.scroll, 0, c.maxScroll())
}

func (c *commentsScreen) jumpRoot(delta int) {
	if len(c.visible) == 0 {
		return
	}
	cur := c.currentVisibleIndex()
	i := cur
	if delta > 0 {
		i++
	} else {
		i--
	}
	for ; i >= 0 && i < len(c.visible); i += delta {
		if c.visible[i].Comment.ParentComment == nil {
			c.scroll = c.nodeStart[c.visible[i].Comment.ShortID]
			if c.mode == modeNavigate {
				c.sel = i
			}
			return
		}
	}
	if delta < 0 {
		c.scroll = 0
	}
}

func (c *commentsScreen) scrollBy(delta int) {
	c.scroll = clamp(c.scroll+delta, 0, c.maxScroll())
}

func (c *commentsScreen) pageSize() int {
	if c.height < 1 {
		return 1
	}
	return c.height
}

func (c *commentsScreen) maxScroll() int {
	if m := len(c.lines) - c.pageSize(); m > 0 {
		return m
	}
	return 0
}

func (c *commentsScreen) View(width, height int) string {
	c.width, c.height = width, height
	if c.err != nil {
		return style.Error.Render("error: " + c.err.Error())
	}
	if c.loading {
		return loadingView(c.spin, "loading comments…")
	}
	if c.tree == nil {
		return style.Status.Render("no comments")
	}
	if len(c.lines) == 0 {
		return style.Status.Render("no comments")
	}
	if c.maxScroll() < c.scroll {
		c.scroll = c.maxScroll()
	}

	end := c.scroll + height
	if end > len(c.lines) {
		end = len(c.lines)
	}
	var b strings.Builder
	for i := c.scroll; i < end; i++ {
		if i > c.scroll {
			b.WriteString("\n")
		}
		line := ansi.Truncate(c.lines[i], width, "…")
		if c.mode == modeNavigate && c.lineID[i] == c.visible[c.sel].Comment.ShortID && c.lineID[i] != "" {
			line = style.Selection.Reverse(true).Render(line)
		}
		b.WriteString(line)
	}
	body := b.String()
	if c.picker.active {
		body = overlayTop(body, c.picker.lines(width, height/2), width)
	}
	return body
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
