package emqutiti

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

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

// clientTopicsSection renders topics and topic input boxes.
func (m *model) clientTopicsSection() (string, string, []topics.ChipBound) {
	var chips []string
	for i, t := range m.topics.Items {
		st := ui.ChipStyle
		if !t.Subscribed {
			st = ui.ChipInactive
		}
		if m.ui.focusOrder[m.ui.focusIndex] == idTopics && i == m.topics.Selected() {
			st = st.BorderForeground(ui.ColPurple)
		}
		chips = append(chips, st.Render(t.Name))
	}

	chipRows, bounds := topics.LayoutChips(chips, m.ui.width-4)
	rowH := lipgloss.Height(ui.ChipStyle.Render("test"))
	maxRows := m.layout.topics.height
	if maxRows <= 0 {
		maxRows = 1
	}
	topicsBoxHeight := maxRows * rowH
	m.topics.VP.Width = m.ui.width - 4
	m.topics.VP.Height = topicsBoxHeight
	m.topics.VP.SetContent(strings.Join(chipRows, "\n"))
	m.topics.EnsureVisible(m.ui.width - 4)
	startLine := m.topics.VP.YOffset
	endLine := startLine + topicsBoxHeight
	topicsSP := -1.0
	if len(chipRows)*rowH > topicsBoxHeight {
		topicsSP = m.topics.VP.ScrollPercent()
	}
	chipContent := m.topics.VP.View()
	visible := []topics.ChipBound{}
	for _, b := range bounds {
		if b.YPos >= startLine && b.YPos < endLine {
			b.YPos -= startLine
			visible = append(visible, b)
		}
	}
	bounds = visible

	active := 0
	for _, t := range m.topics.Items {
		if t.Subscribed {
			active++
		}
	}
	label := fmt.Sprintf("Topics %d/%d", active, len(m.topics.Items))
	topicsFocused := m.ui.focusOrder[m.ui.focusIndex] == idTopics
	topicsBox := ui.LegendBox(chipContent, label, m.ui.width-2, topicsBoxHeight, ui.ColBlue, topicsFocused, topicsSP)

	topicFocused := m.ui.focusOrder[m.ui.focusIndex] == idTopic
	topicBox := ui.LegendBox(m.topics.Input.View(), "Topic", m.ui.width-2, 0, ui.ColBlue, topicFocused, -1)

	return topicsBox, topicBox, bounds
}

// clientHistorySection renders the history list box.
func (m *model) clientHistorySection() string {
	per := m.history.List().Paginator.PerPage
	totalItems := len(m.history.List().Items())
	histSP := -1.0
	if totalItems > per {
		start := m.history.List().Paginator.Page * per
		denom := totalItems - per
		if denom > 0 {
			histSP = float64(start) / float64(denom)
		}
	}

	total := len(m.history.Items())
	if st := m.history.Store(); st != nil {
		total = st.Count(m.history.ShowArchived())
	}
	shown := len(m.history.Items())
	histLabel := fmt.Sprintf("History (%d messages \u2013 Ctrl+C copy)", total)
	if m.history.FilterQuery() != "" && shown != total {
		histLabel = fmt.Sprintf("History (%d/%d messages \u2013 Ctrl+C copy)", shown, total)
	}
	listHeight := m.layout.history.height
	if m.history.FilterQuery() != "" && listHeight > 0 {
		listHeight--
	}
	m.history.List().SetSize(m.ui.width-4, listHeight)
	histContent := m.history.List().View()
	if m.history.FilterQuery() != "" {
		inner := m.ui.width - 4
		filterLine := fmt.Sprintf("Filters: %s", m.history.FilterQuery())
		filterLine = ansi.Truncate(filterLine, inner, "")
		histContent = fmt.Sprintf("%s\n%s", filterLine, histContent)
	}
	historyFocused := m.ui.focusOrder[m.ui.focusIndex] == idHistory
	return ui.LegendBox(histContent, histLabel, m.ui.width-2, m.layout.history.height, ui.ColGreen, historyFocused, histSP)
}
