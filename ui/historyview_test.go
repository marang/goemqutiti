package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// Test that HistoryView lines fit the configured width.
func TestHistoryViewWidth(t *testing.T) {
	hv := NewHistoryView(20, 5)
	hv.SetLines([]string{"one two three"})
	out := hv.View()
	lines := strings.Split(out, "\n")
	for i, l := range lines {
		if lipgloss.Width(l) != 16 {
			t.Fatalf("line %d width=%d want=16", i, lipgloss.Width(l))
		}
	}
}

// Ensure the box renders with aligned borders when wrapped in LegendBox.
func TestHistoryBoxLayout(t *testing.T) {
	hv := NewHistoryView(20, 5)
	hv.SetLines([]string{"foo"})
	box := LegendBox(hv.View(), "Hist", 20, 0, ColGreen, false)
	lines := strings.Split(box, "\n")
	width := lipgloss.Width(lines[0])
	for i, l := range lines {
		if lipgloss.Width(l) != width {
			t.Fatalf("line %d width=%d want=%d", i, lipgloss.Width(l), width)
		}
	}
}
