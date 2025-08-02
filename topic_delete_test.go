package emqutiti

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Test that deleting a topic via confirmation removes it from the list
func TestDeleteTopic(t *testing.T) {
	m, _ := initialModel(nil)
	m.topics.items = []topicItem{{title: "a", subscribed: true}, {title: "b", subscribed: false}}
	m.setFocus(idTopics)
	m.topics.selected = 0
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyDelete})
	if cmd == nil || m.currentMode() != modeConfirmDelete {
		t.Fatalf("expected confirm delete mode")
	}
	if m.confirm.action == nil {
		t.Fatalf("confirm action not set")
	}
	if len(m.topics.items) != 2 {
		t.Fatalf("unexpected topics before confirm: %#v", m.topics.items)
	}
	t.Logf("before confirm: %#v", m.topics.items)
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	t.Logf("after confirm: %#v", m.topics.items)
	if len(m.topics.items) != 1 || m.topics.items[0].title != "b" {
		t.Fatalf("topic not removed: %#v", m.topics.items)
	}
}
