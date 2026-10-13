package style

import (
	"github.com/charmbracelet/lipgloss"

	"noben.org/lbx/theme"
)

// Styles used across the UI. Colors are placeholders until theme.toml loading
// (milestone 6) overrides them via Apply.
var (
	Header      = lipgloss.NewStyle().Bold(true)
	HeaderMuted = lipgloss.NewStyle().Faint(true)

	CategoryOn  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	CategoryOff = lipgloss.NewStyle().Faint(true)

	Title     = lipgloss.NewStyle().Bold(true)
	TitleRead = lipgloss.NewStyle().Faint(true)

	Meta      = lipgloss.NewStyle().Faint(true)
	Domain    = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	Tag       = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	Score     = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	Comments  = lipgloss.NewStyle().Foreground(lipgloss.Color("111"))
	Submitter = lipgloss.NewStyle().Foreground(lipgloss.Color("141"))

	Selection = lipgloss.NewStyle().Bold(true)

	Status = lipgloss.NewStyle().Faint(true)
	Error  = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

// Apply overrides the default palette with a loaded theme. An empty color value
// leaves that element using the terminal's default.
func Apply(t theme.Theme) {
	CategoryOn = lipgloss.NewStyle().Bold(true).Foreground(color(t.CategoryOn))
	CategoryOff = lipgloss.NewStyle().Faint(true).Foreground(color(t.CategoryOff))
	Title = lipgloss.NewStyle().Bold(true).Foreground(color(t.Title))
	TitleRead = lipgloss.NewStyle().Faint(true).Foreground(color(t.TitleRead))
	Meta = lipgloss.NewStyle().Faint(true).Foreground(color(t.Meta))
	Domain = lipgloss.NewStyle().Foreground(color(t.Domain))
	Tag = lipgloss.NewStyle().Foreground(color(t.Tag))
	Score = lipgloss.NewStyle().Foreground(color(t.Score))
	Comments = lipgloss.NewStyle().Foreground(color(t.Comments))
	Submitter = lipgloss.NewStyle().Foreground(color(t.Submitter))
	Selection = lipgloss.NewStyle().Bold(true).Foreground(color(t.Selection))
	Error = lipgloss.NewStyle().Foreground(color(t.Error))
}

func color(v string) lipgloss.TerminalColor {
	if v == "" {
		return lipgloss.NoColor{}
	}
	return lipgloss.Color(v)
}
