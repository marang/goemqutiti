package emqutiti

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/emqutiti/constants"
)

// Test that clicking the help icon opens the help view.
func TestHelpIconClick(t *testing.T) {
	m, _ := initialModel(nil)
	m.ui.width = 80
	msg := tea.MouseMsg{X: 79, Y: 0, Type: tea.MouseLeft, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}
	m.Update(msg)
	if m.CurrentMode() != constants.ModeHelp {
		t.Fatalf("expected ModeHelp, got %v", m.CurrentMode())
	}
}
