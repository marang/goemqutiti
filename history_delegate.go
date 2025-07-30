package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/marang/goemqutiti/ui"
)

// historyDelegate renders history items with two lines and supports highlighting
// selected entries.
type historyDelegate struct{ m *model }

func (d historyDelegate) Height() int                               { return 2 }
func (d historyDelegate) Spacing() int                              { return 0 }
func (d historyDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d historyDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	hi := item.(historyItem)
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

	// Support multi-line payloads by aligning each line individually
	var lines []string
	if hi.kind != "log" {
		header := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Foreground(lblColor).Render(label),
			lipgloss.NewStyle().Foreground(ui.ColGray).Render(" "+ts+":"))
		line1 := lipgloss.PlaceHorizontal(innerWidth, align, header)
		lines = append(lines, line1)
	}
	for idx, l := range strings.Split(hi.payload, "\n") {
		wrapped := ansi.Wrap(l, innerWidth, " ")
		for _, wl := range strings.Split(wrapped, "\n") {
			fg := msgColor
			if hi.kind == "log" && idx == 0 && len(lines) == 0 {
				wl = ts + ": " + wl
				fg = ui.ColGray
			}
			rendered := lipgloss.PlaceHorizontal(innerWidth, align,
				lipgloss.NewStyle().Foreground(fg).Render(wl))
			lines = append(lines, rendered)
		}
	}
	if _, ok := d.m.history.selected[index]; ok {
		for i, l := range lines {
			lines[i] = lipgloss.NewStyle().Background(ui.ColDarkGray).Render(l)
		}
	}
	barColor := ui.ColGray
	if hi.kind == "log" {
		barColor = ui.ColDarkGray
	}
	if _, ok := d.m.history.selected[index]; ok {
		barColor = ui.ColBlue
	}
	if index == d.m.history.list.Index() {
		barColor = ui.ColPurple
	}
	bar := lipgloss.NewStyle().Foreground(barColor)
	lines = ui.FormatHistoryLines(lines, width, bar)
	fmt.Fprint(w, strings.Join(lines, "\n"))
}
