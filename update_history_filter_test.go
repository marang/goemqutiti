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

// Test that filtered results persist after the next update cycle.
func TestHistoryFilterPersists(t *testing.T) {
	m := initialModel(nil)
	hs := &HistoryStore{}
	m.history.store = hs
	ts := time.Now()
	hs.Add(Message{Timestamp: ts, Topic: "foo", Payload: "hello", Kind: "pub"})
	hs.Add(Message{Timestamp: ts, Topic: "bar", Payload: "bye", Kind: "pub"})

	m.startHistoryFilter()
	m.history.filterForm.topic.SetValue("foo")
	m.history.filterForm.start.SetValue("")
	m.history.filterForm.end.SetValue("")
	mv, _ := m.updateHistoryFilter(tea.KeyMsg{Type: tea.KeyEnter})
	m = &mv

	// simulate a subsequent update
	m.updateClient(nil)

	items := m.history.list.Items()
	if len(items) != 1 {
		t.Fatalf("expected 1 item after update, got %d", len(items))
	}
	hi := items[0].(historyItem)
	if hi.topic != "foo" {
		t.Fatalf("unexpected topic %q", hi.topic)
	}
}
