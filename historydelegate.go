package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
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
	var lblColor lipgloss.Color
	var msgColor lipgloss.Color
	switch hi.kind {
	case "sub":
		label = fmt.Sprintf("SUB %s:", hi.topic)
		lblColor = colPink
		msgColor = colPub
	case "pub":
		label = fmt.Sprintf("PUB %s:", hi.topic)
		lblColor = colBlue
		msgColor = colSub
	default:
		label = ""
		lblColor = colGray
		msgColor = colGray
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
		line1 := lipgloss.PlaceHorizontal(innerWidth, align,
			lipgloss.NewStyle().Foreground(lblColor).Render(label))
		lines = append(lines, line1)
	}
	for _, l := range strings.Split(hi.payload, "\n") {
		wrapped := ansi.Wrap(l, innerWidth, " ")
		for _, wl := range strings.Split(wrapped, "\n") {
			rendered := lipgloss.PlaceHorizontal(innerWidth, align,
				lipgloss.NewStyle().Foreground(msgColor).Render(wl))
			lines = append(lines, rendered)
		}
	}
	if _, ok := d.m.selectedHistory[index]; ok {
		for i, l := range lines {
			lines[i] = lipgloss.NewStyle().Background(colDarkGray).Render(l)
		}
	}
	border := lipgloss.NewStyle().Foreground(colGray).Render("┃")
	if _, ok := d.m.selectedHistory[index]; ok {
		border = lipgloss.NewStyle().Foreground(colBlue).Render("┃")
	}
	if index == d.m.history.Index() {
		border = lipgloss.NewStyle().Foreground(colPurple).Render("┃")
	}
	for i, l := range lines {
		lines[i] = border + " " + lipgloss.PlaceHorizontal(width-2, lipgloss.Left, l)
	}
	fmt.Fprint(w, strings.Join(lines, "\n"))
}
