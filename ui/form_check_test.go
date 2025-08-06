package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestCheckFieldToggleKeyboard(t *testing.T) {
	cf := NewCheckField(false)
	cf.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if !cf.Bool() {
		t.Fatalf("expected true after space")
	}
	cf.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if cf.Bool() {
		t.Fatalf("expected false after second space")
	}
}

func TestCheckFieldToggleMouse(t *testing.T) {
	cf := NewCheckField(false)
	cf.Update(tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	if !cf.Bool() {
		t.Fatalf("expected true after click")
	}
}

func TestCheckFieldReadOnly(t *testing.T) {
	cf := NewCheckField(false)
	cf.SetReadOnly(true)

	cf.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	cf.Update(tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	if cf.Bool() {
		t.Fatalf("read-only field toggled")
	}
}
