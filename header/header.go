package header

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"noben.org/lbx/categories"
	"noben.org/lbx/style"
	"noben.org/lbx/version"
)

// Render draws the top bar: the app name and version, an optional muted status
// after them, and the category tabs right-aligned, fitted to width.
func Render(width int, cats []categories.Category, active int, title string) string {
	var tabs []string
	for i, c := range cats {
		if i == active {
			tabs = append(tabs, style.CategoryOn.Render("["+c.Label()+"]"))
			continue
		}
		tabs = append(tabs, style.CategoryOff.Render(c.Label()))
	}

	left := style.Header.Render("🦞 lbx") + " " + style.HeaderMuted.Render(version.Version)
	if title != "" {
		left += "  " + style.HeaderMuted.Render(title)
	}
	right := strings.Join(tabs, " ")
	rightW := lipgloss.Width(right)

	// Keep the right-aligned category tabs visible; truncate the left side to
	// make room, and fall back to truncating the tabs only if they cannot fit.
	if rightW >= width {
		return ansi.Truncate(right, width, "…")
	}
	if avail := width - rightW - 1; lipgloss.Width(left) > avail {
		left = ansi.Truncate(left, avail, "…")
	}

	pad := width - lipgloss.Width(left) - rightW
	if pad < 1 {
		pad = 1
	}
	return left + strings.Repeat(" ", pad) + right
}
