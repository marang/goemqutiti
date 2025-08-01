package main

import (
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// Test that ctrl+a toggles history selection
func TestCtrlATogglesHistorySelection(t *testing.T) {
	m, _ := initialModel(nil)
	m.history.items = []historyItem{
		{timestamp: time.Now(), topic: "t1", payload: "p1", kind: "pub"},
		{timestamp: time.Now(), topic: "t2", payload: "p2", kind: "pub"},
	}
	items := make([]list.Item, len(m.history.items))
	for i, it := range m.history.items {
		items[i] = it
	}
	m.history.list.SetItems(items)
	m.setFocus(idHistory)

	// First ctrl+a selects all
	m.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
	for i, it := range m.history.items {
		if it.isSelected == nil || !*it.isSelected {
			t.Fatalf("item %d not selected", i)
		}
	}
	if m.history.selectionAnchor != 0 {
		t.Fatalf("selectionAnchor = %d, want 0", m.history.selectionAnchor)
	}

	// Second ctrl+a deselects all
	m.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
	for i, it := range m.history.items {
		if it.isSelected != nil {
			t.Fatalf("item %d still selected", i)
		}
	}
	if m.history.selectionAnchor != -1 {
		t.Fatalf("selectionAnchor = %d, want -1", m.history.selectionAnchor)
	}
}
