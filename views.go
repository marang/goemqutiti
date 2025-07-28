package main

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"

	"goemqutiti/ui"
)

func layoutChips(chips []string, width int) (string, []chipBound) {
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
	return strings.Join(lines, "\n"), bounds
}

func (m *model) viewClient() string {
	infoShortcuts := ui.InfoStyle.Render("Switch views: Ctrl+B brokers, Ctrl+T topics, Ctrl+P payloads.")
	clientID := ""
	if m.mqttClient != nil {
		r := m.mqttClient.Client.OptionsReader()
		clientID = r.ClientID()
	}
	status := strings.TrimSpace(m.connection + " " + clientID)
	st := ui.ConnStyle
	if strings.HasPrefix(m.connection, "Connected") {
		st = st.Foreground(ui.ColGreen)
	} else if strings.HasPrefix(m.connection, "Connection lost") || strings.HasPrefix(m.connection, "Failed") {
		st = st.Foreground(ui.ColWarn)
	}
	connLine := st.Render(status)
	infoLine := lipgloss.JoinVertical(lipgloss.Left, infoShortcuts, connLine)

	var chips []string
	for i, t := range m.topics {
		st := ui.ChipStyle
		if !t.active {
			st = ui.ChipInactive
		}
		if m.focusOrder[m.focusIndex] == "topics" && i == m.selectedTopic {
			st = st.BorderForeground(ui.ColPurple)
		}
		chips = append(chips, st.Render(t.title))
	}
	topicsFocused := m.focusOrder[m.focusIndex] == "topics"
	topicFocused := m.focusOrder[m.focusIndex] == "topic"
	messageFocused := m.focusOrder[m.focusIndex] == "message"
	historyFocused := m.focusOrder[m.focusIndex] == "history"

	chipContent, bounds := layoutChips(chips, m.width-4)
	maxRows := m.topicsHeight
	if maxRows <= 0 {
		maxRows = 3
	}
	lines := strings.Split(chipContent, "\n")
	if len(lines) > maxRows {
		chipContent = strings.Join(lines[:maxRows], "\n")
	}
	rowH := lipgloss.Height(ui.ChipStyle.Render("test"))
	visible := []chipBound{}
	for _, b := range bounds {
		if b.y/rowH < maxRows {
			visible = append(visible, b)
		}
	}
	bounds = visible
	active := 0
	for _, t := range m.topics {
		if t.active {
			active++
		}
	}
	label := fmt.Sprintf("Topics %d/%d", active, len(m.topics))
	topicsBoxHeight := maxRows * rowH
	topicsBox := ui.LegendBoxSized(chipContent, label, m.width-2, topicsBoxHeight, topicsFocused)
	topicBox := ui.LegendBox(m.topicInput.View(), "Topic", m.width-2, topicFocused)
	messageBox := ui.LegendBoxSized(m.messageInput.View(), "Message (Ctrl+S publishes)", m.width-2, m.messageHeight, messageFocused)
	messagesBox := ui.LegendGreenBoxSized(m.history.View(), "History (Ctrl+C copy)", m.width-2, m.historyHeight, historyFocused)

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
	help := ui.InfoStyle.Render("[enter] connect/open client  [x] disconnect  [a]dd [e]dit [d]elete")
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	return ui.LegendBox(content, "Brokers", m.width-2, true)
}

func (m model) viewForm() string {
	if m.connForm == nil {
		return ""
	}
	listView := ui.LegendBox(m.connections.ConnectionsList.View(), "Brokers", m.width/2-2, false)
	formLabel := "Add Broker"
	if m.connForm.index >= 0 {
		formLabel = "Edit Broker"
	}
	formView := ui.LegendBox(m.connForm.View(), formLabel, m.width/2-2, true)
	return lipgloss.JoinHorizontal(lipgloss.Top, listView, formView)
}

func (m model) viewConfirmDelete() string {
	border := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ui.ColBlue).Padding(0, 1)
	return border.Render(m.confirmPrompt)
}

func (m model) viewTopics() string {
	listView := m.topicsList.View()
	help := ui.InfoStyle.Render("[space] toggle  [d]elete  [esc] back")
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	return ui.LegendBox(content, "Topics", m.width-2, false)
}

func (m model) viewPayloads() string {
	listView := m.payloadList.View()
	help := ui.InfoStyle.Render("[enter] load  [d]elete  [esc] back")
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	return ui.LegendBox(content, "Payloads", m.width-2, false)
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
