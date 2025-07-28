package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// HistoryView renders lines in the same style as the history log and supports
// scrolling with the arrow keys.
type HistoryView struct {
	vp viewport.Model
}

// NewHistoryView creates a HistoryView sized for the given outer box width and height.
func NewHistoryView(boxWidth, height int) HistoryView {
	vp := viewport.New(boxWidth-4, height)
	return HistoryView{vp: vp}
}

// SetSize adjusts the viewport size for the given outer box width and height.
func (h *HistoryView) SetSize(boxWidth, height int) {
	h.vp.Width = boxWidth - 4
	h.vp.Height = height
}

// SetLines replaces the displayed lines.
func (h *HistoryView) SetLines(lines []string) {
	width := h.vp.Width
	inner := width - 2
	if inner < 0 {
		inner = 0
	}
	bar := lipgloss.NewStyle().Foreground(ColDarkGray).Render("┃")
	var out []string
	for _, l := range lines {
		wrapped := ansi.Wrap(l, inner, " ")
		for _, wl := range strings.Split(wrapped, "\n") {
			out = append(out, bar+" "+lipgloss.PlaceHorizontal(width-2, lipgloss.Left, wl))
		}
	}
	h.vp.SetContent(strings.Join(out, "\n"))
}

// Update forwards messages to the underlying viewport.
func (h *HistoryView) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	h.vp, cmd = h.vp.Update(msg)
	return cmd
}

// View returns the viewport content.
func (h HistoryView) View() string { return h.vp.View() }

// GotoBottom scrolls to the end of the list.
func (h *HistoryView) GotoBottom() { h.vp.GotoBottom() }

// GotoTop scrolls to the start of the list.
func (h *HistoryView) GotoTop() { h.vp.GotoTop() }
