package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/goemqutiti/ui"
)

// formField represents a single input element used by forms.
type formField interface {
	Focus()
	Blur()
	Update(msg tea.Msg) tea.Cmd
	View() string
	Value() string
}

// keyConsumer reports whether a field wants to handle a key itself instead of
// having the form advance focus.
type keyConsumer interface {
	WantsKey(tea.KeyMsg) bool
}

// Form groups a slice of formField and tracks which one has focus.
type Form struct {
	fields []formField
	focus  int
}

// CycleFocus moves focus based on the provided key message.
func (f *Form) CycleFocus(msg tea.KeyMsg) {
	if c, ok := f.fields[f.focus].(keyConsumer); ok && c.WantsKey(msg) {
		return
	}
	switch msg.String() {
	case "tab", "down", "j":
		f.focus++
	case "shift+tab", "up", "k":
		f.focus--
	default:
		return
	}
	if f.focus < 0 {
		f.focus = len(f.fields) - 1
	}
	if f.focus >= len(f.fields) {
		f.focus = 0
	}
}

// ApplyFocus calls Focus on the active field and Blur on all others.
func (f *Form) ApplyFocus() {
	for i := range f.fields {
		if i == f.focus {
			f.fields[i].Focus()
		} else {
			f.fields[i].Blur()
		}
	}
}

// IsFocused reports whether the field at idx is focused and editable.
func (f *Form) IsFocused(idx int) bool {
	if idx != f.focus || idx < 0 || idx >= len(f.fields) {
		return false
	}
	switch fld := f.fields[idx].(type) {
	case *textField:
		return fld.Model.Focused() && !fld.readOnly
	case *selectField:
		return fld.focused && !fld.readOnly
	case *checkField:
		return fld.focused && !fld.readOnly
	default:
		return true
	}
}

// -------------------- field types --------------------

// textField wraps a text input with optional read-only behaviour.
type textField struct {
	textinput.Model
	readOnly bool
}

// newTextField creates a textField with the given value and placeholder.
// If opts[0] is true the field is masked for password entry.
func newTextField(value, placeholder string, opts ...bool) *textField {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.SetValue(value)
	if len(opts) > 0 && opts[0] {
		ti.EchoMode = textinput.EchoPassword
	}
	return &textField{Model: ti}
}

// setReadOnly marks the field read only and blurs it when activated.
func (t *textField) setReadOnly(ro bool) {
	t.readOnly = ro
	if ro {
		t.Blur()
	}
}

// Update forwards messages to the text input unless the field is read only.
func (t *textField) Update(msg tea.Msg) tea.Cmd {
	if t.readOnly {
		return nil
	}
	var cmd tea.Cmd
	t.Model, cmd = t.Model.Update(msg)
	return cmd
}

// Value returns the text content of the field.
func (t *textField) Value() string { return t.Model.Value() }

func (t *textField) Focus()       { t.Model.Focus() }
func (t *textField) Blur()        { t.Model.Blur() }
func (t *textField) View() string { return t.Model.View() }

// suggestField is a text input with auto-completion suggestions.
type suggestField struct {
	*textField
	options     []string
	suggestions []string
	sel         int
}

// newSuggestField creates a suggestField with the given options and placeholder.
func newSuggestField(opts []string, placeholder string) *suggestField {
	tf := newTextField("", placeholder)
	return &suggestField{
		textField:   tf,
		options:     append([]string(nil), opts...),
		suggestions: nil,
		sel:         -1,
	}
}

// Update processes key messages to cycle suggestions while forwarding other
// messages to the underlying text field.
func (s *suggestField) Update(msg tea.Msg) tea.Cmd {
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.String() {
		case "tab", "down":
			if len(s.suggestions) > 0 {
				s.sel = (s.sel + 1) % len(s.suggestions)
				s.SetValue(s.suggestions[s.sel])
				s.CursorEnd()
				return nil
			}
		case "shift+tab", "up":
			if len(s.suggestions) > 0 {
				s.sel--
				if s.sel < 0 {
					s.sel = len(s.suggestions) - 1
				}
				s.SetValue(s.suggestions[s.sel])
				s.CursorEnd()
				return nil
			}
		}
	}
	cmd := s.textField.Update(msg)
	if s.Focused() {
		prefix := s.Value()
		s.suggestions = s.suggestions[:0]
		s.sel = -1
		for _, o := range s.options {
			if prefix == "" || strings.HasPrefix(o, prefix) {
				s.suggestions = append(s.suggestions, o)
				if len(s.suggestions) == 5 {
					break
				}
			}
		}
	}
	return cmd
}

// SuggestionsView renders the suggestion list for the field.
func (s *suggestField) SuggestionsView() string {
	if !s.Focused() || len(s.suggestions) == 0 {
		return ""
	}
	items := make([]string, len(s.suggestions))
	for i, sug := range s.suggestions {
		st := lipgloss.NewStyle().Foreground(ui.ColBlue)
		if i == s.sel {
			st = st.Foreground(ui.ColPink)
		}
		items[i] = st.Render(sug)
	}
	return strings.Join(items, " ")
}

// WantsKey reports whether the field wants to handle the key itself to cycle
// suggestions instead of moving focus.
func (s *suggestField) WantsKey(k tea.KeyMsg) bool {
	switch k.String() {
	case "tab", "shift+tab", "up", "down":
		return len(s.suggestions) > 0
	default:
		return false
	}
}

// selectField allows choosing from a fixed list of options.
type selectField struct {
	options  []string
	index    int
	focused  bool
	readOnly bool
}

func newSelectField(val string, opts []string) *selectField {
	idx := 0
	for i, o := range opts {
		if o == val {
			idx = i
			break
		}
	}
	return &selectField{options: opts, index: idx}
}

func (s *selectField) Focus() {
	if !s.readOnly {
		s.focused = true
	}
}

func (s *selectField) Blur() { s.focused = false }

func (s *selectField) setReadOnly(ro bool) {
	s.readOnly = ro
	if ro {
		s.Blur()
	}
}

func (s *selectField) Update(msg tea.Msg) tea.Cmd {
	if !s.focused || s.readOnly {
		return nil
	}
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "left", "h":
			s.index--
		case "right", "l", " ":
			s.index++
		}
		if s.index < 0 {
			s.index = len(s.options) - 1
		}
		if s.index >= len(s.options) {
			s.index = 0
		}
	}
	return nil
}

func (s *selectField) View() string {
	val := s.options[s.index]
	if s.focused {
		return ui.FocusedStyle.Render(val)
	}
	return val
}

func (s *selectField) Value() string { return s.options[s.index] }

// checkField is a boolean toggle input.
type checkField struct {
	value    bool
	focused  bool
	readOnly bool
}

func newCheckField(val bool) *checkField { return &checkField{value: val} }

func (c *checkField) Focus()              { c.focused = true }
func (c *checkField) Blur()               { c.focused = false }
func (c *checkField) setReadOnly(ro bool) { c.readOnly = ro }

func (c *checkField) Update(msg tea.Msg) tea.Cmd {
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

func (c *checkField) View() string {
	box := "[ ]"
	if c.value {
		box = "[x]"
	}
	if c.focused {
		return ui.FocusedStyle.Render(box)
	}
	return box
}

func (c *checkField) Value() string { return fmt.Sprintf("%v", c.value) }
