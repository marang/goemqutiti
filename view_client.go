package emqutiti

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/marang/emqutiti/topics"
	"github.com/marang/emqutiti/ui"
)

// clientInfoLine renders the connection status.
func (m *model) clientInfoLine() string {
	clientID := ""
	if m.mqttClient != nil {
		r := m.mqttClient.Client.OptionsReader()
		clientID = r.ClientID()
	}
	status := strings.TrimSpace(m.connections.Connection + " " + clientID)
	st := ui.InfoSubtleStyle
	if strings.HasPrefix(m.connections.Connection, "Connected") {
		st = st.Foreground(ui.ColGreen)
	} else if strings.HasPrefix(m.connections.Connection, "Connection lost") || strings.HasPrefix(m.connections.Connection, "Failed") {
		st = st.Foreground(ui.ColWarn)
	}
	return st.Render(status)
}

// topicTooltip renders a tooltip for the selected topic when it exceeds the
// viewport width.
func (m *model) topicTooltip() string {
	sel := m.topics.Selected()
	if sel < 0 || sel >= len(m.topics.Items) || sel >= len(m.topics.ChipBounds) {
		return ""
	}
	b := m.topics.ChipBounds[sel]
	if b.Width <= m.topics.VP.Width {
		return ""
	}
	focused := m.ui.focusOrder[m.ui.focusIndex] == idTopics
	return ui.RenderTooltip(m.topics.Items[sel].Name, b.XPos, b.YPos, focused)
}

// viewClient renders the main client view.
func (m *model) viewClient() string {
	m.ui.elemPos = map[string]int{}
	statusLine := m.clientInfoLine()
	topicsBox, topicBox, bounds := m.renderTopicsSection()
	messageBox := m.message.View()
	messagesBox := m.renderHistorySection()

	content := lipgloss.JoinVertical(lipgloss.Left, topicsBox, topicBox, messageBox, messagesBox)

	y := 1
	m.ui.elemPos[idTopics] = y
	y += lipgloss.Height(topicsBox)
	m.ui.elemPos[idTopic] = y
	y += lipgloss.Height(topicBox)
	m.ui.elemPos[idMessage] = y
	y += lipgloss.Height(messageBox)
	m.ui.elemPos[idHistory] = y

	startX := 2
	startY := m.ui.elemPos[idTopics] + 1
	m.topics.ChipBounds = make([]topics.ChipBound, len(bounds))
	for i, b := range bounds {
		m.topics.ChipBounds[i] = topics.ChipBound{XPos: startX + b.XPos, YPos: startY + b.YPos, Width: b.Width, Height: b.Height}
	}

	box := lipgloss.NewStyle().Width(m.ui.width).Padding(0, 1, 1, 1).Render(content)
	m.ui.viewport.SetContent(box)
	m.ui.viewport.Width = m.ui.width
	// Deduct two lines for the info header.
	m.ui.viewport.Height = m.ui.height - 2

	view := m.ui.viewport.View()
	tip := m.topicTooltip()
	contentView := lipgloss.JoinVertical(lipgloss.Left, statusLine, view)
	if tip != "" {
		contentView = lipgloss.JoinVertical(lipgloss.Left, contentView, tip)
	}
	return m.overlayHelp(contentView)
}
