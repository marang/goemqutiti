package history

import (
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// stubModel implements Model for testing purposes.
type stubModel struct{}

func (stubModel) SetMode(Mode) tea.Cmd        { return nil }
func (stubModel) PreviousMode() Mode          { return nil }
func (stubModel) CurrentMode() Mode           { return nil }
func (stubModel) SetFocus(string) tea.Cmd     { return nil }
func (stubModel) Width() int                  { return 0 }
func (stubModel) Height() int                 { return 0 }
func (stubModel) OverlayHelp(s string) string { return s }

// TestHandleSelection verifies range selection behaviour.
func TestHandleSelection(t *testing.T) {
	c := NewComponent(stubModel{}, nil)
	c.items = []Item{{}, {}, {}}
	items := make([]list.Item, len(c.items))
	for i, it := range c.items {
		items[i] = it
	}
	c.list.SetItems(items)

	c.HandleSelection(0, false)
	if c.items[0].IsSelected != nil {
		t.Fatalf("expected no selection without shift")
	}

	c.HandleSelection(1, true)
	c.HandleSelection(2, true)
	for i := 1; i <= 2; i++ {
		if c.items[i].IsSelected == nil || !*c.items[i].IsSelected {
			t.Fatalf("item %d not selected", i)
		}
	}
}
