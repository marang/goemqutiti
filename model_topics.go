package main

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/list"
)

// hasTopic reports whether the given topic already exists in the list.
func (m *model) hasTopic(topic string) bool {
	for _, t := range m.topics.items {
		if t.title == topic {
			return true
		}
	}
	return false
}

// sortTopics orders the topic list with active topics first and keeps selection.
func (m *model) sortTopics() {
	if len(m.topics.items) == 0 {
		m.topics.enabled.SetItems([]list.Item{})
		m.topics.disabled.SetItems([]list.Item{})
		return
	}
	sel := ""
	if m.topics.selected >= 0 && m.topics.selected < len(m.topics.items) {
		sel = m.topics.items[m.topics.selected].title
	}
	enabledSlice := make([]topicItem, 0, len(m.topics.items))
	disabledSlice := make([]topicItem, 0, len(m.topics.items))
	for _, t := range m.topics.items {
		if t.active {
			enabledSlice = append(enabledSlice, t)
		} else {
			disabledSlice = append(disabledSlice, t)
		}
	}
	sort.Slice(enabledSlice, func(i, j int) bool {
		return enabledSlice[i].title < enabledSlice[j].title
	})
	sort.Slice(disabledSlice, func(i, j int) bool {
		return disabledSlice[i].title < disabledSlice[j].title
	})
	m.topics.items = append(enabledSlice, disabledSlice...)
	if sel != "" {
		for i, t := range m.topics.items {
			if t.title == sel {
				m.topics.selected = i
				break
			}
		}
	}
	if m.currentMode() == modeTopics {
		enPage := m.topics.enabled.Paginator.Page
		disPage := m.topics.disabled.Paginator.Page
		var enItems []list.Item
		var disItems []list.Item
		for _, t := range m.topics.items {
			if t.active {
				enItems = append(enItems, t)
			} else {
				disItems = append(disItems, t)
			}
		}
		m.topics.enabled.SetItems(enItems)
		if enPage >= m.topics.enabled.Paginator.TotalPages {
			enPage = m.topics.enabled.Paginator.TotalPages - 1
		}
		if enPage < 0 {
			enPage = 0
		}
		m.topics.enabled.Paginator.Page = enPage

		m.topics.disabled.SetItems(disItems)
		if disPage >= m.topics.disabled.Paginator.TotalPages {
			disPage = m.topics.disabled.Paginator.TotalPages - 1
		}
		if disPage < 0 {
			disPage = 0
		}
		m.topics.disabled.Paginator.Page = disPage
	}
}

// toggleTopic toggles the subscription state of the topic at index.
func (m *model) toggleTopic(index int) {
	if index < 0 || index >= len(m.topics.items) {
		return
	}
	t := &m.topics.items[index]
	t.active = !t.active
	if m.mqttClient != nil {
		if t.active {
			m.mqttClient.Subscribe(t.title, 0, nil)
			m.appendHistory(t.title, "", "log", fmt.Sprintf("Subscribed to topic: %s", t.title))
		} else {
			m.mqttClient.Unsubscribe(t.title)
			m.appendHistory(t.title, "", "log", fmt.Sprintf("Unsubscribed from topic: %s", t.title))
		}
	}
	m.sortTopics()
}

// removeTopic unsubscribes and deletes the topic at index from the list.
func (m *model) removeTopic(index int) {
	if index < 0 || index >= len(m.topics.items) {
		return
	}
	topic := m.topics.items[index]
	if m.mqttClient != nil {
		m.mqttClient.Unsubscribe(topic.title)
		m.appendHistory(topic.title, "", "log", fmt.Sprintf("Unsubscribed from topic: %s", topic.title))
	}
	m.topics.items = append(m.topics.items[:index], m.topics.items[index+1:]...)
	if len(m.topics.items) == 0 {
		m.topics.selected = -1
	} else if m.topics.selected >= len(m.topics.items) {
		m.topics.selected = len(m.topics.items) - 1
	}
	m.sortTopics()
}
