package view

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"noben.org/lbx/article"
	"noben.org/lbx/help"
	"noben.org/lbx/keymap"
	"noben.org/lbx/lobsters"
	"noben.org/lbx/render"
	"noben.org/lbx/style"
)

// readerKeymap is reader mode's keybindings.
var readerKeymap = keymap.Map{
	{Name: "down", Keys: []string{"j", "down"}, Help: "j / k", Desc: "scroll down / up"},
	{Name: "up", Keys: []string{"k", "up"}},
	{Name: "halfDown", Keys: []string{"ctrl+d"}, Help: "Ctrl-D/Ctrl-U", Desc: "half-page down / up"},
	{Name: "halfUp", Keys: []string{"ctrl+u"}},
	{Name: "top", Keys: []string{"g", "home"}, Help: "g / G", Desc: "top / bottom"},
	{Name: "bottom", Keys: []string{"G", "end"}},
	{Name: "nextHeading", Keys: []string{"n"}, Help: "n / N", Desc: "next / previous heading or match"},
	{Name: "prevHeading", Keys: []string{"N"}},
	{Name: "search", Keys: []string{"/"}, Help: "/", Desc: "search in article"},
	{Name: "links", Keys: []string{"tab"}, Help: "Tab", Desc: "select link"},
	{Name: "browser", Keys: []string{"o"}, Help: "o", Desc: "open in browser"},
	{Name: "esc", Keys: []string{"esc"}, Help: "Esc", Desc: "clear search / back"},
	{Name: "back", Keys: []string{"q"}, Help: "q", Desc: "back"},
}

type readerScreen struct {
	client *http.Client
	story  lobsters.Story
	art    *article.Article

	loading bool
	spin    spinner.Model
	err     error
	dirty   bool

	lines  []render.Line
	scroll int

	width, height int

	searching bool
	query     string
	matches   []int
	matchIdx  int

	picker linkPicker
}

func newReaderScreen(story lobsters.Story) *readerScreen {
	return &readerScreen{
		client: &http.Client{Timeout: 20 * time.Second},
		story:  story,
		spin:   newSpinner(),
	}
}

func (r *readerScreen) Init() tea.Cmd {
	r.loading = true
	client := r.client
	url := r.story.URL
	load := func() tea.Msg {
		if url == "" {
			return articleLoadedMsg{err: article.ErrNotReadable}
		}
		art, err := article.Fetch(context.Background(), client, url)
		return articleLoadedMsg{art: art, err: err}
	}
	return tea.Batch(load, r.spin.Tick)
}

func (r *readerScreen) Title() string {
	if r.art != nil && r.art.Title != "" {
		return r.art.Title
	}
	return r.story.Title
}

func (r *readerScreen) Help() []help.Binding { return readerKeymap.Help() }

// CapturingInput reports whether the in-article search prompt is active.
func (r *readerScreen) CapturingInput() bool { return r.searching }

func (r *readerScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		r.spin, cmd = r.spin.Update(msg)
		return r, cmd
	case articleLoadedMsg:
		r.loading = false
		r.err = msg.err
		if msg.err == nil && msg.art != nil {
			r.art = msg.art
			r.dirty = true
		}
		return r, nil
	case tea.KeyMsg:
		return r.handleKey(msg)
	}
	return r, nil
}

func (r *readerScreen) handleKey(msg tea.KeyMsg) (Screen, tea.Cmd) {
	if r.searching {
		return r.handleSearchKey(msg)
	}
	if cmd, handled := r.picker.handleKey(msg); handled {
		return r, cmd
	}
	switch readerKeymap.Action(msg) {
	case "back":
		return r, func() tea.Msg { return popMsg{} }
	case "esc":
		if len(r.matches) > 0 {
			r.matches = nil
			return r, nil
		}
		return r, func() tea.Msg { return popMsg{} }
	case "down":
		r.scrollBy(1)
	case "up":
		r.scrollBy(-1)
	case "halfDown":
		r.scrollBy(r.pageSize() / 2)
	case "halfUp":
		r.scrollBy(-r.pageSize() / 2)
	case "top":
		r.scroll = 0
	case "bottom":
		r.scroll = r.maxScroll()
	case "nextHeading":
		r.step(1)
	case "prevHeading":
		r.step(-1)
	case "search":
		r.searching = true
		r.query = ""
	case "links":
		if r.art != nil {
			r.picker.toggle(render.Links(r.art.Content))
		}
	case "browser":
		return r, openURL(r.browserURL())
	}
	return r, nil
}

func (r *readerScreen) handleSearchKey(msg tea.KeyMsg) (Screen, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		r.searching = false
		r.commitSearch()
	case tea.KeyEsc:
		r.searching = false
		r.query = ""
	case tea.KeyBackspace:
		if runes := []rune(r.query); len(runes) > 0 {
			r.query = string(runes[:len(runes)-1])
		}
	case tea.KeyRunes, tea.KeySpace:
		r.query += string(msg.Runes)
	}
	return r, nil
}

func (r *readerScreen) commitSearch() {
	r.matches = nil
	r.matchIdx = 0
	q := strings.ToLower(strings.TrimSpace(r.query))
	if q == "" {
		return
	}
	for i, line := range r.lines {
		if strings.Contains(strings.ToLower(ansi.Strip(line.Text)), q) {
			r.matches = append(r.matches, i)
		}
	}
	if len(r.matches) > 0 {
		r.scrollTo(r.matches[0])
	}
}

// browserURL is the URL the reader offers to the system browser: the article
// itself, or the discussion page for a story without an external URL.
func (r *readerScreen) browserURL() string {
	if r.story.URL != "" {
		return r.story.URL
	}
	return r.story.CommentsURL
}

func (r *readerScreen) step(delta int) {
	if len(r.matches) > 0 {
		r.matchIdx = clamp(r.matchIdx+delta, 0, len(r.matches)-1)
		r.scrollTo(r.matches[r.matchIdx])
		return
	}
	heads := r.headingIndices()
	if len(heads) == 0 {
		return
	}
	if delta > 0 {
		for _, h := range heads {
			if h > r.scroll {
				r.scrollTo(h)
				return
			}
		}
		return
	}
	for i := len(heads) - 1; i >= 0; i-- {
		if heads[i] < r.scroll {
			r.scrollTo(heads[i])
			return
		}
	}
}

func (r *readerScreen) headingIndices() []int {
	var out []int
	for i, line := range r.lines {
		if line.Heading {
			out = append(out, i)
		}
	}
	return out
}

func (r *readerScreen) scrollTo(n int) {
	r.scroll = clamp(n, 0, r.maxScroll())
}

func (r *readerScreen) scrollBy(delta int) {
	r.scroll = clamp(r.scroll+delta, 0, r.maxScroll())
}

func (r *readerScreen) pageSize() int {
	if r.height < 1 {
		return 1
	}
	return r.height
}

func (r *readerScreen) maxScroll() int {
	if m := len(r.lines) - r.pageSize(); m > 0 {
		return m
	}
	return 0
}

func (r *readerScreen) rebuild() {
	r.lines = nil
	r.lines = append(r.lines, render.Line{Text: style.Title.Render(r.art.Title), Heading: true})
	if r.art.Byline != "" {
		r.lines = append(r.lines, render.Line{Text: style.Meta.Render(r.art.Byline)})
	}
	r.lines = append(r.lines, render.Line{})
	r.lines = append(r.lines, render.New(r.width).Lines(r.art.Content)...)
	r.dirty = false
}

func (r *readerScreen) View(width, height int) string {
	if width != r.width {
		r.dirty = true
	}
	r.width, r.height = width, height
	if r.err != nil {
		if r.art == nil {
			return style.Error.Render("cannot read article: " + r.err.Error() + "\n\npress o to open it in your browser")
		}
	}
	if r.loading {
		return loadingView(r.spin, "loading article…")
	}
	if r.art == nil {
		return style.Status.Render("no article")
	}
	if r.dirty || len(r.lines) == 0 {
		r.rebuild()
	}
	if r.maxScroll() < r.scroll {
		r.scroll = r.maxScroll()
	}

	bodyHeight := height
	if r.searching {
		bodyHeight--
	}
	if bodyHeight < 1 {
		bodyHeight = 1
	}
	end := r.scroll + bodyHeight
	if end > len(r.lines) {
		end = len(r.lines)
	}

	var b strings.Builder
	for i := r.scroll; i < end; i++ {
		if i > r.scroll {
			b.WriteString("\n")
		}
		line := r.lines[i].Text
		if r.matchIndex(i) >= 0 {
			line = style.Selection.Reverse(true).Render(line)
		}
		b.WriteString(ansi.Truncate(line, width, "…"))
	}

	if r.searching {
		b.WriteString("\n" + style.Status.Render("/"+r.query+"█"))
	}
	body := b.String()
	if r.picker.active {
		body = overlayTop(body, r.picker.lines(width, height/2))
	}
	return body
}

func (r *readerScreen) matchIndex(line int) int {
	for i, m := range r.matches {
		if m == line {
			return i
		}
	}
	return -1
}
