package emqutiti

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/marang/emqutiti/connections"
	"github.com/marang/emqutiti/constants"
)

// TestHandleKeyNavCyclesForm ensures Tab and Shift+Tab move focus
// through fields when editing a connection.
func TestHandleKeyNavCyclesForm(t *testing.T) {
	m, _ := initialModel(nil)
	f := connections.NewForm(connections.Profile{}, -1)
	m.connections.Form = &f
	m.SetMode(constants.ModeEditConnection)

	if _, handled := m.handleKeyNav(tea.KeyMsg{Type: tea.KeyTab}); !handled {
		t.Fatalf("Tab not handled")
	}
	if m.connections.Form.Focus != 1 {
		t.Fatalf("focus=%d want=1", m.connections.Form.Focus)
	}
	if _, handled := m.handleKeyNav(tea.KeyMsg{Type: tea.KeyShiftTab}); !handled {
		t.Fatalf("Shift+Tab not handled")
	}
	if m.connections.Form.Focus != 0 {
		t.Fatalf("focus=%d want=0", m.connections.Form.Focus)
	}
}
