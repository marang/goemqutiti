package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Test that pressing enter with the help icon focused opens the help view.
func TestEnterOpensHelp(t *testing.T) {
	m := initialModel(nil)
	m.setFocus(idHelp)
	if m.ui.focusOrder[m.ui.focusIndex] != idHelp {
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
	m.setFocus(idHelp)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if cmd != nil {
		t.Fatalf("expected nil command")
	}
	if m.ui.mode != modeHelp {
		t.Fatalf("expected help mode, got %v", m.ui.mode)
	}
}

// Test that Esc exits help even after pressing enter again in help mode.
func TestEscFromHelpAfterEnter(t *testing.T) {
	m := initialModel(nil)
	m.setFocus(idHelp)
	m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // open help
	if m.ui.mode != modeHelp {
		t.Fatalf("expected help mode, got %v", m.ui.mode)
	}
	// Press enter again while help is focused
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.ui.mode != modeHelp {
		t.Fatalf("still expected help mode, got %v", m.ui.mode)
	}
	// Esc should return to client mode
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.ui.mode != modeClient {
		t.Fatalf("esc should return to client mode, got %v", m.ui.mode)
	}
}
