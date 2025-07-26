package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Test that pressing enter in the topic input subscribes to that topic
func TestEnterAddsTopic(t *testing.T) {
	m := initialModel(nil)
	m.topicInput.SetValue("foo")
	m.setFocus("topic")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatalf("expected command on enter")
	}
	if len(m.topics) != 1 || m.topics[0].title != "foo" || !m.topics[0].active {
		t.Fatalf("topic not added: %#v", m.topics)
	}
}

// Test that ctrl+s publishes the message in the editor
func TestCtrlSPublishesMessage(t *testing.T) {
	m := initialModel(nil)
	m.topics = []topicItem{{title: "foo", active: true}}
	m.messageInput.SetValue("hello")
	m.setFocus("message")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if cmd == nil {
		t.Fatalf("expected command on ctrl+s")
	}
	if len(m.payloads) != 1 || m.payloads[0].payload != "hello" {
		t.Fatalf("payload not stored: %#v", m.payloads)
	}
	if len(m.history.Items()) != 1 {
		t.Fatalf("history entry not added")
	}
}
