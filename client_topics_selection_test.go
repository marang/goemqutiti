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
	m.focus.Set(4) // idHelp index so tab wraps to idTopics
	m.ui.focusIndex = 4
	m.handleTabKey()
	if m.topics.Selected() != 1 {
		t.Fatalf("expected selected index 1 after Tab, got %d", m.topics.Selected())
	}

	// Simulate focus cycling backward to topics
	m.focus.Set(1) // idTopic index so shift+tab goes to idTopics
	m.ui.focusIndex = 1
	m.handleShiftTabKey()
	if m.topics.Selected() != 1 {
		t.Fatalf("expected selected index 1 after Shift+Tab, got %d", m.topics.Selected())
	}
}
