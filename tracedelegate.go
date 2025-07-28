package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/marang/goemqutiti/history"
	"github.com/marang/goemqutiti/ui"
)

// traceMsgItem holds a trace message with its sequence number.
type traceMsgItem struct {
	idx int
	msg history.Message
}

func (t traceMsgItem) FilterValue() string { return t.msg.Payload }
func (t traceMsgItem) Title() string       { return t.msg.Topic }
func (t traceMsgItem) Description() string { return t.msg.Payload }

// traceMsgDelegate renders trace messages with numbering and timestamp.
type traceMsgDelegate struct{ m *model }

func (d traceMsgDelegate) Height() int                               { return 2 }
func (d traceMsgDelegate) Spacing() int                              { return 0 }
func (d traceMsgDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d traceMsgDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	it := item.(traceMsgItem)
	width := m.Width()
	innerWidth := width - 2
	if innerWidth < 0 {
		innerWidth = 0
	}
	header := fmt.Sprintf("%d %s %s:", it.idx, it.msg.Timestamp.Format(time.RFC3339), it.msg.Topic)
	lines := []string{lipgloss.PlaceHorizontal(innerWidth, lipgloss.Left,
		lipgloss.NewStyle().Foreground(ui.ColBlue).Render(header))}
	for _, l := range strings.Split(it.msg.Payload, "\n") {
		wrapped := ansi.Wrap(l, innerWidth, " ")
		for _, wl := range strings.Split(wrapped, "\n") {
			lines = append(lines, lipgloss.PlaceHorizontal(innerWidth, lipgloss.Left,
				lipgloss.NewStyle().Foreground(ui.ColSub).Render(wl)))
		}
	}
	barColor := ui.ColDarkGray
	if index == d.m.traces.view.Index() {
		barColor = ui.ColPurple
	}
	bar := lipgloss.NewStyle().Foreground(barColor)
	lines = ui.FormatHistoryLines(lines, width, bar)
	fmt.Fprint(w, strings.Join(lines, "\n"))
}
