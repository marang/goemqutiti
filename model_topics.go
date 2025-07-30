package main

import (
	"fmt"
	"sort"
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
		return
	}
	sel := ""
	if m.topics.selected >= 0 && m.topics.selected < len(m.topics.items) {
		sel = m.topics.items[m.topics.selected].title
	}
	sort.SliceStable(m.topics.items, func(i, j int) bool {
		if m.topics.items[i].active != m.topics.items[j].active {
			return m.topics.items[i].active && !m.topics.items[j].active
		}
		return m.topics.items[i].title < m.topics.items[j].title
	})
	if sel != "" {
		for i, t := range m.topics.items {
			if t.title == sel {
				m.topics.selected = i
				break
			}
		}
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
