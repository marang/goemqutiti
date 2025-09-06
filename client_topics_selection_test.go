package emqutiti

import (
	"testing"

	"github.com/marang/emqutiti/topics"
)

func TestTopicSelectionPersistsAcrossFocus(t *testing.T) {
	m, _ := initialModel(nil)
	m.topics.Items = []topics.Item{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	m.topics.SetSelected(1)

	// Simulate focus cycling forward to topics
	m.focus.Set(4) // idHelp index so tab wraps to idTopic then idTopics
	m.ui.focusIndex = 4
	m.handleTabKey()
	m.handleTabKey()
	if m.topics.Selected() != 1 {
		t.Fatalf("expected selected index 1 after Tab, got %d", m.topics.Selected())
	}

	// Simulate focus cycling backward to topics
	m.focus.Set(2) // idMessage index so shift+tab goes to idTopics
	m.ui.focusIndex = 2
	m.handleShiftTabKey()
	if m.topics.Selected() != 1 {
		t.Fatalf("expected selected index 1 after Shift+Tab, got %d", m.topics.Selected())
	}
}

func TestTopicInputInitiallyBlurred(t *testing.T) {
	m, _ := initialModel(nil)
	if m.topics.Input.Focused() {
		t.Fatalf("topic input should not be focused on init")
	}
}

func TestToggleTopicKeepsSelection(t *testing.T) {
	m, _ := initialModel(nil)
	m.topics.Items = []topics.Item{
		{Name: "a", Subscribed: true},
		{Name: "b", Subscribed: true},
		{Name: "c", Subscribed: true},
	}
	m.topics.SetSelected(2)
	m.focus.Set(1)
	m.ui.focusIndex = 1
	m.handleEnterKey()
	if m.topics.Items[m.topics.Selected()].Name != "c" {
		t.Fatalf("expected to stay on topic 'c', got %q", m.topics.Items[m.topics.Selected()].Name)
	}
}

func TestTogglePublishKeepsSelection(t *testing.T) {
	m, _ := initialModel(nil)
	m.topics.Items = []topics.Item{
		{Name: "a", Subscribed: true},
		{Name: "b", Subscribed: true},
		{Name: "c", Subscribed: true},
	}
	m.topics.SetSelected(2)
	m.focus.Set(1)
	m.ui.focusIndex = 1
	m.handleTogglePublishKey()
	if m.topics.Items[m.topics.Selected()].Name != "c" {
		t.Fatalf("expected to stay on topic 'c', got %q", m.topics.Items[m.topics.Selected()].Name)
	}
}
