package main

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/emqutiti/ui"
)

type connectionDelegate struct{}

// Height returns the row height for connection entries.
func (d connectionDelegate) Height() int { return 3 }

// Spacing returns the spacing between connection entries.
func (d connectionDelegate) Spacing() int { return 0 }

// Update performs no action for connection items.
func (d connectionDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

// Render writes a single connection row to the list.
func (d connectionDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	ci := item.(connectionItem)
	width := m.Width()
	border := " "
	if index == m.Index() {
		border = lipgloss.NewStyle().Foreground(ui.ColPurple).Render("┃")
	}
	name := lipgloss.PlaceHorizontal(width-2, lipgloss.Left, ci.title)
	color := ui.ColGray
	switch ci.status {
	case "connected":
		color = ui.ColGreen
	case "disconnected":
		color = ui.ColWarn
	case "connecting":
		color = ui.ColCyan
	}
	status := lipgloss.NewStyle().Foreground(color).Render(ci.status)
	status = lipgloss.PlaceHorizontal(width-2, lipgloss.Left, status)
	detail := lipgloss.PlaceHorizontal(width-2, lipgloss.Left,
		lipgloss.NewStyle().Foreground(ui.ColGray).Render(ci.detail))
	fmt.Fprintf(w, "%s %s\n%s %s\n%s %s", border, name, border, status, border, detail)
}
