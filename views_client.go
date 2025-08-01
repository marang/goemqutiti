package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/marang/goemqutiti/ui"
)

// clientInfoLine renders the connection status and keyboard shortcuts.
func (m *model) clientInfoLine() string {
	infoShortcuts := ui.InfoStyle.Render("Switch views: Ctrl+B brokers, Ctrl+T topics, Ctrl+P payloads, Ctrl+R traces.")
	clientID := ""
	if m.mqttClient != nil {
		r := m.mqttClient.Client.OptionsReader()
		clientID = r.ClientID()
	}
	status := strings.TrimSpace(m.connections.connection + " " + clientID)
	st := ui.ConnStyle
	if strings.HasPrefix(m.connections.connection, "Connected") {
		st = st.Foreground(ui.ColGreen)
	} else if strings.HasPrefix(m.connections.connection, "Connection lost") || strings.HasPrefix(m.connections.connection, "Failed") {
		st = st.Foreground(ui.ColWarn)
	}
	connLine := st.Render(status)
	return lipgloss.JoinVertical(lipgloss.Left, infoShortcuts, connLine)
}

// clientTopicsSection renders topics and topic input boxes.
func (m *model) clientTopicsSection() (string, string, []chipBound) {
	var chips []string
	for i, t := range m.topics.items {
		st := ui.ChipStyle
		if !t.subscribed {
			st = ui.ChipInactive
		}
		if m.ui.focusOrder[m.ui.focusIndex] == idTopics && i == m.topics.selected {
			st = st.BorderForeground(ui.ColPurple)
		}
		chips = append(chips, st.Render(t.title))
	}

	chipRows, bounds := layoutChips(chips, m.ui.width-4)
	rowH := lipgloss.Height(ui.ChipStyle.Render("test"))
	maxRows := m.layout.topics.height
	if maxRows <= 0 {
		maxRows = 1
	}
	topicsBoxHeight := maxRows * rowH
	m.topics.vp.Width = m.ui.width - 4
	m.topics.vp.Height = topicsBoxHeight
	m.topics.vp.SetContent(strings.Join(chipRows, "\n"))
	m.ensureTopicVisible()
	startLine := m.topics.vp.YOffset
	endLine := startLine + topicsBoxHeight
	topicsSP := -1.0
	if len(chipRows)*rowH > topicsBoxHeight {
		topicsSP = m.topics.vp.ScrollPercent()
	}
	chipContent := m.topics.vp.View()
	visible := []chipBound{}
	for _, b := range bounds {
		if b.yPos >= startLine && b.yPos < endLine {
			b.yPos -= startLine
			visible = append(visible, b)
		}
	}
	bounds = visible

	active := 0
	for _, t := range m.topics.items {
		if t.subscribed {
			active++
		}
	}
	label := fmt.Sprintf("Topics %d/%d", active, len(m.topics.items))
	topicsFocused := m.ui.focusOrder[m.ui.focusIndex] == idTopics
	topicsBox := ui.LegendBox(chipContent, label, m.ui.width-2, topicsBoxHeight, ui.ColBlue, topicsFocused, topicsSP)

	topicFocused := m.ui.focusOrder[m.ui.focusIndex] == idTopic
	topicBox := ui.LegendBox(m.topics.input.View(), "Topic", m.ui.width-2, 0, ui.ColBlue, topicFocused, -1)

	return topicsBox, topicBox, bounds
}

// clientMessageSection renders the message input box.
func (m *model) clientMessageSection() string {
	msgContent := m.message.input.View()
	msgLines := m.message.input.LineCount()
	msgHeight := m.layout.message.height
	msgSP := -1.0
	if msgLines > msgHeight {
		off := m.message.input.Line() - msgHeight + 1
		if off < 0 {
			off = 0
		}
		maxOff := msgLines - msgHeight
		if off > maxOff {
			off = maxOff
		}
		if maxOff > 0 {
			msgSP = float64(off) / float64(maxOff)
		}
	}
	messageFocused := m.ui.focusOrder[m.ui.focusIndex] == idMessage
	return ui.LegendBox(msgContent, "Message (Ctrl+S publishes)", m.ui.width-2, msgHeight, ui.ColBlue, messageFocused, msgSP)
}

// clientHistorySection renders the history list box.
func (m *model) clientHistorySection() string {
	per := m.history.list.Paginator.PerPage
	totalItems := len(m.history.list.Items())
	histSP := -1.0
	if totalItems > per {
		start := m.history.list.Paginator.Page * per
		denom := totalItems - per
		if denom > 0 {
			histSP = float64(start) / float64(denom)
		}
	}

	total := len(m.history.items)
	if m.history.store != nil {
		total = m.history.store.Count(m.history.showArchived)
	}
	shown := len(m.history.items)
	histLabel := fmt.Sprintf("History (%d messages \u2013 Ctrl+C copy)", total)
	if m.history.filterQuery != "" && shown != total {
		histLabel = fmt.Sprintf("History (%d/%d messages \u2013 Ctrl+C copy)", shown, total)
	}
	listHeight := m.layout.history.height
	if m.history.filterQuery != "" && listHeight > 0 {
		listHeight--
	}
	m.history.list.SetSize(m.ui.width-4, listHeight)
	histContent := m.history.list.View()
	if m.history.filterQuery != "" {
		inner := m.ui.width - 4
		filterLine := fmt.Sprintf("Filters: %s", m.history.filterQuery)
		filterLine = ansi.Truncate(filterLine, inner, "")
		histContent = fmt.Sprintf("%s\n%s", filterLine, histContent)
	}
	historyFocused := m.ui.focusOrder[m.ui.focusIndex] == idHistory
	return ui.LegendBox(histContent, histLabel, m.ui.width-2, m.layout.history.height, ui.ColGreen, historyFocused, histSP)
}
