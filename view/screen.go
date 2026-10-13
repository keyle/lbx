package view

import (
	tea "github.com/charmbracelet/bubbletea"

	"noben.org/lbx/help"
)

// Screen is one view in the stack. Screens handle their own keys and emit
// popMsg/quitMsg when they want the stack to change.
type Screen interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (Screen, tea.Cmd)
	View(width, height int) string
	Title() string
	Help() []help.Binding
}
