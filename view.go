package emqutiti

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"github.com/marang/emqutiti/ui"
)

func (m *model) overlayHelp(view string) string {
	help := ui.HelpStyle.Render("?")
	if m.help.Focused() {
		help = ui.HelpFocused.Render("?")
	}
	m.ui.elemPos[idHelp] = 0

	info := "Switch views: Ctrl+B brokers, Ctrl+T topics, Ctrl+P payloads, Ctrl+R traces, Ctrl+D quit."
	pad := lipgloss.Width(ui.InfoStyle.Render(""))
	available := m.ui.width - lipgloss.Width(help) - pad
	if available < 0 {
		available = 0
	}
	if runewidth.StringWidth(info) > available {
		info = runewidth.Truncate(info, available, "")
	}
	infoShortcuts := ui.InfoStyle.Render(info)
	lines := []string{}
	if view != "" {
		lines = strings.Split(view, "\n")
	}
	lines = append([]string{infoShortcuts}, lines...)

	first := lipgloss.NewStyle().Width(available+pad).Render(lines[0]) + help
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
	if c, ok := m.components[m.CurrentMode()]; ok {
		return c.View()
	}
	return ""
}
