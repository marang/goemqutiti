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

// topicTooltip returns tooltip content and coordinates when the selected topic
// chip is truncated.
func (m *model) topicTooltip() (string, int, int) {
	sel := m.topics.Selected()
	if sel < 0 || sel >= len(m.topics.Items) || sel >= len(m.topics.ChipBounds) {
		return "", 0, 0
	}
	b := m.topics.ChipBounds[sel]
	if !b.Truncated {
		return "", 0, 0
	}
	style := ui.TooltipStyle
	if m.ui.focusOrder[m.ui.focusIndex] == idTopics {
		style = ui.TooltipFocused
	}
	tip := ui.Tooltip{Text: m.topics.Items[sel].Name, Width: lipgloss.Width(m.topics.Items[sel].Name), Style: &style}.View(0, 0)
	return tip, b.XPos, b.YPos
}

// overlayAt draws top onto base at coordinates x, y.
func overlayAt(base, top string, x, y int) string {
	baseLines := strings.Split(base, "\n")
	topLines := strings.Split(top, "\n")
	for i, tl := range topLines {
		by := y + i
		if by >= len(baseLines) {
			break
		}
		line := []rune(baseLines[by])
		tip := []rune(strings.Repeat(" ", x) + tl)
		if len(line) < len(tip) {
			line = append(line, make([]rune, len(tip)-len(line))...)
		}
		for j, r := range tip {
			if r != ' ' {
				line[j] = r
			}
		}
		baseLines[by] = string(line)
	}
	return strings.Join(baseLines, "\n")
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
		m.topics.ChipBounds[i] = topics.ChipBound{
			XPos:      startX + b.XPos,
			YPos:      startY + b.YPos,
			Width:     b.Width,
			Height:    b.Height,
			Truncated: b.Truncated,
		}
	}

	box := lipgloss.NewStyle().Width(m.ui.width).Padding(0, 1, 1, 1).Render(content)
	m.ui.viewport.SetContent(box)
	m.ui.viewport.Width = m.ui.width
	// Deduct two lines for the info header.
	m.ui.viewport.Height = m.ui.height - 2

	view := m.ui.viewport.View()
	tip, tx, ty := m.topicTooltip()
	contentView := lipgloss.JoinVertical(lipgloss.Left, statusLine, view)
	finalView := m.overlayHelp(contentView)
	if tip != "" {
		finalView = overlayAt(finalView, tip, tx, ty+1)
	}
	return finalView
}
