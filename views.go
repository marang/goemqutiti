package main

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
)

func layoutChips(chips []string, width int) (string, []chipBound) {
	var lines []string
	var row []string
	var bounds []chipBound
	curX := 0
	rowTop := 0
	chipH := lipgloss.Height(chipStyle.Render("test"))
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
	return strings.Join(lines, "\n"), bounds
}

func (m *model) viewClient() string {
	infoShortcuts := infoStyle.Render("Switch views: Ctrl+B brokers, Ctrl+T topics, Ctrl+P payloads.")
	clientID := ""
	if m.mqttClient != nil {
		r := m.mqttClient.Client.OptionsReader()
		clientID = r.ClientID()
	}
	connLine := connStyle.Render(strings.TrimSpace(m.connection + " " + clientID))
	infoLine := lipgloss.JoinVertical(lipgloss.Left, infoShortcuts, connLine)

	var chips []string
	for i, t := range m.topics {
		st := chipStyle
		if !t.active {
			st = chipInactive
		}
		if m.focusOrder[m.focusIndex] == "topics" && i == m.selectedTopic {
			st = st.BorderForeground(colPurple)
		}
		chips = append(chips, st.Render(t.title))
	}
	topicsFocused := m.focusOrder[m.focusIndex] == "topics"
	topicFocused := m.focusOrder[m.focusIndex] == "topic"
	messageFocused := m.focusOrder[m.focusIndex] == "message"
	historyFocused := m.focusOrder[m.focusIndex] == "history"

	chipContent, bounds := layoutChips(chips, m.width-4)
	topicsBox := legendBox(chipContent, "Topics", m.width-2, topicsFocused)
	topicBox := legendBox(m.topicInput.View(), "Topic", m.width-2, topicFocused)
	messageBox := legendBox(m.messageInput.View(), "Message (Ctrl+S publishes)", m.width-2, messageFocused)
	messagesBox := legendGreenBox(m.history.View(), "History (Ctrl+C copy)", m.width-2, historyFocused)

	content := lipgloss.JoinVertical(lipgloss.Left, topicsBox, topicBox, messageBox, messagesBox)

	y := 1
	m.elemPos["topics"] = y
	y += lipgloss.Height(topicsBox)
	m.elemPos["topic"] = y
	y += lipgloss.Height(topicBox)
	m.elemPos["message"] = y
	y += lipgloss.Height(messageBox)
	m.elemPos["history"] = y

	startX := 2
	startY := m.elemPos["topics"] + 2
	m.chipBounds = make([]chipBound, len(bounds))
	for i, b := range bounds {
		m.chipBounds[i] = chipBound{x: startX + b.x, y: startY + b.y, w: b.w, h: b.h}
	}

	box := lipgloss.NewStyle().Width(m.width).Padding(0, 1, 1, 1).Render(content)
	m.viewport.SetContent(box)
	m.viewport.Width = m.width
	// Deduct two lines for the info header rendered above the viewport.
	m.viewport.Height = m.height - 2
	return lipgloss.JoinVertical(lipgloss.Left, infoLine, m.viewport.View())
}

func (m model) viewConnections() string {
	listView := m.connections.ConnectionsList.View()
	help := infoStyle.Render("[enter] connect  [x] disconnect  [a]dd [e]dit [d]elete")
	var parts []string
	if strings.TrimSpace(m.connection) != "" {
		parts = append(parts, connStyle.Render(strings.TrimSpace(m.connection)))
	}
	parts = append(parts, listView, help)
	content := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return legendBox(content, "Brokers", m.width-2, true)
}

func (m model) viewForm() string {
	if m.connForm == nil {
		return ""
	}
	listView := legendBox(m.connections.ConnectionsList.View(), "Brokers", m.width/2-2, false)
	formLabel := "Add Broker"
	if m.connForm.index >= 0 {
		formLabel = "Edit Broker"
	}
	formView := legendBox(m.connForm.View(), formLabel, m.width/2-2, true)
	return lipgloss.JoinHorizontal(lipgloss.Top, listView, formView)
}

func (m model) viewConfirmDelete() string {
	border := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colBlue).Padding(0, 1)
	return border.Render(m.confirmPrompt)
}

func (m model) viewTopics() string {
	listView := m.topicsList.View()
	help := infoStyle.Render("[space] toggle  [d]elete  [esc] back")
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	return legendBox(content, "Topics", m.width-2, false)
}

func (m model) viewPayloads() string {
	listView := m.payloadList.View()
	help := infoStyle.Render("[enter] load  [d]elete  [esc] back")
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	return legendBox(content, "Payloads", m.width-2, false)
}

func (m *model) View() string {
	switch m.mode {
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
	default:
		return ""
	}
}
