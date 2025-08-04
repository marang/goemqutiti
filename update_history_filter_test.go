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
	m.history.SetStore(hs)
	ts := time.Now()
	hs.Append(history.Message{Timestamp: ts, Topic: "foo", Payload: "hello", Kind: "pub"})

	m.startHistoryFilter()
	m.history.FilterForm().Topic().SetValue("foo")
	m.history.FilterForm().Start().SetValue("")
	m.history.FilterForm().End().SetValue("")
	m.history.UpdateFilter(tea.KeyMsg{Type: tea.KeyEnter})

	if len(m.history.List().Items()) != 1 {
		t.Fatalf("expected 1 item, got %d", len(m.history.List().Items()))
	}
	if len(m.history.Items()) != 1 {
		t.Fatalf("expected history.items to contain 1 item, got %d", len(m.history.Items()))
	}
}

// Test that filtered results persist after the next update cycle.
func TestHistoryFilterPersists(t *testing.T) {
	m, _ := initialModel(nil)
	hs := &historyStore{}
	m.history.SetStore(hs)
	ts := time.Now()
	hs.Append(history.Message{Timestamp: ts, Topic: "foo", Payload: "hello", Kind: "pub"})
	hs.Append(history.Message{Timestamp: ts, Topic: "bar", Payload: "bye", Kind: "pub"})

	m.startHistoryFilter()
	m.history.FilterForm().Topic().SetValue("foo")
	m.history.FilterForm().Start().SetValue("")
	m.history.FilterForm().End().SetValue("")
	m.history.UpdateFilter(tea.KeyMsg{Type: tea.KeyEnter})

	// simulate a subsequent update
	m.updateClient(nil)

	items := m.history.List().Items()
	if len(items) != 1 {
		t.Fatalf("expected 1 item after update, got %d", len(items))
	}
	if len(m.history.Items()) != 1 {
		t.Fatalf("history.items length = %d, want 1", len(m.history.Items()))
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
	m.history.SetStore(hs)
	ts := time.Now()
	hs.Append(history.Message{Timestamp: ts, Topic: "foo", Payload: "hello", Kind: "pub"})
	hs.Append(history.Message{Timestamp: ts, Topic: "bar", Payload: "bye", Kind: "pub"})

	m.startHistoryFilter()
	m.history.FilterForm().Topic().SetValue("foo")
	m.history.FilterForm().Start().SetValue("")
	m.history.FilterForm().End().SetValue("")
	m.history.UpdateFilter(tea.KeyMsg{Type: tea.KeyEnter})
	m.Update(tea.WindowSizeMsg{Width: 40, Height: 20})

	view := m.viewClient()
	if !strings.Contains(view, "History (1/2 messages") {
		t.Fatalf("expected count in label, got %q", view)
	}
}

// Test that the archived checkbox filters messages accordingly.
func TestHistoryFilterArchived(t *testing.T) {
	m, _ := initialModel(nil)
	hs := &historyStore{}
	m.history.SetStore(hs)
	ts := time.Now()
	hs.Append(history.Message{Timestamp: ts, Topic: "foo", Payload: "hello", Kind: "pub"})
	hs.Append(history.Message{Timestamp: ts, Topic: "bar", Payload: "bye", Kind: "pub", Archived: true})

	// default unchecked state shows unarchived messages
	m.startHistoryFilter()
	m.history.FilterForm().Start().SetValue("")
	m.history.FilterForm().End().SetValue("")
	m.history.UpdateFilter(tea.KeyMsg{Type: tea.KeyEnter})
	if len(m.history.Items()) != 1 || m.history.Items()[0].Archived {
		t.Fatalf("expected one active message, got %v", m.history.Items())
	}

	// enabling checkbox returns archived messages
	m.startHistoryFilter()
	m.history.FilterForm().Start().SetValue("")
	m.history.FilterForm().End().SetValue("")
	m.history.FilterForm().Archived().SetBool(true)
	m.history.UpdateFilter(tea.KeyMsg{Type: tea.KeyEnter})
	if len(m.history.Items()) != 1 || !m.history.Items()[0].Archived {
		t.Fatalf("expected one archived message, got %v", m.history.Items())
	}
}
