package emqutiti

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/marang/emqutiti/topics"
	"github.com/marang/emqutiti/ui"
)

// renderTopicsSection renders topics and topic input boxes.
func (m *model) renderTopicsSection() (string, string, []topics.ChipBound) {
	var chips []string
	for i, t := range m.topics.Items {
		st := ui.ChipInactive
		switch {
		case t.Publish && t.Subscribed:
			st = ui.ChipPublish
			if i == m.topics.Selected() {
				st = ui.ChipPublishFocused
			}
		case t.Subscribed:
			st = ui.Chip
			if i == m.topics.Selected() {
				st = ui.ChipFocused
			}
		default:
			if i == m.topics.Selected() {
				st = ui.ChipInactiveFocused
			}
		}
		// if i == m.topics.Selected() {
		// 	st = ui.ChipStyleFocused // st.BorderForeground(ui.ColPurple)
		// }
		chips = append(chips, st.Render(t.Name))
	}

	chipRows, bounds := topics.LayoutChips(chips, m.ui.width-4)
	rowH := lipgloss.Height(ui.Chip.Render("test"))
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
