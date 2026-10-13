package render

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/net/html"
)

// Styles holds the terminal styles used when rendering HTML.
type Styles struct {
	Bold      lipgloss.Style
	Italic    lipgloss.Style
	Code      lipgloss.Style
	CodeBlock lipgloss.Style
	Link      lipgloss.Style
	Heading   lipgloss.Style
	Quote     lipgloss.Style
	Bullet    lipgloss.Style
}

// DefaultStyles returns the built-in palette.
func DefaultStyles() Styles {
	return Styles{
		Bold:      lipgloss.NewStyle().Bold(true),
		Italic:    lipgloss.NewStyle().Italic(true),
		Code:      lipgloss.NewStyle().Foreground(lipgloss.Color("215")),
		CodeBlock: lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		Link:      lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Underline(true),
		Heading:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")),
		Quote:     lipgloss.NewStyle().Foreground(lipgloss.Color("244")),
		Bullet:    lipgloss.NewStyle().Foreground(lipgloss.Color("244")),
	}
}

// Renderer converts HTML into styled, width-wrapped terminal text.
type Renderer struct {
	Width  int
	Styles Styles
}

// New returns a renderer that wraps to width.
func New(width int) *Renderer {
	if width < 1 {
		width = 80
	}
	return &Renderer{Width: width, Styles: DefaultStyles()}
}

// HTML renders an HTML fragment to terminal text at the given width.
func HTML(input string, width int) string { return New(width).Render(input) }

type context struct {
	bold    bool
	italic  bool
	code    bool
	link    bool
	pre     bool
	heading bool
}

// Render converts an HTML fragment into styled terminal text.
func (r *Renderer) Render(input string) string {
	lines := r.Lines(input)
	texts := make([]string, len(lines))
	for i, line := range lines {
		texts[i] = line.Text
	}
	return strings.Join(texts, "\n")
}

// Lines converts an HTML fragment into rendered lines, tagging headings so
// callers (reader mode) can navigate between them.
func (r *Renderer) Lines(input string) []Line {
	doc, err := html.Parse(strings.NewReader(input))
	if err != nil {
		return []Line{{Text: strings.TrimSpace(input)}}
	}
	lw := &lineWriter{width: r.Width}
	r.walk(doc, lw, context{})
	lw.flush()
	return trimBlankEdges(lw.lines)
}

func (r *Renderer) walk(n *html.Node, lw *lineWriter, c context) {
	switch n.Type {
	case html.TextNode:
		r.text(n.Data, lw, c)
	case html.ElementNode:
		r.element(n, lw, c)
	default:
		r.children(n, lw, c)
	}
}

func (r *Renderer) children(n *html.Node, lw *lineWriter, c context) {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		r.walk(child, lw, c)
	}
}

func (r *Renderer) element(n *html.Node, lw *lineWriter, c context) {
	switch n.Data {
	case "script", "style", "head", "title", "noscript":
		return
	case "p", "div":
		lw.blank()
		r.children(n, lw, c)
		lw.blank()
	case "br":
		lw.newline()
	case "h1", "h2", "h3", "h4", "h5", "h6":
		lw.blank()
		c.bold = true
		c.heading = true
		lw.heading = true
		r.children(n, lw, c)
		lw.flush()
		lw.heading = false
		lw.blank()
	case "strong", "b":
		c.bold = true
		r.children(n, lw, c)
	case "em", "i":
		c.italic = true
		r.children(n, lw, c)
	case "code":
		if c.pre {
			r.children(n, lw, c)
			return
		}
		c.code = true
		r.children(n, lw, c)
	case "pre":
		r.pre(n, lw)
	case "a":
		c.link = true
		r.children(n, lw, c)
		r.appendHref(n, lw, c)
	case "ul", "ol":
		r.list(n, lw, c)
	case "blockquote":
		r.quote(n, lw, c)
	case "img":
		r.image(n, lw)
	default:
		r.children(n, lw, c)
	}
}

func (r *Renderer) text(s string, lw *lineWriter, c context) {
	style := r.inlineStyle(c)
	for _, word := range strings.Fields(s) {
		lw.word(style.Render(word))
	}
}

func (r *Renderer) inlineStyle(c context) lipgloss.Style {
	switch {
	case c.heading:
		return r.Styles.Heading
	case c.code:
		return r.Styles.Code
	case c.link:
		return r.Styles.Link
	case c.bold && c.italic:
		return lipgloss.NewStyle().Bold(true).Italic(true)
	case c.bold:
		return r.Styles.Bold
	case c.italic:
		return r.Styles.Italic
	default:
		return lipgloss.NewStyle()
	}
}

func (r *Renderer) pre(n *html.Node, lw *lineWriter) {
	code := strings.TrimRight(nodeText(n), "\n")
	lw.blank()
	prev := lw.prefix
	lw.prefix = "  "

	lines := highlightCode(code, codeLanguage(n))
	highlighted := lines != nil
	if !highlighted {
		lines = strings.Split(code, "\n")
	}
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			lw.rawLine("")
			continue
		}
		if highlighted {
			lw.rawLine(line)
			continue
		}
		lw.rawLine(r.Styles.CodeBlock.Render(line))
	}
	lw.prefix = prev
	lw.blank()
}

func (r *Renderer) list(n *html.Node, lw *lineWriter, c context) {
	lw.blank()
	prev := lw.prefix
	bullet := r.Styles.Bullet.Render("• ")
	for item := n.FirstChild; item != nil; item = item.NextSibling {
		if item.Type != html.ElementNode || item.Data != "li" {
			continue
		}
		lw.prefix = prev + "  "
		lw.word(bullet)
		r.children(item, lw, c)
		lw.flush()
	}
	lw.prefix = prev
	lw.blank()
}

func (r *Renderer) quote(n *html.Node, lw *lineWriter, c context) {
	lw.blank()
	prev := lw.prefix
	lw.prefix = prev + r.Styles.Quote.Render("> ")
	r.children(n, lw, c)
	lw.flush()
	lw.prefix = prev
	lw.blank()
}

func (r *Renderer) appendHref(n *html.Node, lw *lineWriter, c context) {
	for _, attr := range n.Attr {
		if attr.Key != "href" || attr.Val == "" {
			continue
		}
		if strings.HasPrefix(attr.Val, "#") {
			return
		}
		text := strings.TrimSpace(nodeText(n))
		if text == attr.Val || strings.Contains(text, attr.Val) {
			return
		}
		lw.word(r.Styles.Link.Render("(" + attr.Val + ")"))
		return
	}
}

func (r *Renderer) image(n *html.Node, lw *lineWriter) {
	alt := ""
	for _, attr := range n.Attr {
		if attr.Key == "alt" && attr.Val != "" {
			alt = attr.Val
		}
	}
	if alt == "" {
		alt = "image"
	}
	lw.word(r.Styles.Quote.Render("[image: " + alt + "]"))
}

func nodeText(n *html.Node) string {
	var b strings.Builder
	var collect func(*html.Node)
	collect = func(node *html.Node) {
		if node.Type == html.TextNode {
			b.WriteString(node.Data)
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			collect(child)
		}
	}
	collect(n)
	return b.String()
}

func trimBlankEdges(lines []Line) []Line {
	for len(lines) > 0 && strings.TrimSpace(lines[0].Text) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1].Text) == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
