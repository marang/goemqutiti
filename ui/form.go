package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/emqutiti/constants"
)

// Field represents a single input element used by forms.
type Field interface {
	Focus()
	Blur()
	Update(msg tea.Msg) tea.Cmd
	View() string
	Value() string
	ReadOnly() bool
	SetReadOnly(bool)
}

// KeyConsumer reports whether a field wants to handle a key itself instead of
// having the form advance focus.
type KeyConsumer interface {
	WantsKey(tea.KeyMsg) bool
}

// Form groups a slice of Field and tracks which one has focus.
type Form struct {
	Fields []Field
	Focus  int
}

// CycleFocus moves focus based on the provided key message.
func (f *Form) CycleFocus(msg tea.KeyMsg) {
	if c, ok := f.Fields[f.Focus].(KeyConsumer); ok && c.WantsKey(msg) {
		return
	}
	switch msg.String() {
	case constants.KeyTab, constants.KeyDown, constants.KeyJ:
		f.Focus++
	case constants.KeyShiftTab, constants.KeyUp, constants.KeyK:
		f.Focus--
	default:
		return
	}
	if f.Focus < 0 {
		f.Focus = len(f.Fields) - 1
	}
	if f.Focus >= len(f.Fields) {
		f.Focus = 0
	}
}

// ApplyFocus calls Focus on the active field and Blur on all others.
func (f *Form) ApplyFocus() {
	for i := range f.Fields {
		if i == f.Focus {
			f.Fields[i].Focus()
		} else {
			f.Fields[i].Blur()
		}
	}
}

// IsFocused reports whether the field at idx is focused and editable.
func (f *Form) IsFocused(idx int) bool {
	if idx != f.Focus || idx < 0 || idx >= len(f.Fields) {
		return false
	}
	return !f.Fields[idx].ReadOnly()
}
