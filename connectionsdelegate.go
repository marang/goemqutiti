package main

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type connectionDelegate struct{}

func (d connectionDelegate) Height() int                               { return 3 }
func (d connectionDelegate) Spacing() int                              { return 0 }
func (d connectionDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d connectionDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	ci := item.(connectionItem)
	width := m.Width()
	border := " "
	if index == m.Index() {
		border = lipgloss.NewStyle().Foreground(colPurple).Render("┃")
	}
	name := lipgloss.PlaceHorizontal(width-2, lipgloss.Left, ci.title)
	status := lipgloss.NewStyle().Foreground(colGray).Render(ci.status)
	status = lipgloss.PlaceHorizontal(width-2, lipgloss.Left, status)
	detail := lipgloss.PlaceHorizontal(width-2, lipgloss.Left,
		lipgloss.NewStyle().Foreground(colGray).Render(ci.detail))
	fmt.Fprintf(w, "%s %s\n%s %s\n%s %s", border, name, border, status, border, detail)
}
