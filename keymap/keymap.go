// Package keymap defines a view's keybindings once, so key dispatch and the
// help overlay share a single source of truth.
package keymap

import (
	tea "github.com/charmbracelet/bubbletea"

	"noben.org/lbx/help"
)

// Binding is one action and the keys that trigger it.
type Binding struct {
	Name string   // stable id used for dispatch
	Keys []string // tea.KeyMsg.String() values
	Help string   // display text for the overlay; empty hides the binding
	Desc string   // description shown next to Help
}

// Matches reports whether msg triggers the binding.
func (b Binding) Matches(msg tea.KeyMsg) bool {
	s := msg.String()
	for _, k := range b.Keys {
		if k == s {
			return true
		}
	}
	return false
}

// Map is an ordered set of bindings for one view.
type Map []Binding

// Action returns the name of the binding triggered by msg, or "".
func (m Map) Action(msg tea.KeyMsg) string {
	for _, b := range m {
		if b.Matches(msg) {
			return b.Name
		}
	}
	return ""
}

// Help renders the map as overlay rows, skipping bindings without Help text.
func (m Map) Help() []help.Binding {
	out := make([]help.Binding, 0, len(m))
	for _, b := range m {
		if b.Help == "" {
			continue
		}
		out = append(out, help.Binding{Keys: b.Help, Desc: b.Desc})
	}
	return out
}
