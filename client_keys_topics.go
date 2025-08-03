package emqutiti

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// handleLeftKey moves topic selection left.
func (m *model) handleLeftKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idTopics && len(m.topics.items) > 0 {
		sel := m.topics.Selected()
		m.topics.SetSelected((sel - 1 + len(m.topics.items)) % len(m.topics.items))
		m.topics.EnsureVisible(m.ui.width - 4)
	}
	return nil
}

// handleRightKey moves topic selection right.
func (m *model) handleRightKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idTopics && len(m.topics.items) > 0 {
		sel := m.topics.Selected()
		m.topics.SetSelected((sel + 1) % len(m.topics.items))
		m.topics.EnsureVisible(m.ui.width - 4)
	}
	return nil
}

// handleTopicScroll handles scroll keys when topics are focused.
func (m *model) handleTopicScroll(key string) tea.Cmd {
	delta := -1
	if key == "down" || key == "j" {
		delta = 1
	}
	m.topics.Scroll(delta)
	return nil
}

// handleEnterKey handles Enter for topic input, toggling, and history detail.
func (m *model) handleEnterKey() tea.Cmd {
	switch m.ui.focusOrder[m.ui.focusIndex] {
	case idTopic:
		topic := strings.TrimSpace(m.topics.input.Value())
		if topic != "" && !m.topics.HasTopic(topic) {
			m.topics.items = append(m.topics.items, topicItem{title: topic, subscribed: true})
			m.topics.SortTopics()
			if m.currentMode() == modeTopics {
				m.topics.RebuildActiveTopicList()
			}
			m.topics.input.SetValue("")
			return func() tea.Msg { return topicToggleMsg{topic: topic, subscribed: true} }
		}
	case idTopics:
		sel := m.topics.Selected()
		if sel >= 0 && sel < len(m.topics.items) {
			cmd := m.topics.ToggleTopic(sel)
			m.topics.EnsureVisible(m.ui.width - 4)
			return cmd
		}
	case idHistory:
		return m.handleHistoryViewKey()
	}
	return nil
}

// handleDeleteTopicKey deletes the selected topic.
func (m *model) handleDeleteTopicKey() tea.Cmd {
	idx := m.topics.Selected()
	name := m.topics.items[idx].title
	rf := func() tea.Cmd { return m.setFocus(m.ui.focusOrder[m.ui.focusIndex]) }
	m.startConfirm(fmt.Sprintf("Delete topic '%s'? [y/n]", name), "", rf, func() tea.Cmd {
		cmd := m.topics.RemoveTopic(idx)
		if m.currentMode() == modeTopics {
			m.topics.RebuildActiveTopicList()
		}
		return cmd
	}, nil)
	return nil
}
