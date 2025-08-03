package topics

import (
	tea "github.com/charmbracelet/bubbletea"
	"testing"
)

func TestHandleClickToggles(t *testing.T) {
	c := newTestComponent()
	c.Items = []Item{{Name: "a", Subscribed: true}}
	c.ChipBounds = []ChipBound{{XPos: 0, YPos: 0, Width: 5, Height: 1}}
	msg := tea.MouseMsg{Type: tea.MouseLeft, X: 0, Y: 0}
	c.HandleClick(msg, 0)
	if c.Items[0].Subscribed {
		t.Fatalf("expected toggle on click")
	}
}
