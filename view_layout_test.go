package emqutiti

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestOverlayHelpStacksOnNarrowWidth(t *testing.T) {
	m, _ := initialModel(nil)
	m.Update(tea.WindowSizeMsg{Width: helpReflowWidth - 10, Height: 20})
	out := m.overlayHelp("status")
	lines := strings.Split(out, "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least two lines, got %d", len(lines))
	}
	if strings.Contains(lines[0], "?") {
		t.Fatalf("expected first line without help icon: %q", lines[0])
	}
	if !strings.Contains(lines[1], "?") || !strings.Contains(lines[1], "status") {
		t.Fatalf("expected help and status on second line: %q", lines[1])
	}
}

func TestOverlayHelpInlineOnWideWidth(t *testing.T) {
	m, _ := initialModel(nil)
	m.Update(tea.WindowSizeMsg{Width: helpReflowWidth + 10, Height: 20})
	out := m.overlayHelp("status")
	lines := strings.Split(out, "\n")
	if !strings.Contains(lines[0], "?") || !strings.Contains(lines[0], "Switch views") {
		t.Fatalf("expected help and info on first line: %q", lines[0])
	}
	if len(lines) < 2 || strings.Contains(lines[1], "?") || !strings.Contains(lines[1], "status") {
		t.Fatalf("unexpected second line: %v", lines)
	}
}
