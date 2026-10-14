package header

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"noben.org/lbx/categories"
)

func TestRenderFitsWidth(t *testing.T) {
	cats := categories.Default()
	title := strings.Repeat("a very long story title ", 10)
	for _, w := range []int{20, 40, 80, 200} {
		out := Render(w, cats, 0, title)
		if got := lipgloss.Width(out); got > w {
			t.Errorf("width %d: rendered line width %d exceeds it", w, got)
		}
	}
}

func TestRenderShowsTabs(t *testing.T) {
	out := Render(80, categories.Default(), 1, "")
	if !strings.Contains(out, "newest") {
		t.Errorf("category tabs missing: %q", out)
	}
}
