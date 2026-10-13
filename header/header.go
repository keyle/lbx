package header

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"noben.org/lbx/categories"
	"noben.org/lbx/style"
)

// Version is the app version shown next to the name in the top bar. The
// makefile sets it at build time via -ldflags "-X noben.org/lbx/header.Version=...".
var Version = "dev"

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

	left := style.Header.Render("🦞 lbx") + " " + style.HeaderMuted.Render(Version)
	if title != "" {
		left += "  " + style.HeaderMuted.Render(title)
	}
	right := strings.Join(tabs, " ")

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	pad := width - leftW - rightW
	if pad < 1 {
		pad = 1
	}
	return left + strings.Repeat(" ", pad) + right
}
