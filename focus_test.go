package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Test that setFocus correctly focuses the message input
func TestSetFocusMessage(t *testing.T) {
	m := initialModel(nil)
	if m.messageInput.Focused() {
		t.Fatalf("message input should start blurred")
	}
	cmd := m.setFocus("message")
	if !m.messageInput.Focused() {
		t.Fatalf("message input not focused after setFocus")
	}
	if m.focusIndex != 2 {
		t.Fatalf("focusIndex expected 2, got %d", m.focusIndex)
	}
	if cmd == nil {
		t.Fatalf("expected non-nil command from setFocus")
	}
}

// Test that pressing Tab cycles focus from topic to message
// Test that pressing Tab cycles focus from topics to topic input
func TestTabCyclesToTopic(t *testing.T) {
	m := initialModel(nil)
	if m.focusIndex != 0 {
		t.Fatalf("initial focus index should be 0")
	}
	msg := tea.KeyMsg{Type: tea.KeyTab}
	_, cmd := m.Update(msg)
	if !m.topicInput.Focused() {
		t.Fatalf("topic input should be focused after tab")
	}
	if m.focusIndex != 1 {
		t.Fatalf("focus index should be 1 after tab, got %d", m.focusIndex)
	}
	if cmd == nil {
		t.Fatalf("update should return a command")
	}
}
