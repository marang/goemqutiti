package emqutiti

import (
	"testing"

	"github.com/marang/emqutiti/constants"
)

// Test that leaving the help screen restores the previous mode and
// does not leave help in the navigation history.
func TestHelpModeRemovedFromStack(t *testing.T) {
	m, _ := initialModel(nil)
	m.SetMode(constants.ModeConnections)
	m.SetMode(constants.ModeHelp)
	m.SetPreviousMode()
	if m.CurrentMode() != constants.ModeConnections {
		t.Fatalf("expected to return to connections mode, got %v", m.CurrentMode())
	}
	if m.PreviousMode() == constants.ModeHelp {
		t.Fatalf("help mode should not remain in the stack")
	}
}
