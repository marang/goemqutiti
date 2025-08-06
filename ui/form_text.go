package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/emqutiti/constants"
)

// TextField wraps a text input with optional read-only behaviour.
type TextField struct {
	textinput.Model
	readOnly bool
}

// NewTextField creates a TextField with the given value and placeholder.
// If opts[0] is true the field is masked for password entry.
func NewTextField(value, placeholder string, opts ...bool) *TextField {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.SetValue(value)
	if len(opts) > 0 && opts[0] {
		ti.EchoMode = textinput.EchoPassword
	}
	return &TextField{Model: ti}
}

// SetReadOnly marks the field read only and blurs it when activated.
func (t *TextField) SetReadOnly(ro bool) {
	t.readOnly = ro
	if ro {
		t.Blur()
	}
}

// ReadOnly reports whether the field is read only.
func (t *TextField) ReadOnly() bool { return t.readOnly }

// Update forwards messages to the text input unless the field is read only.
func (t *TextField) Update(msg tea.Msg) tea.Cmd {
	if t.readOnly {
		return nil
	}
	var cmd tea.Cmd
	t.Model, cmd = t.Model.Update(msg)
	return cmd
}

// Value returns the text content of the field.
func (t *TextField) Value() string { return t.Model.Value() }
func (t *TextField) Focus() {
	if !t.readOnly {
		t.Model.Focus()
	}
}
func (t *TextField) Blur()        { t.Model.Blur() }
func (t *TextField) View() string { return t.Model.View() }

// WantsKey reports whether the field wants to handle navigation keys itself
// instead of letting the form cycle focus. Plain "j" and "k" are treated as
// normal input so users can type them without jumping to another field.
func (t *TextField) WantsKey(k tea.KeyMsg) bool {
	switch k.String() {
	case constants.KeyJ, constants.KeyK:
		return true
	default:
		return false
	}
}
