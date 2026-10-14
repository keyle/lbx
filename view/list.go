package view

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/pkg/browser"

	"noben.org/lbx/categories"
	"noben.org/lbx/favorites"
	"noben.org/lbx/help"
	"noben.org/lbx/history"
	"noben.org/lbx/keymap"
	"noben.org/lbx/lobsters"
	"noben.org/lbx/style"
	"noben.org/lbx/timeago"
)

// listKeymap is the list view's keybindings.
var listKeymap = keymap.Map{
	{Name: "next", Keys: []string{"j", "down"}, Help: "j / k", Desc: "next / previous story"},
	{Name: "prev", Keys: []string{"k", "up"}},
	{Name: "halfDown", Keys: []string{"ctrl+d"}, Help: "Ctrl-D/Ctrl-U", Desc: "half-page down / up"},
	{Name: "halfUp", Keys: []string{"ctrl+u"}},
	{Name: "top", Keys: []string{"g", "home"}, Help: "g / G", Desc: "top / bottom"},
	{Name: "bottom", Keys: []string{"G", "end"}},
	{Name: "comments", Keys: []string{"enter", "l"}, Help: "Enter / l", Desc: "open comments"},
	{Name: "reader", Keys: []string{" "}, Help: "Space", Desc: "open reader mode"},
	{Name: "openStory", Keys: []string{"o"}, Help: "o / c", Desc: "open story / comments in browser"},
	{Name: "openComments", Keys: []string{"c"}},
	{Name: "tab", Keys: []string{"tab"}, Help: "Tab", Desc: "next category"},
	{Name: "refresh", Keys: []string{"r"}, Help: "r", Desc: "refresh"},
	{Name: "favorite", Keys: []string{"f"}, Help: "f", Desc: "toggle favorite"},
	{Name: "unfavorite", Keys: []string{"x"}, Help: "x", Desc: "remove favorite"},
	{Name: "toggleRead", Keys: []string{"u"}, Help: "u", Desc: "toggle read"},
	{Name: "search", Keys: []string{"/"}, Help: "/", Desc: "search"},
	{Name: "help", Keys: []string{"i"}, Help: "? / i", Desc: "help"},
	{Name: "quit", Keys: []string{"q"}, Help: "q", Desc: "quit"},
}

// rowLines is a story row (title, meta) plus one blank separator line.
const rowLines = 3

type listScreen struct {
	client *lobsters.Client
	cats   []categories.Category
	active int

	fav    *favorites.Store
	hist   *history.Store
	noHist bool

	searching bool
	query     string

	stories []lobsters.Story
	cursor  int
	offset  int
	loading bool
	spin    spinner.Model
	err     error
	key     string

	pendingSelect string

	width, height int
}

func newListScreen(client *lobsters.Client, cats []categories.Category, fav *favorites.Store, hist *history.Store, noHist bool) *listScreen {
	return &listScreen{client: client, cats: cats, fav: fav, hist: hist, noHist: noHist, spin: newSpinner()}
}

func (l *listScreen) current() categories.Category { return l.cats[l.active] }

func (l *listScreen) Init() tea.Cmd {
	l.loading = true
	l.key = l.current().Key()
	return tea.Batch(loadFeedCmd(l.client, l.current(), l.fav), l.spin.Tick)
}

func (l *listScreen) Title() string {
	if l.err != nil {
		return "error"
	}
	return ""
}

func (l *listScreen) Help() []help.Binding { return listKeymap.Help() }

// CapturingInput reports whether the search prompt is active.
func (l *listScreen) CapturingInput() bool { return l.searching }

func (l *listScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case feedLoadedMsg:
		if msg.key != l.key {
			return l, nil
		}
		l.loading = false
		l.err = msg.err
		if msg.err == nil {
			l.stories = msg.stories
			l.restoreSelection()
			l.clampCursor()
		}
		return l, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		l.spin, cmd = l.spin.Update(msg)
		return l, cmd
	case tea.KeyMsg:
		return l.handleKey(msg)
	}
	return l, nil
}

func (l *listScreen) handleKey(msg tea.KeyMsg) (Screen, tea.Cmd) {
	if l.searching {
		return l.handleSearchKey(msg)
	}
	switch listKeymap.Action(msg) {
	case "next":
		l.move(1)
	case "prev":
		l.move(-1)
	case "top":
		l.cursor = 0
	case "bottom":
		if len(l.stories) > 0 {
			l.cursor = len(l.stories) - 1
		}
	case "halfDown":
		l.move(l.pageSize() / 2)
	case "halfUp":
		l.move(-l.pageSize() / 2)
	case "tab":
		l.active = (l.active + 1) % len(l.cats)
		l.cursor, l.offset = 0, 0
		l.pendingSelect = ""
		l.key = l.current().Key()
		l.loading = true
		return l, tea.Batch(loadFeedCmd(l.client, l.current(), l.fav), func() tea.Msg {
			return categoryChangedMsg{index: l.active}
		})
	case "refresh":
		if s, ok := l.selected(); ok {
			l.pendingSelect = s.ShortID
		}
		l.loading = true
		return l, loadFeedCmd(l.client, l.current(), l.fav)
	case "comments":
		if s, ok := l.selected(); ok {
			return l, func() tea.Msg { return openCommentsMsg{story: s} }
		}
	case "reader":
		if s, ok := l.selected(); ok {
			l.markRead(s)
			return l, func() tea.Msg { return openReaderMsg{story: s} }
		}
	case "openStory":
		if s, ok := l.selected(); ok {
			l.markRead(s)
			return l, openURL(s.URL)
		}
	case "openComments":
		if s, ok := l.selected(); ok {
			l.markRead(s)
			return l, openURL(s.CommentsURL)
		}
	case "favorite":
		if s, ok := l.selected(); ok && l.fav != nil {
			if l.current().Kind == categories.KindFavorites {
				_ = l.fav.Remove(s.ShortID)
				l.loading = true
				return l, loadFeedCmd(l.client, l.current(), l.fav)
			}
			_ = l.fav.Add(favorites.FromStory(s))
		}
	case "unfavorite":
		if s, ok := l.selected(); ok && l.fav != nil {
			_ = l.fav.Remove(s.ShortID)
			if l.current().Kind == categories.KindFavorites {
				l.loading = true
				return l, loadFeedCmd(l.client, l.current(), l.fav)
			}
		}
	case "toggleRead":
		if s, ok := l.selected(); ok && l.hist != nil && !l.noHist {
			_, _ = l.hist.Toggle(s.ShortID, s.CommentCount, time.Now())
		}
	case "search":
		l.searching = true
		l.query = ""
	case "quit":
		return l, func() tea.Msg { return quitMsg{} }
	}
	return l, nil
}

func (l *listScreen) handleSearchKey(msg tea.KeyMsg) (Screen, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		q := strings.TrimSpace(l.query)
		l.searching = false
		if q == "" {
			return l, nil
		}
		return l, openURL(lobsters.BaseURL + "/search?q=" + url.QueryEscape(q) + "&what=stories")
	case tea.KeyEsc:
		l.searching = false
		l.query = ""
	case tea.KeyBackspace:
		if r := []rune(l.query); len(r) > 0 {
			l.query = string(r[:len(r)-1])
		}
	case tea.KeyRunes, tea.KeySpace:
		l.query += string(msg.Runes)
	}
	return l, nil
}

func (l *listScreen) selected() (lobsters.Story, bool) {
	if l.cursor < 0 || l.cursor >= len(l.stories) {
		return lobsters.Story{}, false
	}
	return l.stories[l.cursor], true
}

func (l *listScreen) restoreSelection() {
	id := l.pendingSelect
	l.pendingSelect = ""
	if id == "" {
		return
	}
	for i, s := range l.stories {
		if s.ShortID == id {
			l.cursor = i
			return
		}
	}
}

func (l *listScreen) move(delta int) {
	if len(l.stories) == 0 {
		return
	}
	l.cursor += delta
	l.clampCursor()
}

func (l *listScreen) clampCursor() {
	if l.cursor < 0 {
		l.cursor = 0
	}
	if l.cursor >= len(l.stories) {
		l.cursor = len(l.stories) - 1
	}
	if l.cursor < 0 {
		l.cursor = 0
	}
}

func (l *listScreen) pageSize() int {
	visible := (l.height + 1) / rowLines
	if visible < 1 {
		visible = 1
	}
	return visible
}

func (l *listScreen) View(width, height int) string {
	l.width, l.height = width, height
	if l.err != nil {
		return style.Error.Render("error: " + l.err.Error())
	}
	if len(l.stories) == 0 {
		if l.loading {
			return loadingView(l.spin, "loading…")
		}
		return style.Status.Render("no stories")
	}

	visible := l.pageSize()
	if l.searching {
		visible = l.height / rowLines
		if visible < 1 {
			visible = 1
		}
	}
	l.ensureCursorVisible(visible)

	var b strings.Builder
	end := l.offset + visible
	if end > len(l.stories) {
		end = len(l.stories)
	}
	for i := l.offset; i < end; i++ {
		if i > l.offset {
			b.WriteString("\n\n")
		}
		b.WriteString(l.renderRow(i, width))
	}
	body := b.String()
	if l.searching {
		body += "\n" + style.Status.Render("/"+l.query+"█")
	}
	return body
}

func (l *listScreen) ensureCursorVisible(visible int) {
	if l.cursor < l.offset {
		l.offset = l.cursor
	}
	if l.cursor >= l.offset+visible {
		l.offset = l.cursor - visible + 1
	}
	if l.offset < 0 {
		l.offset = 0
	}
}

func (l *listScreen) renderRow(i, width int) string {
	s := l.stories[i]
	marker := "  "
	if i == l.cursor {
		marker = "▸ "
	}

	titleStyle := style.Title
	if i == l.cursor {
		titleStyle = selectedStyle()
	}
	if l.isRead(s.ShortID) {
		titleStyle = style.TitleRead
	}
	title := titleStyle.Render(s.Title)
	if domain := host(s.URL); domain != "" {
		title += " " + style.Domain.Render("("+domain+")")
	}

	meta := []string{
		style.Score.Render(fmt.Sprintf("%d pts", s.Score)),
		style.Comments.Render(fmt.Sprintf("%d comments", s.CommentCount)),
		style.Submitter.Render(s.SubmitterUser),
		style.Meta.Render(timeago.Since(s.CreatedAt)),
	}
	if len(s.Tags) > 0 {
		meta = append(meta, style.Tag.Render("["+strings.Join(s.Tags, " ")+"]"))
	}

	line1 := marker + title
	line2 := "  " + strings.Join(meta, style.Meta.Render(" | "))
	return ansi.Truncate(line1, width, "…") + "\n" + ansi.Truncate(line2, width, "…")
}

// selectedStyle is the title style of the cursor row: bold white unless the
// theme sets a selection color.
func selectedStyle() lipgloss.Style {
	st := style.Selection
	if _, unset := st.GetForeground().(lipgloss.NoColor); unset {
		st = st.Foreground(lipgloss.Color("15"))
	}
	return st
}

// markRead records a visit for a story opened outside the in-app comments view.
func (l *listScreen) markRead(s lobsters.Story) {
	if l.hist != nil && !l.noHist {
		_ = l.hist.MarkIfUnread(s.ShortID, s.CommentCount, time.Now())
	}
}

func (l *listScreen) isRead(id string) bool {
	return l.hist != nil && !l.noHist && l.hist.IsRead(id)
}

func host(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return ""
	}
	return strings.TrimPrefix(u.Host, "www.")
}

func loadFeedCmd(client *lobsters.Client, cat categories.Category, fav *favorites.Store) tea.Cmd {
	return func() tea.Msg {
		stories, err := fetchFeed(context.Background(), client, cat, fav)
		return feedLoadedMsg{key: cat.Key(), stories: stories, err: err}
	}
}

func fetchFeed(ctx context.Context, client *lobsters.Client, cat categories.Category, fav *favorites.Store) ([]lobsters.Story, error) {
	switch cat.Kind {
	case categories.KindHottest:
		return client.Hottest(ctx)
	case categories.KindNewest:
		return client.Newest(ctx)
	case categories.KindActive:
		return client.Active(ctx)
	case categories.KindTag:
		return client.Tag(ctx, cat.Tag)
	case categories.KindFavorites:
		if fav == nil {
			return nil, nil
		}
		items := fav.List()
		stories := make([]lobsters.Story, len(items))
		for i, it := range items {
			stories[i] = it.Story()
		}
		return stories, nil
	default:
		return nil, nil
	}
}

func openURL(raw string) tea.Cmd {
	return func() tea.Msg {
		if raw == "" {
			return nil
		}
		_ = browser.OpenURL(raw)
		return nil
	}
}
