package history

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/marang/emqutiti/ui"
)

const historyPreviewLimit = 256

// historyDelegate renders history items with two lines and supports highlighting
// selected entries. It has no direct dependency on the application model.
type historyDelegate struct{}

// Height returns the fixed height for history entries.
func (d historyDelegate) Height() int { return 2 }

// Spacing returns the row spacing for history entries.
func (d historyDelegate) Spacing() int { return 0 }

// Update performs no update and returns nil.
func (d historyDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

// Render prints a history item with its label and payload.
func (d historyDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	hi := item.(Item)
	width := m.Width()
	var label string
	ts := hi.timestamp.Format("2006-01-02 15:04:05.000")
	var lblColor lipgloss.Color
	var msgColor lipgloss.Color
	switch hi.kind {
	case "sub":
		label = fmt.Sprintf("SUB %s", hi.topic)
		lblColor = ui.ColPink
		msgColor = ui.ColPub
	case "pub":
		label = fmt.Sprintf("PUB %s", hi.topic)
		lblColor = ui.ColBlue
		msgColor = ui.ColSub
	default:
		label = ""
		lblColor = ui.ColGray
		msgColor = ui.ColGray
	}
	align := lipgloss.Left
	if hi.kind == "pub" {
		align = lipgloss.Right
	}
	innerWidth := width - 2
	if innerWidth < 0 {
		innerWidth = 0
	}

	// Render at most two lines so the list height stays consistent
	var lines []string
	if hi.kind != "log" {
		header := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Foreground(lblColor).Render(label),
			lipgloss.NewStyle().Foreground(ui.ColGray).Render(" "+ts+":"))
		lines = append(lines, lipgloss.PlaceHorizontal(innerWidth, align, header))
	}
	payload := strings.ReplaceAll(hi.payload, "\r\n", "\n")
	payload = strings.ReplaceAll(payload, "\n", "\u23ce")
	more := utf8.RuneCountInString(payload) > historyPreviewLimit
	if more {
		payload = ansi.Truncate(payload, historyPreviewLimit, "")
	}
	trunc := ansi.Truncate(hi.payload, innerWidth, "")
	trunc = strings.NewReplacer("\r\n", "\u23ce", "\n", "\u23ce").Replace(trunc)
	if more || lipgloss.Width(hi.payload) > innerWidth {
		if lipgloss.Width(trunc) >= innerWidth {
			trunc = ansi.Truncate(trunc, innerWidth-1, "")
		}
		trunc += "\u2026"
	}
	fg := msgColor
	if hi.kind == "log" && len(lines) == 0 {
		trunc = ts + ": " + trunc
		fg = ui.ColGray
	}
	lines = append(lines, lipgloss.PlaceHorizontal(innerWidth, align,
		lipgloss.NewStyle().Foreground(fg).Render(trunc)))
	if len(lines) < 2 {
		lines = append(lines, lipgloss.PlaceHorizontal(innerWidth, align, ""))
	}
	if hi.isSelected != nil && *hi.isSelected {
		for i, l := range lines {
			lines[i] = lipgloss.NewStyle().Background(ui.ColDarkGray).Render(l)
		}
	}
	barColor := ui.ColGray
	if hi.kind == "log" {
		barColor = ui.ColDarkGray
	}
	if hi.isSelected != nil && *hi.isSelected {
		barColor = ui.ColBlue
	}
	if index == m.Index() {
		barColor = ui.ColPurple
	}
	bar := lipgloss.NewStyle().Foreground(barColor)
	lines = ui.FormatHistoryLines(lines, width, bar)
	fmt.Fprint(w, strings.Join(lines, "\n"))
}
