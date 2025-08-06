package emqutiti

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/emqutiti/constants"
)

// Test that clicking the help icon opens the help view.
func TestHelpIconClick(t *testing.T) {
	m, _ := initialModel(nil)
	m.ui.width = 40
	msg := tea.MouseMsg{X: 39, Y: 0, Type: tea.MouseLeft, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}
	m.Update(msg)
	if m.CurrentMode() != constants.ModeHelp {
		t.Fatalf("expected ModeHelp, got %v", m.CurrentMode())
	}
}

// Test that the help icon stays on the first line when space is limited.
func TestHelpIconSticky(t *testing.T) {
	m, _ := initialModel(nil)
	m.ui.width = 20
	view := m.overlayHelp("")
	lines := strings.Split(view, "\n")
	if !strings.HasSuffix(lines[0], "?") {
		t.Fatalf("help icon moved off first line: %q", view)
	}
}
