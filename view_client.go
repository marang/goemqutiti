package emqutiti

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
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

// selectedTopicInfo renders the full topic name when the focused chip is
// truncated.
func (m *model) selectedTopicInfo() string {
	sel := m.topics.Selected()
	if sel < 0 || sel >= len(m.topics.Items) || sel >= len(m.topics.ChipBounds) {
		return ""
	}
	if !m.topics.ChipBounds[sel].Truncated {
		return ""
	}
	w := m.ui.width - 2
	if w < 0 {
		w = 0
	}
	name := ansi.Wrap(m.topics.Items[sel].Name, w, " ")
	return ui.InfoStyle.Render(name)
}

// viewClient renders the main client view.
func (m *model) viewClient() string {
	m.ui.elemPos = map[string]int{}
	statusLine := m.clientInfoLine()
	topicsBox, topicBox, bounds := m.renderTopicsSection()
	messageBox := m.message.View()
	messagesBox := m.renderHistorySection()

	m.topics.ChipBounds = make([]topics.ChipBound, len(bounds))
	for i, b := range bounds {
		m.topics.ChipBounds[i] = topics.ChipBound{
			XPos:      b.XPos,
			YPos:      b.YPos,
			Width:     b.Width,
			Height:    b.Height,
			Truncated: b.Truncated,
		}
	}

	infoLine := m.selectedTopicInfo()
	var content string
	y := 1
	if infoLine != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, infoLine, topicsBox, topicBox, messageBox, messagesBox)
		y += lipgloss.Height(infoLine)
	} else {
		content = lipgloss.JoinVertical(lipgloss.Left, topicsBox, topicBox, messageBox, messagesBox)
	}

	m.ui.elemPos[idTopics] = y
	y += lipgloss.Height(topicsBox)
	m.ui.elemPos[idTopic] = y
	y += lipgloss.Height(topicBox)
	m.ui.elemPos[idMessage] = y
	y += lipgloss.Height(messageBox)
	m.ui.elemPos[idHistory] = y

	startX := 2
	startY := m.ui.elemPos[idTopics] + 1
	for i := range m.topics.ChipBounds {
		m.topics.ChipBounds[i].XPos += startX
		m.topics.ChipBounds[i].YPos += startY
	}

	box := lipgloss.NewStyle().Width(m.ui.width).Padding(0, 1, 1, 1).Render(content)
	m.ui.viewport.SetContent(box)
	m.ui.viewport.Width = m.ui.width
	// Deduct two lines for the info header.
	m.ui.viewport.Height = m.ui.height - 2

	view := m.ui.viewport.View()
	contentView := lipgloss.JoinVertical(lipgloss.Left, statusLine, view)
	return m.overlayHelp(contentView)
}
