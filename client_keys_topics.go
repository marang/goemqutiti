package emqutiti

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// handleLeftKey moves topic selection left.
func (m *model) handleLeftKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idTopics && len(m.topics.items) > 0 {
		m.topics.selected = (m.topics.selected - 1 + len(m.topics.items)) % len(m.topics.items)
		m.ensureTopicVisible()
	}
	return nil
}

// handleRightKey moves topic selection right.
func (m *model) handleRightKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idTopics && len(m.topics.items) > 0 {
		m.topics.selected = (m.topics.selected + 1) % len(m.topics.items)
		m.ensureTopicVisible()
	}
	return nil
}

// handleTopicScroll handles scroll keys when topics are focused.
func (m *model) handleTopicScroll(key string) tea.Cmd {
	delta := -1
	if key == "down" || key == "j" {
		delta = 1
	}
	m.scrollTopics(delta)
	return nil
}

// handleEnterKey handles Enter for topic input, toggling, and history detail.
func (m *model) handleEnterKey() tea.Cmd {
	switch m.ui.focusOrder[m.ui.focusIndex] {
	case idTopic:
		topic := strings.TrimSpace(m.topics.input.Value())
		if topic != "" && !m.hasTopic(topic) {
			m.topics.items = append(m.topics.items, topicItem{title: topic, subscribed: true})
			m.sortTopics()
			if m.currentMode() == modeTopics {
				m.rebuildActiveTopicList()
			}
			if m.mqttClient != nil {
				m.mqttClient.Subscribe(topic, 0, nil)
			}
			m.appendHistory(topic, "", "log", fmt.Sprintf("Subscribed to topic: %s", topic))
			m.topics.input.SetValue("")
		}
	case idTopics:
		if m.topics.selected >= 0 && m.topics.selected < len(m.topics.items) {
			m.toggleTopic(m.topics.selected)
			m.ensureTopicVisible()
			if m.currentMode() == modeTopics {
				m.rebuildActiveTopicList()
			}
		}
	case idHistory:
		return m.handleHistoryViewKey()
	}
	return nil
}

// handleDeleteTopicKey deletes the selected topic.
func (m *model) handleDeleteTopicKey() tea.Cmd {
	idx := m.topics.selected
	name := m.topics.items[idx].title
	m.confirm.returnFocus = m.ui.focusOrder[m.ui.focusIndex]
	m.startConfirm(fmt.Sprintf("Delete topic '%s'? [y/n]", name), "", func() {
		m.removeTopic(idx)
		if m.currentMode() == modeTopics {
			m.rebuildActiveTopicList()
		}
	})
	return nil
}
