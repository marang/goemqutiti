package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Test that historyDelegate renders lines that fit the list width
func TestHistoryDelegateWidth(t *testing.T) {
	m := initialModel(nil)
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

// Test that the history box has aligned borders when rendered
func TestHistoryBoxLayout(t *testing.T) {
	m := initialModel(nil)
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
	m := initialModel(nil)
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
	m := initialModel(nil)
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
	items := make([]list.Item, len(msgs))
	m.history.items = make([]historyItem, len(msgs))
	for i, mm := range msgs {
		hi := historyItem{timestamp: mm.Timestamp, topic: mm.Topic, payload: mm.Payload, kind: mm.Kind, archived: mm.Archived}
		items[i] = hi
		m.history.items[i] = hi
	}
	m.history.list.SetItems(items)
	view = m.viewClient()
	if !strings.Contains(view, "History (1/2 messages") {
		t.Fatalf("expected filtered count in label, got %q", view)
	}
}
