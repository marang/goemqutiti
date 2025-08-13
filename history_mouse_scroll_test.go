package emqutiti

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHistoryMouseWheelScroll(t *testing.T) {
	m, _ := initialModel(nil)
	m.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	for i := 0; i < 50; i++ {
		m.history.Append("t", "msg", "pub", false, "")
	}
	m.viewClient()
	m.history.List().Select(0)
	m.SetFocus(idHistory)

	before := m.history.List().Index()
	m.Update(tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonWheelDown})
	after := m.history.List().Index()
	if after == before {
		t.Fatalf("expected history to scroll on mouse wheel")
	}
}
