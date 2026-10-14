package view

import (
	"net/url"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"noben.org/lbx/lobsters"
	"noben.org/lbx/render"
	"noben.org/lbx/style"
)

// linkPicker is an in-place link selector shared by the comment and reader
// screens.
type linkPicker struct {
	links  []render.Link
	active bool
	sel    int
	offset int
}

// toggle opens the picker with the given links, or closes it when already open.
func (p *linkPicker) toggle(links []render.Link) {
	if p.active {
		p.active = false
		return
	}
	p.links = links
	p.sel = 0
	p.offset = 0
	p.active = len(links) > 0
}

func (p *linkPicker) move(delta int) {
	if len(p.links) == 0 {
		return
	}
	p.sel = clamp(p.sel+delta, 0, len(p.links)-1)
}

func (p *linkPicker) selected() (render.Link, bool) {
	if !p.active || p.sel < 0 || p.sel >= len(p.links) {
		return render.Link{}, false
	}
	return p.links[p.sel], true
}

// lines renders the picker as overlay lines fitted to width and height.
func (p *linkPicker) lines(width, height int) []string {
	if !p.active || len(p.links) == 0 {
		return nil
	}
	maxRows := height - 2
	if maxRows < 1 {
		maxRows = 1
	}
	if p.sel < p.offset {
		p.offset = p.sel
	}
	if p.sel >= p.offset+maxRows {
		p.offset = p.sel - maxRows + 1
	}

	out := []string{style.Header.Render("links")}
	end := p.offset + maxRows
	if end > len(p.links) {
		end = len(p.links)
	}
	for i := p.offset; i < end; i++ {
		label := p.links[i].Text
		if label == "" {
			label = p.links[i].Href
		}
		line := label + "  " + style.Meta.Render(p.links[i].Href)
		if i == p.sel {
			line = style.Selection.Reverse(true).Render(line)
		}
		out = append(out, ansi.Truncate(line, width, "…"))
	}
	return out
}

// handleKey routes a key to the picker when it is open. It returns true when
// the key was consumed.
func (p *linkPicker) handleKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	if !p.active {
		return nil, false
	}
	switch msg.String() {
	case "j", "down":
		p.move(1)
	case "k", "up":
		p.move(-1)
	case "enter":
		if link, ok := p.selected(); ok {
			return openLinkCmd(link), true
		}
	case "tab", "q", "esc":
		p.active = false
	}
	return nil, true
}

// openLinkCmd opens a link in place: lobste.rs story links become the comment
// section, everything else opens in reader mode.
func openLinkCmd(link render.Link) tea.Cmd {
	return func() tea.Msg {
		if id, ok := lobstersStoryID(link.Href); ok {
			return openCommentsMsg{story: lobsters.Story{ShortID: id, Title: link.Text, CommentsURL: link.Href, URL: link.Href}}
		}
		return openReaderMsg{story: lobsters.Story{Title: link.Text, URL: link.Href}}
	}
}

func lobstersStoryID(raw string) (string, bool) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", false
	}
	if u.Host != "" && u.Host != "lobste.rs" && u.Host != "www.lobste.rs" {
		return "", false
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) >= 2 && parts[0] == "s" && parts[1] != "" {
		return parts[1], true
	}
	return "", false
}

// overlayTop replaces the first lines of base with the overlay lines.
func overlayTop(base string, overlay []string) string {
	if len(overlay) == 0 {
		return base
	}
	baseLines := strings.Split(base, "\n")
	for i, line := range overlay {
		if i >= len(baseLines) {
			break
		}
		baseLines[i] = line
	}
	return strings.Join(baseLines, "\n")
}
