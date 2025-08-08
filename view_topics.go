package emqutiti

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/marang/emqutiti/topics"
	"github.com/marang/emqutiti/ui"
)

// renderTopicChips builds styled topic chips and applies selection state.
func renderTopicChips(items []topics.Item, selected int) []string {
	var chips []string
	for i, t := range items {
		st := ui.ChipInactive
		switch {
		case t.Publish && t.Subscribed:
			st = ui.ChipPublish
			if i == selected {
				st = ui.ChipPublishFocused
			}
		case t.Subscribed:
			st = ui.Chip
			if i == selected {
				st = ui.ChipFocused
			}
		default:
			if i == selected {
				st = ui.ChipInactiveFocused
			}
		}
		chips = append(chips, st.Render(t.Name))
	}
	return chips
}

// layoutTopicViewport sets up the topic viewport and returns visible chip bounds.
func (m *model) layoutTopicViewport(chips []string) (string, []topics.ChipBound, int, int, float64) {
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
	stateInfo := ui.InfoSubtleStyle.Render("blue=sub  fill=pub  gray=off  pink=sel")
	keyInfo := ui.InfoSubtleStyle.Render("[←/→] move  [enter] toggle  [p] pub  [del] del")
	chipContent = lipgloss.JoinVertical(lipgloss.Left, chipContent, stateInfo, keyInfo)
	infoHeight := 2
	visible := []topics.ChipBound{}
	for _, b := range bounds {
		if b.YPos >= startLine && b.YPos < endLine {
			b.YPos -= startLine
			visible = append(visible, b)
		}
	}
	bounds = visible
	return chipContent, bounds, topicsBoxHeight, infoHeight, topicsSP
}

// buildTopicBoxes assembles the legend boxes for topics and the input field.
func (m *model) buildTopicBoxes(content string, boxHeight, infoHeight int, scrollPercent float64) (string, string) {
	active := 0
	for _, t := range m.topics.Items {
		if t.Subscribed {
			active++
		}
	}
	label := fmt.Sprintf("Topics %d/%d", active, len(m.topics.Items))
	topicsFocused := m.ui.focusOrder[m.ui.focusIndex] == idTopics
	scroll := scrollPercent
	if scroll >= 0 {
		scroll = scroll * float64(boxHeight-1) / float64(boxHeight+infoHeight-1)
	}
	topicsBox := ui.LegendBox(content, label, m.ui.width-2, boxHeight+infoHeight, ui.ColBlue, topicsFocused, scroll)

	topicFocused := m.ui.focusOrder[m.ui.focusIndex] == idTopic
	topicBox := ui.LegendBox(m.topics.Input.View(), "Topic", m.ui.width-2, 0, ui.ColBlue, topicFocused, -1)
	return topicsBox, topicBox
}

// renderTopicsSection renders topics and topic input boxes.
func (m *model) renderTopicsSection() (string, string, []topics.ChipBound) {
	chips := renderTopicChips(m.topics.Items, m.topics.Selected())
	content, bounds, boxHeight, infoHeight, scroll := m.layoutTopicViewport(chips)
	topicsBox, topicBox := m.buildTopicBoxes(content, boxHeight, infoHeight, scroll)
	return topicsBox, topicBox, bounds
}
