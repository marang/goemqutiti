package emqutiti

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/emqutiti/history"
)

// Test that applying history filters populates the list with results.
func TestUpdateHistoryFilter(t *testing.T) {
	m, _ := initialModel(nil)
	hs := &historyStore{}
	m.history.store = hs
	ts := time.Now()
	hs.Append(history.Message{Timestamp: ts, Topic: "foo", Payload: "hello", Kind: "pub"})

	m.startHistoryFilter()
	m.history.filterForm.topic.SetValue("foo")
	m.history.filterForm.start.SetValue("")
	m.history.filterForm.end.SetValue("")
	m.history.UpdateFilter(tea.KeyMsg{Type: tea.KeyEnter})

	if len(m.history.list.Items()) != 1 {
		t.Fatalf("expected 1 item, got %d", len(m.history.list.Items()))
	}
	if len(m.history.items) != 1 {
		t.Fatalf("expected history.items to contain 1 item, got %d", len(m.history.items))
	}
}

// Test that filtered results persist after the next update cycle.
func TestHistoryFilterPersists(t *testing.T) {
	m, _ := initialModel(nil)
	hs := &historyStore{}
	m.history.store = hs
	ts := time.Now()
	hs.Append(history.Message{Timestamp: ts, Topic: "foo", Payload: "hello", Kind: "pub"})
	hs.Append(history.Message{Timestamp: ts, Topic: "bar", Payload: "bye", Kind: "pub"})

	m.startHistoryFilter()
	m.history.filterForm.topic.SetValue("foo")
	m.history.filterForm.start.SetValue("")
	m.history.filterForm.end.SetValue("")
	m.history.UpdateFilter(tea.KeyMsg{Type: tea.KeyEnter})

	// simulate a subsequent update
	m.updateClient(nil)

	items := m.history.list.Items()
	if len(items) != 1 {
		t.Fatalf("expected 1 item after update, got %d", len(items))
	}
	if len(m.history.items) != 1 {
		t.Fatalf("history.items length = %d, want 1", len(m.history.items))
	}
	hi := items[0].(history.Item)
	if hi.Topic != "foo" {
		t.Fatalf("unexpected topic %q", hi.Topic)
	}
}

// Test that filtering updates the history label counts.
func TestHistoryFilterUpdatesCounts(t *testing.T) {
	m, _ := initialModel(nil)
	hs := &historyStore{}
	m.history.store = hs
	ts := time.Now()
	hs.Append(history.Message{Timestamp: ts, Topic: "foo", Payload: "hello", Kind: "pub"})
	hs.Append(history.Message{Timestamp: ts, Topic: "bar", Payload: "bye", Kind: "pub"})

	m.startHistoryFilter()
	m.history.filterForm.topic.SetValue("foo")
	m.history.filterForm.start.SetValue("")
	m.history.filterForm.end.SetValue("")
	m.history.UpdateFilter(tea.KeyMsg{Type: tea.KeyEnter})
	m.Update(tea.WindowSizeMsg{Width: 40, Height: 20})

	view := m.viewClient()
	if !strings.Contains(view, "History (1/2 messages") {
		t.Fatalf("expected count in label, got %q", view)
	}
}
