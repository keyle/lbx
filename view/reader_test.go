package view

import (
	"strings"
	"testing"

	"noben.org/lbx/article"
	"noben.org/lbx/lobsters"
)

func TestReaderReflowsOnResize(t *testing.T) {
	r := newReaderScreen(lobsters.Story{Title: "T", URL: "https://e.com/x"})
	r.art = &article.Article{Title: "T", Content: "<p>" + strings.Repeat("word ", 200) + "</p>"}

	r.View(120, 20)
	wide := len(r.lines)
	r.View(40, 20)
	narrow := len(r.lines)
	if narrow <= wide {
		t.Errorf("narrowing should wrap into more lines: wide=%d narrow=%d", wide, narrow)
	}
}

func TestReaderBrowserURL(t *testing.T) {
	r := newReaderScreen(lobsters.Story{URL: "https://e.com/x"})
	if got := r.browserURL(); got != "https://e.com/x" {
		t.Errorf("browserURL = %q", got)
	}
	r = newReaderScreen(lobsters.Story{CommentsURL: "https://lobste.rs/s/a"})
	if got := r.browserURL(); got != "https://lobste.rs/s/a" {
		t.Errorf("browserURL fallback = %q", got)
	}
}
