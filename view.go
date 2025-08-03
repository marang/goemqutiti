package emqutiti

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/marang/emqutiti/ui"
)

func (m *model) overlayHelp(view string) string {
	help := ui.HelpStyle.Render("?")
	if m.help.Focused() {
		help = ui.HelpFocused.Render("?")
	}
	m.ui.elemPos[idHelp] = 0
	lines := strings.Split(view, "\n")
	if len(lines) == 0 {
		return help
	}
	first := lipgloss.NewStyle().Width(m.ui.width-lipgloss.Width(help)).Render(lines[0]) + help
	if len(lines) == 1 {
		return first
	}
	lines[0] = first
	return strings.Join(lines, "\n")
}

// OverlayHelp wraps overlayHelp to satisfy component interfaces.
func (m *model) OverlayHelp(view string) string { return m.overlayHelp(view) }

// View renders the application UI based on the current mode.
func (m *model) View() string {
	if c, ok := m.components[m.currentMode()]; ok {
		return c.View()
	}
	return ""
}
