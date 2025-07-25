package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) viewClient() string {
	header := borderStyle.Copy().Width(m.width - 4).Render("GoEmqutiti - MQTT Client")
	info := borderStyle.Copy().Width(m.width - 4).Render("Press Ctrl+M for connections, Ctrl+T topics, Ctrl+P payloads")
	conn := borderStyle.Copy().Width(m.width - 4).Render("Connection: " + m.connection)

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
	topicsBox := borderStyle.Copy().Width(m.width - 4).Render("Topics:\n" + lipgloss.JoinHorizontal(lipgloss.Top, chips...))

	m.history.SetSize(m.width-4, m.height/3)
	messagesBox := legendBox(m.history.View(), "History (Ctrl+C copy)", m.width-4)

	inputs := lipgloss.JoinVertical(lipgloss.Left,
		"Topic:\n"+m.topicInput.View(),
		"Message:\n"+m.messageInput.View(),
	)
	inputsBox := borderStyle.Copy().Width(m.width - 4).Render(inputs)

	var payloadLines []string
	for topic, payload := range m.payloads {
		payloadLines = append(payloadLines, fmt.Sprintf("- %s: %s", topic, payload))
	}
	payloadHelp := "Stored Payloads (press Ctrl+P to manage):"
	payloadBox := borderStyle.Copy().Width(m.width - 4).Render(payloadHelp + "\n" + strings.Join(payloadLines, "\n"))

	content := lipgloss.JoinVertical(lipgloss.Left, header, info, conn, topicsBox, messagesBox, inputsBox, payloadBox)
	return lipgloss.NewStyle().Width(m.width).Height(m.height).Padding(1, 1).Render(content)
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

func (m model) View() string {
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
