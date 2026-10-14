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
	{Name: "down", Keys: []string{"j", "down"}, Help: "j / k", Desc: "next / previous comment"},
	{Name: "up", Keys: []string{"k", "up"}},
	{Name: "halfDown", Keys: []string{"ctrl+d"}, Help: "Ctrl-D/Ctrl-U", Desc: "half-page down / up"},
	{Name: "halfUp", Keys: []string{"ctrl+u"}},
	{Name: "top", Keys: []string{"g", "home"}, Help: "g / G", Desc: "top / bottom"},
	{Name: "bottom", Keys: []string{"G", "end"}},
	{Name: "expand", Keys: []string{"l"}, Help: "l / h", Desc: "expand / collapse comment"},
	{Name: "collapse", Keys: []string{"h"}},
	{Name: "expandThread", Keys: []string{"L"}, Help: "L", Desc: "expand whole sub-thread"},
	{Name: "expandAll", Keys: []string{"a"}, Help: "a / c", Desc: "expand / collapse all"},
	{Name: "collapseAll", Keys: []string{"c"}},
	{Name: "links", Keys: []string{"tab"}, Help: "Tab", Desc: "select link"},
	{Name: "help", Help: "? / i", Desc: "this help"},
	{Name: "back", Keys: []string{"q", "esc"}, Help: "q / Esc", Desc: "back"},
}

const maxIndent = 10

type commentsScreen struct {
	client *lobsters.Client
	story  lobsters.Story
	tree   *domain.Tree
	since  time.Time

	loading bool
	spin    spinner.Model
	err     error

	collapsed map[string]bool

	scroll int
	sel    int    // index into visible
	selID  string // short id of the highlighted comment

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
	case "down":
		c.selectMove(1)
	case "up":
		c.selectMove(-1)
	case "halfDown":
		c.scrollBy(c.pageSize() / 2)
		c.syncSelToScroll(1)
	case "halfUp":
		c.scrollBy(-c.pageSize() / 2)
		c.syncSelToScroll(-1)
	case "top":
		c.selectIndex(0)
		c.scroll = 0
	case "bottom":
		c.selectIndex(len(c.visible) - 1)
		c.scroll = c.maxScroll()
	case "collapse":
		c.setCollapsed(c.currentNode(), true)
	case "expand":
		c.setCollapsed(c.currentNode(), false)
	case "expandThread":
		c.expandThread(c.currentNode())
	case "expandAll":
		c.expandAll()
	case "collapseAll":
		c.collapseAll()
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

		header := indent + style.Submitter.Render(n.Comment.CommentingUser)
		if n.HasChildren() {
			marker := "[-]"
			if c.collapsed[id] {
				marker = "[+]"
			}
			header += " " + style.Submitter.Render(marker)
		}
		header += " " + style.Score.Render(fmt.Sprintf("%d pts", n.Comment.Score)) +
			" " + style.Meta.Render(timeago.Since(n.Comment.CreatedAt))
		if !c.since.IsZero() && n.Comment.CreatedAt.After(c.since) {
			header += " " + style.Comments.Render("new")
		}
		c.appendLine(header, id, n.Comment.ParentComment == nil)

		body := c.commentBody(n, indent)
		for _, line := range body {
			c.appendLine(indent+line, id, false)
		}
		c.appendLine("", id, false)
	}

	c.resolveSel()
	c.scroll = clamp(c.scroll, 0, c.maxScroll())
}

// resolveSel keeps the highlight on selID, or on its nearest visible ancestor
// when it was hidden by a collapse, defaulting to the first comment.
func (c *commentsScreen) resolveSel() {
	c.sel = 0
	if len(c.visible) == 0 {
		c.selID = ""
		return
	}
	id := c.selID
	for id != "" {
		if _, ok := c.nodeStart[id]; ok {
			break
		}
		n := c.tree.Find(id)
		if n == nil || n.Comment.ParentComment == nil {
			id = ""
			break
		}
		id = *n.Comment.ParentComment
	}
	if id == "" {
		id = c.visible[0].Comment.ShortID
	}
	c.selID = id
	for i, n := range c.visible {
		if n.Comment.ShortID == id {
			c.sel = i
			break
		}
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
	return c.visible[clamp(c.sel, 0, len(c.visible)-1)]
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

func (c *commentsScreen) expandThread(n *domain.Node) {
	if n == nil {
		return
	}
	var walk func(n *domain.Node)
	walk = func(n *domain.Node) {
		delete(c.collapsed, n.Comment.ShortID)
		for _, child := range n.Children {
			walk(child)
		}
	}
	walk(n)
	c.rebuild()
	c.ensureSelectionVisible()
}

func (c *commentsScreen) expandAll() {
	c.collapsed = map[string]bool{}
	c.rebuild()
	c.ensureSelectionVisible()
}

func (c *commentsScreen) collapseAll() {
	if c.tree == nil {
		return
	}
	c.collapsed = map[string]bool{}
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
	c.rebuild()
	c.ensureSelectionVisible()
}

func (c *commentsScreen) selectIndex(i int) {
	if len(c.visible) == 0 {
		return
	}
	c.sel = clamp(i, 0, len(c.visible)-1)
	c.selID = c.visible[c.sel].Comment.ShortID
}

func (c *commentsScreen) selectMove(delta int) {
	c.selectIndex(c.sel + delta)
	c.ensureSelectionVisible()
}

// syncSelToScroll moves the highlight onto a comment that is on screen after a
// page scroll, if it scrolled out of view.
func (c *commentsScreen) syncSelToScroll(dir int) {
	if len(c.visible) == 0 {
		return
	}
	start := c.nodeStart[c.selID]
	if start >= c.scroll && start < c.scroll+c.pageSize() {
		return
	}
	if dir > 0 {
		for i, n := range c.visible {
			if c.nodeStart[n.Comment.ShortID] >= c.scroll {
				c.selectIndex(i)
				return
			}
		}
		c.selectIndex(len(c.visible) - 1)
		return
	}
	for i := len(c.visible) - 1; i >= 0; i-- {
		if c.nodeStart[c.visible[i].Comment.ShortID] < c.scroll+c.pageSize() {
			c.selectIndex(i)
			return
		}
	}
}

func (c *commentsScreen) ensureSelectionVisible() {
	if len(c.visible) == 0 {
		return
	}
	start := c.nodeStart[c.selID]
	if start < c.scroll {
		c.scroll = start
	}
	if start >= c.scroll+c.pageSize() {
		c.scroll = start - c.pageSize() + 1
	}
	c.scroll = clamp(c.scroll, 0, c.maxScroll())
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
		if c.selID != "" && i == c.nodeStart[c.selID] {
			line = style.Selection.Reverse(true).Render(ansi.Strip(line))
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
