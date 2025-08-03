package emqutiti

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Test that pressing enter in the topic input subscribes to that topic
func TestEnterAddsTopic(t *testing.T) {
	m, _ := initialModel(nil)
	m.topics.input.SetValue("foo")
	m.setFocus(idTopic)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatalf("expected command on enter")
	}
	if len(m.topics.items) != 1 || m.topics.items[0].title != "foo" || !m.topics.items[0].subscribed {
		t.Fatalf("topic not added: %#v", m.topics.items)
	}
}

// Test that ctrl+s publishes the message in the editor
func TestCtrlSPublishesMessage(t *testing.T) {
	m, _ := initialModel(nil)
	m.topics.items = []topicItem{{title: "foo", subscribed: true}}
	m.message.input.SetValue("hello")
	m.setFocus(idMessage)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if cmd == nil {
		t.Fatalf("expected command on ctrl+s")
	}
	items := m.payloads.Items()
	if len(items) != 1 || items[0].payload != "hello" {
		t.Fatalf("payload not stored: %#v", items)
	}
	if len(m.history.List().Items()) != 1 {
		t.Fatalf("history entry not added")
	}
}
