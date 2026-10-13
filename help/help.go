package help

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Binding is a single key description shown in the help overlay.
type Binding struct {
	Keys string
	Desc string
}

// Render draws the keybindings for a view inside a bordered box.
func Render(title string, bindings []Binding) string {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Render(title))
	b.WriteString("\n\n")
	for i, bd := range bindings {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(lipgloss.NewStyle().Bold(true).Width(12).Render(bd.Keys))
		b.WriteString(lipgloss.NewStyle().Faint(true).Render(bd.Desc))
	}
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 2).Render(b.String())
}
