package main

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
)

func wrapChips(chips []string, width int) string {
	var lines []string
	var row []string
	cur := 0
	for _, c := range chips {
		cw := lipgloss.Width(c)
		if cur+cw > width && len(row) > 0 {
			line := lipgloss.JoinHorizontal(lipgloss.Top, row...)
			line = strings.TrimRightFunc(line, unicode.IsSpace)
			lines = append(lines, line)
			row = []string{c}
			cur = cw
		} else {
			row = append(row, c)
			cur += cw
		}
	}
	if len(row) > 0 {
		line := lipgloss.JoinHorizontal(lipgloss.Top, row...)
		line = strings.TrimRightFunc(line, unicode.IsSpace)
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (m *model) viewClient() string {
	infoLine := infoStyle.Render("Info: Press Ctrl+B for brokers, Ctrl+T topics, Ctrl+P payloads. " + m.connection)

	var chips []string
	for i, t := range m.topics {
		st := chipStyle
		if !t.active {
			st = chipInactive
		}
		if m.focusOrder[m.focusIndex] == "topics" && i == m.selectedTopic {
			st = st.BorderForeground(lipgloss.Color("212"))
		}
		chips = append(chips, st.Render(t.title))
	}
	topicsFocused := m.focusOrder[m.focusIndex] == "topics"
	topicFocused := m.focusOrder[m.focusIndex] == "topic"
	messageFocused := m.focusOrder[m.focusIndex] == "message"
	historyFocused := m.focusOrder[m.focusIndex] == "history"

	topicsBox := legendBox(wrapChips(chips, m.width-4), "Topics", m.width-2, topicsFocused)
	topicBox := legendBox(m.topicInput.View(), "Topic", m.width-2, topicFocused)
	messageBox := legendBox(m.messageInput.View(), "Message", m.width-2, messageFocused)
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

	box := lipgloss.NewStyle().Width(m.width).Padding(1, 1).Render(content)
	m.viewport.SetContent(box)
	m.viewport.Width = m.width
	m.viewport.Height = m.height - 1
	return lipgloss.JoinVertical(lipgloss.Left, infoLine, m.viewport.View())
}

func (m model) viewConnections() string {
	listView := m.connections.ConnectionsList.View()
	help := "[enter] connect  [a]dd [e]dit [d]elete  [esc] back"
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	return borderStyle.Width(m.width - 2).Height(m.height - 2).Render(content)
}

func (m model) viewForm() string {
	if m.connForm == nil {
		return ""
	}
	listView := m.connections.ConnectionsList.View()
	formView := m.connForm.View()
	left := borderStyle.Width(m.width/2 - 2).Render(listView)
	right := borderStyle.Width(m.width/2 - 2).Render(formView)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m model) viewConfirmDelete() string {
	border := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Padding(0, 1)
	return border.Render(m.confirmPrompt)
}

func (m model) viewTopics() string {
	listView := m.topicsList.View()
	help := "[space] toggle  [d]elete  [esc] back"
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	return legendBox(content, "Topics", m.width-2, false)
}

func (m model) viewPayloads() string {
	listView := m.payloadList.View()
	help := "[enter] load  [d]elete  [esc] back"
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
