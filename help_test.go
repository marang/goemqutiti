package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Test that pressing enter with the help icon focused opens the help view.
func TestEnterOpensHelp(t *testing.T) {
	m := initialModel(nil)
	m.setFocus("help")
	if m.ui.focusOrder[m.ui.focusIndex] != "help" {
		t.Fatalf("help not focused")
	}
	if m.ui.mode != modeClient {
		t.Fatalf("initial mode not client: %v", m.ui.mode)
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatalf("expected nil command")
	}
	if m.ui.mode != modeHelp {
		t.Fatalf("expected help mode, got %v", m.ui.mode)
	}
}

// Test that space also opens help when focused.
func TestSpaceOpensHelp(t *testing.T) {
	m := initialModel(nil)
	m.setFocus("help")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if cmd != nil {
		t.Fatalf("expected nil command")
	}
	if m.ui.mode != modeHelp {
		t.Fatalf("expected help mode, got %v", m.ui.mode)
	}
}
