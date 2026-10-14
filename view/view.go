package view

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"noben.org/lbx/categories"
	"noben.org/lbx/favorites"
	"noben.org/lbx/header"
	"noben.org/lbx/help"
	"noben.org/lbx/history"
	"noben.org/lbx/keymap"
	"noben.org/lbx/lobsters"
)

// globalKeymap holds keys the root handles before dispatching to a screen.
var globalKeymap = keymap.Map{
	{Name: "help", Keys: []string{"?", "i"}},
	{Name: "wide", Keys: []string{"z"}},
}

// inputCapturer is implemented by screens that are collecting typed input, so
// the root suspends global keys (except Ctrl-C) while they are active.
type inputCapturer interface{ CapturingInput() bool }

// wideMin is the column count at which dual-pane mode becomes available.
const wideMin = 180

// Options configures the root model.
type Options struct {
	Client     *lobsters.Client
	Categories []categories.Category
	WideView   string // "auto" (default), "never", or "always"
	Favorites  *favorites.Store
	History    *history.Store
	NoHistory  bool
}

// Model is the root Bubble Tea model. It owns a stack of screens; the last one
// is active.
type Model struct {
	client *lobsters.Client
	cats   []categories.Category
	stack  []Screen
	fav    *favorites.Store
	hist   *history.Store
	noHist bool

	activeCat int
	width     int
	height    int
	showHelp  bool

	wideMode string
	wide     bool
	wideSet  bool
}

// New builds the root model with a list screen at the bottom of the stack.
func New(opts Options) Model {
	cats := opts.Categories
	if len(cats) == 0 {
		cats = categories.Default()
	}
	mode := opts.WideView
	if mode == "" {
		mode = "auto"
	}
	m := Model{client: opts.Client, cats: cats, wideMode: mode, fav: opts.Favorites, hist: opts.History, noHist: opts.NoHistory}
	m.stack = []Screen{newListScreen(opts.Client, cats, opts.Favorites, opts.History, opts.NoHistory)}
	return m
}

func (m Model) Init() tea.Cmd { return m.stack[0].Init() }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.updateWide()
		return m, nil
	case spinner.TickMsg:
		// Every spinner in the stack gets ticks, so they keep animating
		// even when they are not the top screen. Each one ignores ticks that are not its own.
		for i, s := range m.stack {
			next, cmd := s.Update(msg)
			m.stack[i] = next
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	case categoryChangedMsg:
		m.activeCat = msg.index
		return m, nil
	case popMsg:
		if len(m.stack) > 1 {
			m.stack = m.stack[:len(m.stack)-1]
		}
		return m, nil
	case quitMsg:
		return m, tea.Quit
	case openCommentsMsg:
		var since time.Time
		if !m.noHist && m.hist != nil {
			if e, ok := m.hist.Entry(msg.story.ShortID); ok {
				since = e.LastVisit
			}
			_ = m.hist.Mark(msg.story.ShortID, msg.story.CommentCount, time.Now())
		}
		m.stack = append(m.stack, newCommentsScreen(m.client, msg.story, since))
		return m, m.stack[len(m.stack)-1].Init()
	case openReaderMsg:
		if !m.noHist && m.hist != nil {
			_ = m.hist.MarkIfUnread(msg.story.ShortID, msg.story.CommentCount, time.Now())
		}
		m.stack = append(m.stack, newReaderScreen(msg.story))
		return m, m.stack[len(m.stack)-1].Init()
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if !m.capturingInput() {
			switch globalKeymap.Action(msg) {
			case "help":
				m.showHelp = !m.showHelp
				return m, nil
			case "wide":
				m.wideSet = true
				m.wide = !m.wide
				return m, nil
			}
		}
		if m.showHelp {
			if msg.String() == "esc" || msg.String() == "q" {
				m.showHelp = false
			}
			return m, nil
		}
	}

	if len(m.stack) == 0 {
		return m, nil
	}
	top := m.stack[len(m.stack)-1]
	next, cmd := top.Update(msg)
	m.stack[len(m.stack)-1] = next
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *Model) capturingInput() bool {
	if len(m.stack) == 0 {
		return false
	}
	c, ok := m.stack[len(m.stack)-1].(inputCapturer)
	return ok && c.CapturingInput()
}

func (m *Model) updateWide() {
	switch {
	case m.wideMode == "never":
		m.wide = false
	case m.wideSet:
		// keep the user's session choice
	case m.wideMode == "always":
		m.wide = true
	default:
		m.wide = m.width >= wideMin
	}
	if m.width < wideMin {
		m.wide = false
	}
}

func (m Model) View() string {
	if m.width == 0 || len(m.stack) == 0 {
		return ""
	}
	top := m.stack[len(m.stack)-1]
	head := header.Render(m.width, m.cats, m.activeCat, top.Title())

	// One row for the header and one blank padding row below it.
	bodyHeight := m.height - 2
	if bodyHeight < 1 {
		bodyHeight = 1
	}

	if m.showHelp {
		keysTitle := "keys"
		if t := top.Title(); t != "" {
			keysTitle += " — " + t
		}
		box := help.Render(keysTitle, append(top.Help(), help.Binding{Keys: "z", Desc: "toggle wide view"}))
		return head + "\n\n" + lipgloss.Place(m.width, bodyHeight, lipgloss.Center, lipgloss.Center, box)
	}

	if m.wide && m.width >= wideMin {
		return head + "\n\n" + m.wideBody(top, m.width, bodyHeight)
	}
	pad := padFor(top)
	return head + "\n\n" + indent(top.View(m.width-pad, bodyHeight), pad)
}

// Left padding per screen kind.
const (
	listPad     = 1
	commentsPad = 3
	contentsPad = 3
	defaultPad  = 2
)

// padFor returns the left padding for a screen.
func padFor(s Screen) int {
	switch s.(type) {
	case *listScreen:
		return listPad
	case *commentsScreen:
		return commentsPad
	case *readerScreen:
		return contentsPad
	}
	return defaultPad
}

// indent prefixes every line of s with n spaces.
func indent(s string, n int) string {
	pad := strings.Repeat(" ", n)
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = pad + l
	}
	return strings.Join(lines, "\n")
}

func (m Model) wideBody(top Screen, width, height int) string {
	listW := width * 2 / 5
	if listW < 40 {
		listW = 40
	}
	rightW := width - listW
	if rightW < 20 {
		rightW = 20
	}
	left := indent(m.stack[0].View(listW-listPad, height), listPad)
	right := m.rightPane(top, rightW, height)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m Model) rightPane(top Screen, width, height int) string {
	if len(m.stack) > 1 {
		pad := padFor(top)
		return indent(top.View(width-pad, height), pad)
	}
	return indent(lipgloss.NewStyle().Faint(true).Render("select a story, then press Enter to read its comments"), commentsPad)
}
