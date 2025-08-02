package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/marang/goemqutiti/ui"
)

func (m *model) overlayHelp(view string) string {
	help := ui.HelpStyle.Render("?")
	if m.help.focused {
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

// View renders the application UI based on the current mode.
func (m *model) View() string {
	switch m.currentMode() {
	case modeClient:
		return m.viewClient()
	case modeConnections:
		return m.viewConnections()
	case modeEditConnection:
		return m.viewForm()
	case modeConfirmDelete:
		return m.viewConfirmDelete()
	case modeTopics:
		return m.viewTopics()
	case modePayloads:
		return m.viewPayloads()
	case modeTracer:
		return m.viewTraces()
	case modeEditTrace:
		return m.viewTraceForm()
	case modeViewTrace:
		return m.viewTraceMessages()
	case modeImporter:
		return m.viewImporter()
	case modeHistoryFilter:
		return m.viewHistoryFilter()
	case modeHistoryDetail:
		return m.viewHistoryDetail()
	case modeHelp:
		return m.viewHelp()
	default:
		return ""
	}
}
