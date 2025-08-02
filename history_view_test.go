package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dgraph-io/badger/v4"
)

// Test that historyDelegate renders lines that fit the list width
func TestHistoryDelegateWidth(t *testing.T) {
	m, _ := initialModel(nil)
	d := historyDelegate{m: m}
	m.history.list.SetSize(30, 4)
	hi := historyItem{timestamp: time.Now(), topic: "foo", payload: "bar", kind: "pub"}
	var buf bytes.Buffer
	d.Render(&buf, m.history.list, 0, hi)
	lines := strings.Split(buf.String(), "\n")
	for i, line := range lines {
		if lipgloss.Width(line) != 30 {
			t.Fatalf("line %d width=%d want=30", i, lipgloss.Width(line))
		}
	}
}

// Test that short multiline payloads are shown without truncation.
func TestHistoryPreviewShortMultiline(t *testing.T) {
	m, _ := initialModel(nil)
	d := historyDelegate{m: m}
	m.history.list.SetSize(80, 4)
	hi := historyItem{timestamp: time.Now(), topic: "foo", payload: "one\ntwo", kind: "pub"}
	var buf bytes.Buffer
	d.Render(&buf, m.history.list, 0, hi)
	out := buf.String()
	if strings.Contains(out, "\u2026") {
		t.Fatalf("expected no ellipsis, got %q", out)
	}
	if !strings.Contains(out, "one two") {
		t.Fatalf("expected joined payload, got %q", out)
	}
}

// Test that the history box has aligned borders when rendered
func TestHistoryBoxLayout(t *testing.T) {
	m, _ := initialModel(nil)
	m.Update(tea.WindowSizeMsg{Width: 40, Height: 30})
	m.appendHistory("foo", "bar", "pub", "")
	view := m.viewClient()
	lines := strings.Split(view, "\n")
	var hist []string
	collecting := false
	for _, l := range lines {
		if strings.Contains(l, "History") {
			collecting = true
		}
		if collecting {
			hist = append(hist, l)
			if strings.Contains(l, "\u2518") { // bottom right corner '┘'
				break
			}
		}
	}
	if len(hist) == 0 {
		t.Fatalf("history box not found in view")
	}
	width := lipgloss.Width(hist[0])
	for i, l := range hist {
		if lipgloss.Width(l) != width {
			t.Fatalf("history line %d width=%d want=%d", i, lipgloss.Width(l), width)
		}
	}
}

// Test that active filters are shown inside the history box rather than in the border.
func TestHistoryFilterDisplayedInsideBox(t *testing.T) {
	m, _ := initialModel(nil)
	m.history.filterQuery = "topic=foo"
	m.history.store = &HistoryStore{}
	m.appendHistory("foo", "bar", "pub", "")
	m.Update(tea.WindowSizeMsg{Width: 40, Height: 30})
	view := m.viewClient()
	lines := strings.Split(view, "\n")
	var hist []string
	collecting := false
	for _, l := range lines {
		if strings.Contains(l, "History") {
			collecting = true
		}
		if collecting {
			hist = append(hist, l)
			if strings.Contains(l, "\u2518") {
				break
			}
		}
	}
	if len(hist) < 2 {
		t.Fatalf("history box not found in view")
	}
	if strings.Contains(hist[0], "topic=foo") {
		t.Fatalf("filter query should not appear in border")
	}
	found := false
	for _, l := range hist[1:] {
		if strings.Contains(l, "topic=foo") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("filter query not found inside history box")
	}
}

// Test that the history label reports total and filtered message counts.
func TestHistoryLabelCounts(t *testing.T) {
	m, _ := initialModel(nil)
	m.history.store = &HistoryStore{}
	m.appendHistory("foo", "bar", "pub", "")
	m.appendHistory("bar", "baz", "sub", "")
	m.Update(tea.WindowSizeMsg{Width: 40, Height: 30})
	view := m.viewClient()
	if !strings.Contains(view, "History (2 messages") {
		t.Fatalf("expected total count in label, got %q", view)
	}

	// apply filter showing only "foo"
	m.history.filterQuery = "topic=foo"
	msgs := m.history.store.Search([]string{"foo"}, time.Time{}, time.Time{}, "")
	var items []list.Item
	m.history.items, items = messagesToHistoryItems(msgs)
	m.history.list.SetItems(items)
	view = m.viewClient()
	if !strings.Contains(view, "History (1/2 messages") {
		t.Fatalf("expected filtered count in label, got %q", view)
	}
}

// Test that a long filter query does not break the history box layout.
func TestHistoryFilterLineWidth(t *testing.T) {
	m, _ := initialModel(nil)
	long := strings.Repeat("x", 100)
	m.history.filterQuery = "topic=" + long
	m.history.store = &HistoryStore{}
	m.appendHistory("foo", "bar", "pub", "")
	m.Update(tea.WindowSizeMsg{Width: 40, Height: 30})
	view := m.viewClient()
	lines := strings.Split(view, "\n")
	var hist []string
	collecting := false
	for _, l := range lines {
		if strings.Contains(l, "History") {
			collecting = true
		}
		if collecting {
			hist = append(hist, l)
			if strings.Contains(l, "\u2518") {
				break
			}
		}
	}
	if len(hist) == 0 {
		t.Fatalf("history box not found in view")
	}
	width := lipgloss.Width(hist[0])
	for i, l := range hist {
		if lipgloss.Width(l) != width {
			t.Fatalf("history line %d width=%d want=%d", i, lipgloss.Width(l), width)
		}
	}
}

// Test that the history help legend remains visible after applying a filter.
func TestHistoryHelpVisibleWithFilter(t *testing.T) {
	m, _ := initialModel(nil)
	m.history.store = &HistoryStore{}
	m.history.filterQuery = "topic=foo"
	m.appendHistory("foo", "bar", "pub", "")
	m.Update(tea.WindowSizeMsg{Width: 40, Height: 30})
	view := m.viewClient()
	if !strings.Contains(view, "\u2191/k up") {
		t.Fatalf("expected history shortcuts in view, got %q", view)
	}
}

// Test that an error during archiving is logged and reported to the user.
func TestArchiveErrorFeedback(t *testing.T) {
	m, _ := initialModel(nil)
	m.ui.focusIndex = 3 // focus history
	hi := historyItem{timestamp: time.Now(), topic: "foo", payload: "bar", kind: "pub"}
	m.history.items = []historyItem{hi}
	m.history.list.SetItems([]list.Item{hi})
	m.history.store = &HistoryStore{}
	m.handleArchiveKey()
	if len(m.history.items) != 2 {
		t.Fatalf("expected original item plus error log, got %d items", len(m.history.items))
	}
	last := m.history.items[1]
	if last.kind != "log" || !strings.Contains(last.payload, "Failed to archive") {
		t.Fatalf("expected archive error log, got %#v", last)
	}
}

// Test that an error during deletion is logged and reported to the user.
func TestDeleteErrorFeedback(t *testing.T) {
	m, _ := initialModel(nil)
	hi := historyItem{timestamp: time.Now(), topic: "foo", payload: "bar", kind: "pub"}
	m.history.items = []historyItem{hi}
	m.history.list.SetItems([]list.Item{hi})
	m.history.list.Select(0)
	db, _ := badger.Open(badger.DefaultOptions("").WithInMemory(true))
	_ = db.Close()
	m.history.store = &HistoryStore{db: db}
	m.handleDeleteHistoryKey()
	if m.confirmAction == nil {
		t.Fatalf("confirmAction not set")
	}
	m.confirmAction()
	if len(m.history.items) != 2 {
		t.Fatalf("expected original item plus error log, got %d items", len(m.history.items))
	}
	last := m.history.items[1]
	if last.kind != "log" || !strings.Contains(last.payload, "Failed to delete") {
		t.Fatalf("expected delete error log, got %#v", last)
 	}
}

// Test that triggering the detail view shows the complete payload.
func TestHistoryDetailViewShowsPayload(t *testing.T) {
	m, _ := initialModel(nil)
	long := strings.Repeat("x", historyPreviewLimit+10)
	m.appendHistory("foo", long, "pub", "")
	m.setFocus(idHistory)
	m.handleEnterKey()
	if m.currentMode() != modeHistoryDetail {
		t.Fatalf("expected detail mode, got %v", m.currentMode())
	}
	if m.history.detailItem.payload != long {
		t.Fatalf("payload not preserved in detail view")
	}
}
