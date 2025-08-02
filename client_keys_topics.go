package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *model) handleLeftKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idTopics && len(m.topics.items) > 0 {
		m.topics.selected = (m.topics.selected - 1 + len(m.topics.items)) % len(m.topics.items)
		m.ensureTopicVisible()
	}
	return nil
}

func (m *model) handleRightKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idTopics && len(m.topics.items) > 0 {
		m.topics.selected = (m.topics.selected + 1) % len(m.topics.items)
		m.ensureTopicVisible()
	}
	return nil
}

func (m *model) handleEnterKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idTopic {
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
	} else if m.ui.focusOrder[m.ui.focusIndex] == idTopics && m.topics.selected >= 0 && m.topics.selected < len(m.topics.items) {
		m.toggleTopic(m.topics.selected)
		m.ensureTopicVisible()
		if m.currentMode() == modeTopics {
			m.rebuildActiveTopicList()
		}
	}
	return nil
}

func (m *model) handleTopicsScrollKeys(msg tea.KeyMsg) tea.Cmd {
	delta := -1
	if msg.String() == "down" || msg.String() == "j" {
		delta = 1
	}
	m.scrollTopics(delta)
	return nil
}

func (m *model) handleTopicsDeleteKey() tea.Cmd {
	if m.topics.selected < 0 || m.topics.selected >= len(m.topics.items) {
		return nil
	}
	idx := m.topics.selected
	name := m.topics.items[idx].title
	m.confirmReturnFocus = m.ui.focusOrder[m.ui.focusIndex]
	m.startConfirm(fmt.Sprintf("Delete topic '%s'? [y/n]", name), "", func() {
		m.removeTopic(idx)
		if m.currentMode() == modeTopics {
			m.rebuildActiveTopicList()
		}
	})
	return nil
}
