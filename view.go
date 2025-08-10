package emqutiti

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"github.com/marang/emqutiti/ui"
)

const helpReflowWidth = 60

func (m *model) overlayHelp(view string) string {
	help := ui.HelpStyle.Render("?")
	if m.help.Focused() {
		help = ui.HelpFocused.Render("?")
	}
	m.ui.elemPos[idHelp] = 0

	info := "Switch views: Ctrl+B brokers, Ctrl+T topics, Ctrl+P payloads, Ctrl+R traces, Ctrl+L logs, Ctrl+D quit."
	pad := lipgloss.Width(ui.InfoStyle.Render(""))

	lines := []string{}
	if view != "" {
		lines = strings.Split(view, "\n")
	}
	lines = append([]string{""}, lines...)

	if m.ui.width < helpReflowWidth {
		available := m.ui.width - pad
		if available < 0 {
			available = 0
		}
		if runewidth.StringWidth(info) > available {
			info = runewidth.Truncate(info, available, "")
		}
		lines[0] = lipgloss.NewStyle().Width(available + pad).Render(ui.InfoStyle.Render(info))
		if len(lines) > 1 {
			secondAvail := m.ui.width - lipgloss.Width(help)
			if secondAvail < 0 {
				secondAvail = 0
			}
			second := lipgloss.JoinHorizontal(lipgloss.Top,
				lipgloss.NewStyle().Width(secondAvail).Render(lines[1]), help)
			lines[1] = second
		} else {
			lines = append(lines, help)
		}
	} else {
		available := m.ui.width - lipgloss.Width(help) - pad
		if available < 0 {
			available = 0
		}
		if runewidth.StringWidth(info) > available {
			info = runewidth.Truncate(info, available, "")
		}
		first := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(available+pad).Render(ui.InfoStyle.Render(info)), help)
		lines[0] = first
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// OverlayHelp wraps overlayHelp to satisfy component interfaces.
func (m *model) OverlayHelp(view string) string { return m.overlayHelp(view) }

// View renders the application UI based on the current mode.
func (m *model) View() string {
	if c, ok := m.components[m.CurrentMode()]; ok {
		return c.View()
	}
	return ""
}
