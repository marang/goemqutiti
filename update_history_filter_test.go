package main

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Test that applying history filters populates the list with results.
func TestUpdateHistoryFilter(t *testing.T) {
	m := initialModel(nil)
	hs := &HistoryStore{}
	m.history.store = hs
	ts := time.Now()
	hs.Add(Message{Timestamp: ts, Topic: "foo", Payload: "hello", Kind: "pub"})

	m.startHistoryFilter()
	m.history.filterForm.topic.SetValue("foo")
	m.history.filterForm.start.SetValue("")
	m.history.filterForm.end.SetValue("")
	mv, _ := m.updateHistoryFilter(tea.KeyMsg{Type: tea.KeyEnter})
	m = &mv

	items := m.history.list.Items()
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}
