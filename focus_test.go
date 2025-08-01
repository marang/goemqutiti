package main

import "testing"

// Test that setFocus correctly focuses the message input
func TestSetFocusMessage(t *testing.T) {
	m, _ := initialModel(nil)
	if m.message.input.Focused() {
		t.Fatalf("message input should start blurred")
	}
	m.setFocus(idMessage)
	if !m.message.input.Focused() {
		t.Fatalf("message input not focused after setFocus")
	}
	if m.focus.Index() != 2 {
		t.Fatalf("focusIndex expected 2, got %d", m.focus.Index())
	}
}

// Test that pressing Tab cycles focus from topic to message
// Test that pressing Tab cycles focus from topics to topic input
func TestTabCyclesToTopic(t *testing.T) {
	m, _ := initialModel(nil)
	if m.focus.Index() != 0 {
		t.Fatalf("initial focus index should be 0")
	}
	m.focus.Next()
	if !m.topics.input.Focused() {
		t.Fatalf("topic input should be focused after tab")
	}
	if m.focus.Index() != 1 {
		t.Fatalf("focus index should be 1 after tab, got %d", m.focus.Index())
	}
}
