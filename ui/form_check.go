package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// CheckField is a boolean toggle input.
type CheckField struct {
	value    bool
	focused  bool
	readOnly bool
}

// NewCheckField creates a CheckField with initial value.
func NewCheckField(val bool) *CheckField { return &CheckField{value: val} }

func (c *CheckField) Focus()              { c.focused = true }
func (c *CheckField) Blur()               { c.focused = false }
func (c *CheckField) SetReadOnly(ro bool) { c.readOnly = ro }

// ReadOnly reports whether the field is read only.
func (c *CheckField) ReadOnly() bool { return c.readOnly }

// Bool returns the boolean state of the field.
func (c *CheckField) Bool() bool { return c.value }

// SetBool updates the boolean state.
func (c *CheckField) SetBool(v bool) { c.value = v }

func (c *CheckField) Update(msg tea.Msg) tea.Cmd {
	switch m := msg.(type) {
	case tea.KeyMsg:
		if !c.readOnly && m.String() == " " {
			c.value = !c.value
		}
	case tea.MouseMsg:
		if !c.readOnly && m.Action == tea.MouseActionPress && m.Button == tea.MouseButtonLeft {
			c.value = !c.value
		}
	}
	return nil
}

func (c *CheckField) View() string {
	box := "[ ]"
	if c.value {
		box = "[x]"
	}
	switch {
	case c.readOnly:
		return BlurredStyle.Render(box)
	case c.focused:
		return FocusedStyle.Render(box)
	default:
		return box
	}
}

func (c *CheckField) Value() string { return fmt.Sprintf("%v", c.value) }
