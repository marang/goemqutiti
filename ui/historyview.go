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

// FormatHistoryLines wraps lines and prefixes them with a colored bar.
func FormatHistoryLines(lines []string, width int, bar lipgloss.Style) []string {
	inner := width - 2
	if inner < 0 {
		inner = 0
	}
	var out []string
	prefix := bar.Render("┃")
	for _, l := range lines {
		wrapped := ansi.Wrap(l, inner, " ")
		for _, wl := range strings.Split(wrapped, "\n") {
			out = append(out, prefix+" "+lipgloss.PlaceHorizontal(width-2, lipgloss.Left, wl))
		}
	}
	return out
}

// FormatHistoryLinesWithFocus wraps lines and prefixes them with a colored bar,
// with an option to highlight specific lines.
func FormatHistoryLinesWithFocus(lines []string, width int, bar lipgloss.Style, focusedLine int) []string {
	inner := width - 2
	if inner < 0 {
		inner = 0
	}
	var out []string
	prefix := bar.Render("┃")
	focusedPrefix := lipgloss.NewStyle().Foreground(ColPink).Render("┃")

	for i, l := range lines {
		currentPrefix := prefix
		if i == focusedLine {
			currentPrefix = focusedPrefix
		}

		wrapped := ansi.Wrap(l, inner, " ")
		for _, wl := range strings.Split(wrapped, "\n") {
			content := wl
			if i == focusedLine {
				content = lipgloss.NewStyle().Foreground(ColPink).Render(wl)
			}
			out = append(out, currentPrefix+" "+lipgloss.PlaceHorizontal(width-2, lipgloss.Left, content))
		}
	}
	return out
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
	bar := lipgloss.NewStyle().Foreground(ColDarkGray)
	out := FormatHistoryLines(lines, h.vp.Width, bar)
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

// ScrollPercent returns the scroll position as a fraction between 0 and 1.
func (h HistoryView) ScrollPercent() float64 { return h.vp.ScrollPercent() }

// GotoBottom scrolls to the end of the list.
func (h *HistoryView) GotoBottom() { h.vp.GotoBottom() }

// GotoTop scrolls to the start of the list.
func (h *HistoryView) GotoTop() { h.vp.GotoTop() }
