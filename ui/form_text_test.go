package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Test that plain j/k are consumed by TextField and do not move focus.
func TestTextFieldConsumesJK(t *testing.T) {
	tf1 := NewTextField("", "")
	tf2 := NewTextField("", "")
	f := Form{Fields: []Field{tf1, tf2}, Focus: 0}
	f.ApplyFocus()

	for _, tt := range []struct{ r rune }{{'j'}, {'k'}} {
		key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tt.r}}
		f.CycleFocus(key)
		f.ApplyFocus()
		f.Fields[f.Focus].Update(key)
		if f.Focus != 0 {
			t.Fatalf("focus moved on %q", string(tt.r))
		}
		if tf1.Value() != string(tt.r) {
			t.Fatalf("field value=%q want=%q", tf1.Value(), string(tt.r))
		}
		tf1.SetValue("")
	}
}

// Test that SuggestField also consumes j/k without losing focus.
func TestSuggestFieldConsumesJK(t *testing.T) {
	sf := NewSuggestField([]string{"foo"}, "topic")
	tf := NewTextField("", "")
	f := Form{Fields: []Field{sf, tf}, Focus: 0}
	f.ApplyFocus()

	key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	f.CycleFocus(key)
	f.ApplyFocus()
	f.Fields[f.Focus].Update(key)
	if f.Focus != 0 {
		t.Fatalf("focus moved on j")
	}
	if sf.Value() != "j" {
		t.Fatalf("field value=%q want=j", sf.Value())
	}
}
