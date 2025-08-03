package emqutiti

import (
	connections "github.com/marang/emqutiti/connections"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/emqutiti/history"
)

// Test copy behavior when history items are selected.
func TestHandleClientKeyCopySelected(t *testing.T) {
	m, _ := initialModel(nil)
	sel := true
	hi := history.Item{Timestamp: time.Now(), Topic: "t1", Payload: "msg1", Kind: "pub", IsSelected: &sel}
	m.history.SetItems([]history.Item{hi})
	m.history.List().SetItems([]list.Item{hi})
	m.history.List().Select(0)

	m.handleClientKey(tea.KeyMsg{Type: tea.KeyCtrlC})

	if len(m.history.Items()) != 2 {
		t.Fatalf("expected error appended to history, got %d items", len(m.history.Items()))
	}
	if m.history.Items()[1].Kind != "log" {
		t.Fatalf("expected last item kind 'log', got %q", m.history.Items()[1].Kind)
	}
}

// Test disconnect behavior clears connection state.
func TestHandleClientKeyDisconnect(t *testing.T) {
	conn := connections.NewConnectionsModel()
	conn.Profiles = []connections.Profile{{Name: "test"}}
	conn.Statuses["test"] = "connected"
	conn.Errors["test"] = "oops"

	m, _ := initialModel(&conn)
	m.mqttClient = &MQTTClient{}
	m.connections.Connection = "test"
	m.connections.Active = "test"
	m.connections.SetConnected("test")
	m.connections.Manager.Errors["test"] = "boom"

	m.handleClientKey(tea.KeyMsg{Type: tea.KeyCtrlX})

	if m.mqttClient != nil {
		t.Fatalf("expected mqttClient nil after disconnect")
	}
	if m.connections.Connection != "" || m.connections.Active != "" {
		t.Fatalf("expected connection cleared, got %q %q", m.connections.Connection, m.connections.Active)
	}
	if st := m.connections.Manager.Statuses["test"]; st != "disconnected" {
		t.Fatalf("expected status 'disconnected', got %q", st)
	}
	if err := m.connections.Manager.Errors["test"]; err != "" {
		t.Fatalf("expected error cleared, got %q", err)
	}
}

// Test that pressing '/' while history is focused starts the filter form.
func TestHandleClientKeyFilterInitiation(t *testing.T) {
	m, _ := initialModel(nil)
	idx := 3 // history index in focus order
	m.focus.Set(idx)
	m.ui.focusIndex = idx

	m.handleClientKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})

	if m.ui.modeStack[0] != modeHistoryFilter {
		t.Fatalf("expected modeHistoryFilter, got %v", m.ui.modeStack[0])
	}
	if m.history.FilterForm() == nil {
		t.Fatalf("expected filter form to be initialized")
	}
}
