package emqutiti

import "testing"

// Test that SetFocus correctly focuses the message input
func TestSetFocusMessage(t *testing.T) {
	m, _ := initialModel(nil)
	if m.message.Input().Focused() {
		t.Fatalf("message input should start blurred")
	}
	m.SetFocus(idMessage)
	if !m.message.Input().Focused() {
		t.Fatalf("message input not focused after SetFocus")
	}
	if m.focus.Index() != 2 {
		t.Fatalf("focusIndex expected 2, got %d", m.focus.Index())
	}
}

// Test that pressing Tab cycles focus from topic input to topics list
func TestTabCyclesFromTopicInput(t *testing.T) {
	m, _ := initialModel(nil)
	if m.focus.Index() != 1 {
		t.Fatalf("initial focus index should be 1")
	}
	m.SetFocus(idTopic)
	if !m.topics.Input.Focused() {
		t.Fatalf("topic input should be focused after SetFocus")
	}
	m.focus.Next()
	if m.topics.Input.Focused() {
		t.Fatalf("topic input should be blurred after tab")
	}
	if m.focus.Index() != 1 {
		t.Fatalf("focus index should be 1 after tab, got %d", m.focus.Index())
	}
}
