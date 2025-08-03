package emqutiti

import (
	"time"

	"github.com/charmbracelet/bubbles/list"
)

// saveCurrent persists topics and payloads for the active connection.
func (m *model) saveCurrent() {
	if m.connections.active == "" {
		return
	}
	m.connections.saved[m.connections.active] = connectionData{Topics: m.topics.items, Payloads: m.payloads.Items()}
	saveState(m.connections.saved)
}

// restoreState loads saved state for the named connection.
func (m *model) restoreState(name string) {
	if data, ok := m.connections.saved[name]; ok {
		m.topics.items = data.Topics
		m.payloads.SetItems(data.Payloads)
		m.sortTopics()
		m.rebuildActiveTopicList()
	} else {
		m.topics.items = []topicItem{}
		m.payloads.Clear()
	}
}

// appendHistory stores a message in the history list and optional store.
func (m *model) appendHistory(topic, payload, kind, logText string) {
	ts := time.Now()
	text := payload
	if kind == "log" {
		text = logText
	}
	hi := historyItem{timestamp: ts, topic: topic, payload: text, kind: kind, archived: false}
	if m.history.store != nil {
		m.history.store.Append(Message{Timestamp: ts, Topic: topic, Payload: payload, Kind: kind, Archived: false})
	}
	if !m.history.showArchived {
		if m.history.filterQuery != "" {
			var items []list.Item
			m.history.items, items = applyHistoryFilter(m.history.filterQuery, m.history.store, m.history.showArchived)
			m.history.list.SetItems(items)
			m.history.list.Select(len(items) - 1)
		} else {
			m.history.items = append(m.history.items, hi)
			items := make([]list.Item, len(m.history.items))
			for i, it := range m.history.items {
				items[i] = it
			}
			m.history.list.SetItems(items)
			m.history.list.Select(len(items) - 1)
		}
	}
}
