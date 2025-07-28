package main

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"

	"github.com/marang/goemqutiti/ui"
)

func layoutChips(chips []string, width int) ([]string, []chipBound) {
	var lines []string
	var row []string
	var bounds []chipBound
	curX := 0
	rowTop := 0
	chipH := lipgloss.Height(ui.ChipStyle.Render("test"))
	// Each chip is exactly chipH lines tall. Newlines inserted between rows
	// simply stack the rows without extra blank lines, so the vertical offset
	// for the next row increases by chipH.
	rowSpacing := chipH
	for _, c := range chips {
		cw := lipgloss.Width(c)
		if curX+cw > width && len(row) > 0 {
			line := lipgloss.JoinHorizontal(lipgloss.Top, row...)
			line = strings.TrimRightFunc(line, unicode.IsSpace)
			lines = append(lines, line)
			row = []string{}
			curX = 0
			rowTop += rowSpacing
		}
		row = append(row, c)
		bounds = append(bounds, chipBound{x: curX, y: rowTop, w: cw, h: chipH})
		curX += cw
	}
	if len(row) > 0 {
		line := lipgloss.JoinHorizontal(lipgloss.Top, row...)
		line = strings.TrimRightFunc(line, unicode.IsSpace)
		lines = append(lines, line)
	}
	return lines, bounds
}

func (m *model) viewClient() string {
	infoShortcuts := ui.InfoStyle.Render("Switch views: Ctrl+B brokers, Ctrl+T topics, Ctrl+P payloads, Ctrl+R traces.")
	clientID := ""
	if m.mqttClient != nil {
		r := m.mqttClient.Client.OptionsReader()
		clientID = r.ClientID()
	}
	status := strings.TrimSpace(m.connections.connection + " " + clientID)
	st := ui.ConnStyle
	if strings.HasPrefix(m.connections.connection, "Connected") {
		st = st.Foreground(ui.ColGreen)
	} else if strings.HasPrefix(m.connections.connection, "Connection lost") || strings.HasPrefix(m.connections.connection, "Failed") {
		st = st.Foreground(ui.ColWarn)
	}
	connLine := st.Render(status)
	infoLine := lipgloss.JoinVertical(lipgloss.Left, infoShortcuts, connLine)

	var chips []string
	for i, t := range m.topics.items {
		st := ui.ChipStyle
		if !t.active {
			st = ui.ChipInactive
		}
		if m.ui.focusOrder[m.ui.focusIndex] == "topics" && i == m.topics.selected {
			st = st.BorderForeground(ui.ColPurple)
		}
		chips = append(chips, st.Render(t.title))
	}
	topicsFocused := m.ui.focusOrder[m.ui.focusIndex] == "topics"
	topicFocused := m.ui.focusOrder[m.ui.focusIndex] == "topic"
	messageFocused := m.ui.focusOrder[m.ui.focusIndex] == "message"
	historyFocused := m.ui.focusOrder[m.ui.focusIndex] == "history"

	chipRows, bounds := layoutChips(chips, m.ui.width-4)
	rowH := lipgloss.Height(ui.ChipStyle.Render("test"))
	maxRows := m.layout.topics.height
	if maxRows <= 0 {
		maxRows = 3
	}
	topicsBoxHeight := maxRows * rowH
	m.topics.vp.Width = m.ui.width - 4
	m.topics.vp.Height = topicsBoxHeight
	m.topics.vp.SetContent(strings.Join(chipRows, "\n"))
	m.ensureTopicVisible()
	startLine := m.topics.vp.YOffset
	endLine := startLine + topicsBoxHeight
	topicsSP := -1.0
	if len(chipRows)*rowH > topicsBoxHeight {
		topicsSP = m.topics.vp.ScrollPercent()
	}
	chipContent := m.topics.vp.View()
	visible := []chipBound{}
	for _, b := range bounds {
		if b.y >= startLine && b.y < endLine {
			b.y -= startLine
			visible = append(visible, b)
		}
	}
	bounds = visible
	active := 0
	for _, t := range m.topics.items {
		if t.active {
			active++
		}
	}
	label := fmt.Sprintf("Topics %d/%d", active, len(m.topics.items))
	topicsBox := ui.LegendBox(chipContent, label, m.ui.width-2, topicsBoxHeight, ui.ColBlue, topicsFocused, topicsSP)
	topicBox := ui.LegendBox(m.topics.input.View(), "Topic", m.ui.width-2, 0, ui.ColBlue, topicFocused, -1)
	msgContent := m.message.input.View()
	msgLines := m.message.input.LineCount()
	msgHeight := m.layout.message.height
	msgSP := -1.0
	if msgLines > msgHeight {
		off := m.message.input.Line() - msgHeight + 1
		if off < 0 {
			off = 0
		}
		maxOff := msgLines - msgHeight
		if off > maxOff {
			off = maxOff
		}
		if maxOff > 0 {
			msgSP = float64(off) / float64(maxOff)
		}
	}
	messageBox := ui.LegendBox(msgContent, "Message (Ctrl+S publishes)", m.ui.width-2, msgHeight, ui.ColBlue, messageFocused, msgSP)
	// Calculate scroll percent for the history list
	per := m.history.list.Paginator.PerPage
	total := len(m.history.list.Items())
	histSP := -1.0
	if total > per {
		start := m.history.list.Paginator.Page * per
		denom := total - per
		if denom > 0 {
			histSP = float64(start) / float64(denom)
		}
	}
	messagesBox := ui.LegendBox(m.history.list.View(), "History (Ctrl+C copy)", m.ui.width-2, m.layout.history.height, ui.ColGreen, historyFocused, histSP)

	content := lipgloss.JoinVertical(lipgloss.Left, topicsBox, topicBox, messageBox, messagesBox)

	y := 1
	m.ui.elemPos["topics"] = y
	y += lipgloss.Height(topicsBox)
	m.ui.elemPos["topic"] = y
	y += lipgloss.Height(topicBox)
	m.ui.elemPos["message"] = y
	y += lipgloss.Height(messageBox)
	m.ui.elemPos["history"] = y

	startX := 2
	startY := m.ui.elemPos["topics"] + 1
	m.topics.chipBounds = make([]chipBound, len(bounds))
	for i, b := range bounds {
		m.topics.chipBounds[i] = chipBound{x: startX + b.x, y: startY + b.y, w: b.w, h: b.h}
	}

	box := lipgloss.NewStyle().Width(m.ui.width).Padding(0, 1, 1, 1).Render(content)
	m.ui.viewport.SetContent(box)
	m.ui.viewport.Width = m.ui.width
	// Deduct two lines for the info header rendered above the viewport.
	m.ui.viewport.Height = m.ui.height - 2

	view := m.ui.viewport.View()
	return lipgloss.JoinVertical(lipgloss.Left, infoLine, view)
}

func (m model) viewConnections() string {
	listView := m.connections.manager.ConnectionsList.View()
	help := ui.InfoStyle.Render("[enter] connect/open client  [x] disconnect  [a]dd [e]dit [d]elete  Ctrl+R traces")
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	return ui.LegendBox(content, "Brokers", m.ui.width-2, 0, ui.ColBlue, true, -1)
}

func (m model) viewForm() string {
	if m.connections.form == nil {
		return ""
	}
	listView := ui.LegendBox(m.connections.manager.ConnectionsList.View(), "Brokers", m.ui.width/2-2, 0, ui.ColBlue, false, -1)
	formLabel := "Add Broker"
	if m.connections.form.index >= 0 {
		formLabel = "Edit Broker"
	}
	formView := ui.LegendBox(m.connections.form.View(), formLabel, m.ui.width/2-2, 0, ui.ColBlue, true, -1)
	return lipgloss.JoinHorizontal(lipgloss.Top, listView, formView)
}

func (m model) viewConfirmDelete() string {
	border := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ui.ColBlue).Padding(0, 1)
	return border.Render(m.confirmPrompt)
}

func (m model) viewTopics() string {
	listView := m.topics.list.View()
	help := ui.InfoStyle.Render("[space] toggle  [d]elete  [esc] back")
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	return ui.LegendBox(content, "Topics", m.ui.width-2, 0, ui.ColBlue, false, -1)
}

func (m model) viewPayloads() string {
	listView := m.message.list.View()
	help := ui.InfoStyle.Render("[enter] load  [d]elete  [esc] back")
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	return ui.LegendBox(content, "Payloads", m.ui.width-2, 0, ui.ColBlue, false, -1)
}

func (m model) viewTraces() string {
	listView := m.traces.list.View()
	help := ui.InfoStyle.Render("[a] add  [enter] start/stop  [d] delete  [esc] back")
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	return ui.LegendBox(content, "Traces", m.ui.width-2, 0, ui.ColBlue, false, -1)
}

func (m model) viewTraceForm() string {
	content := m.traces.form.View()
	return ui.LegendBox(content, "New Trace", m.ui.width-2, 0, ui.ColBlue, false, -1)
}

func (m *model) View() string {
	switch m.ui.mode {
	case modeClient:
		return m.viewClient()
	case modeConnections:
		return m.viewConnections()
	case modeEditConnection:
		return m.viewForm()
	case modeConfirmDelete:
		return m.viewConfirmDelete()
	case modeTopics:
		return m.viewTopics()
	case modePayloads:
		return m.viewPayloads()
	case modeTracer:
		return m.viewTraces()
	case modeEditTrace:
		return m.viewTraceForm()
	default:
		return ""
	}
}
