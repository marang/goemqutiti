package emqutiti

import "github.com/charmbracelet/lipgloss"

// viewClient renders the main client view.
func (m *model) viewClient() string {
	m.ui.elemPos = map[string]int{}
	infoLine := m.clientInfoLine()
	topicsBox, topicBox, bounds := m.clientTopicsSection()
	messageBox := m.clientMessageSection()
	messagesBox := m.clientHistorySection()

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
	m.topics.chipBounds = make([]chipBound, len(bounds))
	for i, b := range bounds {
		m.topics.chipBounds[i] = chipBound{xPos: startX + b.xPos, yPos: startY + b.yPos, width: b.width, height: b.height}
	}

	box := lipgloss.NewStyle().Width(m.ui.width).Padding(0, 1, 1, 1).Render(content)
	m.ui.viewport.SetContent(box)
	m.ui.viewport.Width = m.ui.width
	// Deduct two lines for the info header rendered above the viewport.
	m.ui.viewport.Height = m.ui.height - 2

	view := m.ui.viewport.View()
	return m.overlayHelp(lipgloss.JoinVertical(lipgloss.Left, infoLine, view))
}
