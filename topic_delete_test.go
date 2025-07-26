package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Test that deleting a topic via confirmation removes it from the list
func TestDeleteTopic(t *testing.T) {
	m := initialModel(nil)
	m.topics = []topicItem{{title: "a", active: true}, {title: "b", active: false}}
	m.setFocus("topics")
	m.selectedTopic = 0
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if cmd == nil || m.mode != modeConfirmDelete {
		t.Fatalf("expected confirm delete mode")
	}
	if m.confirmAction == nil {
		t.Fatalf("confirm action not set")
	}
	if len(m.topics) != 2 {
		t.Fatalf("unexpected topics before confirm: %#v", m.topics)
	}
	t.Logf("before confirm: %#v", m.topics)
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	t.Logf("after confirm: %#v", m.topics)
	if len(m.topics) != 1 || m.topics[0].title != "b" {
		t.Fatalf("topic not removed: %#v", m.topics)
	}
}
