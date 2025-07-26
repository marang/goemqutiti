package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
		lblColor = lipgloss.Color("205")
		msgColor = lipgloss.Color("219")
	case "pub":
		label = fmt.Sprintf("PUB %s:", hi.topic)
		lblColor = lipgloss.Color("63")
		msgColor = lipgloss.Color("81")
	default:
		fmt.Fprint(w, lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Width(width).Render(hi.payload))
		return
	}
	align := lipgloss.Left
	if hi.kind == "pub" {
		align = lipgloss.Right
	}
	line1 := lipgloss.PlaceHorizontal(width, align, lipgloss.NewStyle().Foreground(lblColor).Render(label))
	line2 := lipgloss.PlaceHorizontal(width, align, lipgloss.NewStyle().Foreground(msgColor).Render(hi.payload))
	lines := []string{line1, line2}
	if _, ok := d.m.selectedHistory[index]; ok {
		for i, l := range lines {
			lines[i] = lipgloss.NewStyle().Background(lipgloss.Color("236")).Render(l)
		}
	}
	border := " "
	if _, ok := d.m.selectedHistory[index]; ok {
		border = lipgloss.NewStyle().Foreground(lipgloss.Color("63")).Render("┃")
	}
	if index == d.m.history.Index() {
		border = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Render("┃")
	}
	for i, l := range lines {
		lines[i] = border + " " + lipgloss.PlaceHorizontal(width-2, lipgloss.Left, l)
	}
	fmt.Fprint(w, strings.Join(lines, "\n"))
}
