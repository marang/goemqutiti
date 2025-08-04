package emqutiti

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/marang/emqutiti/topics"
	"github.com/marang/emqutiti/ui"
)

// clientInfoLine renders the connection status and keyboard shortcuts.
func (m *model) clientInfoLine() string {
	infoShortcuts := ui.InfoStyle.Render("Switch views: Ctrl+B brokers, Ctrl+T topics, Ctrl+P payloads, Ctrl+R traces.")
	clientID := ""
	if m.mqttClient != nil {
		r := m.mqttClient.Client.OptionsReader()
		clientID = r.ClientID()
	}
	status := strings.TrimSpace(m.connections.Connection + " " + clientID)
	st := ui.ConnStyle
	if strings.HasPrefix(m.connections.Connection, "Connected") {
		st = st.Foreground(ui.ColGreen)
	} else if strings.HasPrefix(m.connections.Connection, "Connection lost") || strings.HasPrefix(m.connections.Connection, "Failed") {
		st = st.Foreground(ui.ColWarn)
	}
	connLine := st.Render(status)
	return lipgloss.JoinVertical(lipgloss.Left, infoShortcuts, connLine)
}

// viewClient renders the main client view.
func (m *model) viewClient() string {
	m.ui.elemPos = map[string]int{}
	infoLine := m.clientInfoLine()
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
	// Deduct two lines for the info header and one for the footer note.
	m.ui.viewport.Height = m.ui.height - 3

	view := m.ui.viewport.View()
	publishNote := ui.ConnStyle.Copy().Foreground(ui.ColDarkGray).Render("Topics with blue background will be published")
	return m.overlayHelp(lipgloss.JoinVertical(lipgloss.Left, infoLine, view, publishNote))
}
