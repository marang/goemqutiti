package emqutiti

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/marang/emqutiti/constants"
)

// Test that Ctrl+L opens the log view and Esc returns.
func TestLogViewToggle(t *testing.T) {
	m, _ := initialModel(nil)
	m.handleModeSwitchKey(tea.KeyMsg{Type: tea.KeyCtrlL})
	if m.CurrentMode() != constants.ModeLogs {
		t.Fatalf("expected log mode, got %v", m.CurrentMode())
	}
	m.logs.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.CurrentMode() != constants.ModeClient {
		t.Fatalf("expected to return to client mode, got %v", m.CurrentMode())
	}
}
