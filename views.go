package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *model) viewClient() string {
	header := legendBox("GoEmqutiti - MQTT Client", "App", m.width-4, false)
	info := legendBox("Press Ctrl+M for connections, Ctrl+T topics, Ctrl+P payloads", "Help", m.width-4, false)
	conn := legendBox(m.connection, "Connection", m.width-4, false)

	var chips []string
	for i, t := range m.topics {
		st := chipStyle
		if !t.active {
			st = chipInactive
		}
		if m.focusIndex == 2 && i == m.selectedTopic {
			st = st.Copy().BorderForeground(lipgloss.Color("212"))
		}
		chips = append(chips, st.Render(t.title))
	}
	topicsBox := legendBox(lipgloss.JoinHorizontal(lipgloss.Top, chips...), "Topics", m.width-4, m.focusIndex == 2)

	messagesBox := legendGreenBox(m.history.View(), "History (Ctrl+C copy)", m.width-4)

	topicBox := legendBox(m.topicInput.View(), "Topic", m.width-4, m.focusIndex == 0)
	messageBox := legendBox(m.messageInput.View(), "Message", m.width-4, m.focusIndex == 1)

	inputsBox := lipgloss.JoinVertical(lipgloss.Left, topicBox, messageBox)

	var payloadLines []string
	for topic, payload := range m.payloads {
		payloadLines = append(payloadLines, fmt.Sprintf("- %s: %s", topic, payload))
	}
	payloadBox := legendBox(strings.Join(payloadLines, "\n"), "Payloads", m.width-4, false)
	content := lipgloss.JoinVertical(lipgloss.Left, header, info, conn, topicsBox, messagesBox, inputsBox, payloadBox)

	y := 1
	y += lipgloss.Height(header)
	y += lipgloss.Height(info)
	y += lipgloss.Height(conn)
	y += lipgloss.Height(topicsBox)
	y += lipgloss.Height(messagesBox)
	m.elemPos["topic"] = y
	y += lipgloss.Height(topicBox)
	m.elemPos["message"] = y

	box := lipgloss.NewStyle().Width(m.width).Padding(1, 1).Render(content)
	m.viewport.SetContent(box)
	m.viewport.Width = m.width
	m.viewport.Height = m.height
	return m.viewport.View()
}

func (m model) viewConnections() string {
	listView := m.connections.ConnectionsList.View()
	help := "[enter] connect  [a]dd [e]dit [d]elete  [esc] back"
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	return borderStyle.Copy().Width(m.width - 2).Height(m.height - 2).Render(content)
}

func (m model) viewForm() string {
	if m.connForm == nil {
		return ""
	}
	listView := m.connections.ConnectionsList.View()
	formView := m.connForm.View()
	left := borderStyle.Copy().Width(m.width/2 - 2).Render(listView)
	right := borderStyle.Copy().Width(m.width/2 - 2).Render(formView)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m model) viewConfirmDelete() string {
	var name string
	if m.deleteIndex >= 0 && m.deleteIndex < len(m.connections.Profiles) {
		name = m.connections.Profiles[m.deleteIndex].Name
	}
	border := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Padding(0, 1)
	return border.Render(fmt.Sprintf("Delete connection '%s'? [y/n]", name))
}

func (m model) viewTopics() string {
	listView := m.topicsList.View()
	help := "[space] toggle  [d]elete  [esc] back"
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	return borderStyle.Copy().Width(m.width - 2).Height(m.height - 2).Render(content)
}

func (m model) viewPayloads() string {
	listView := m.payloadList.View()
	help := "[enter] load  [d]elete  [esc] back"
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	return borderStyle.Copy().Width(m.width - 2).Height(m.height - 2).Render(content)
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
