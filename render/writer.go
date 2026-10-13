package render

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Line is one rendered terminal line with light metadata.
type Line struct {
	Text    string
	Heading bool
}

// lineWriter accumulates styled words into width-wrapped lines with an
// optional per-line prefix (indentation, quote markers).
type lineWriter struct {
	width   int
	prefix  string
	heading bool

	lines   []Line
	segs    []string
	curW    int
	started bool
}

func (w *lineWriter) budget() int {
	b := w.width - lipgloss.Width(w.prefix)
	if b < 1 {
		b = 1
	}
	return b
}

// word appends a styled word, wrapping to a new line when it would overflow
// and separating consecutive words with a single space.
func (w *lineWriter) word(seg string) {
	if seg == "" {
		return
	}
	sw := lipgloss.Width(seg)
	if w.started && w.curW+1+sw > w.budget() {
		w.newline()
	}
	if w.started {
		w.segs = append(w.segs, " ")
		w.curW++
	}
	w.segs = append(w.segs, seg)
	w.curW += sw
	w.started = true
}

// newline ends the current line, emitting a blank line if nothing was written.
func (w *lineWriter) newline() {
	if w.started {
		w.lines = append(w.lines, Line{Text: w.prefix + strings.Join(w.segs, ""), Heading: w.heading})
	} else {
		w.lines = append(w.lines, Line{Heading: w.heading})
	}
	w.segs = w.segs[:0]
	w.curW = 0
	w.started = false
}

// flush ends the current line if any words are pending.
func (w *lineWriter) flush() {
	if w.started {
		w.newline()
	}
}

// blank ensures a single empty line separates the previous block.
func (w *lineWriter) blank() {
	w.flush()
	if n := len(w.lines); n > 0 && w.lines[n-1].Text != "" {
		w.lines = append(w.lines, Line{})
	}
}

// rawLine appends a line verbatim (used for preformatted code).
func (w *lineWriter) rawLine(s string) {
	w.flush()
	w.lines = append(w.lines, Line{Text: w.prefix + s})
}
