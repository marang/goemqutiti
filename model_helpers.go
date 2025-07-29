package main

import "github.com/charmbracelet/bubbles/list"

func (m *model) topicAtPosition(x, y int) int {
	for i, b := range m.topics.chipBounds {
		if x >= b.xPos && x < b.xPos+b.width && y >= b.yPos && y < b.yPos+b.height {
			return i
		}
	}
	return -1
}

func (m *model) historyIndexAt(y int) int {
	rel := y - (m.ui.elemPos["history"] + 1) + m.ui.viewport.YOffset
	if rel < 0 {
		return -1
	}
	h := 2 // historyDelegate height
	idx := rel / h
	start := m.history.list.Paginator.Page * m.history.list.Paginator.PerPage
	i := start + idx
	if i >= len(m.history.list.Items()) || i < 0 {
		return -1
	}
	return i
}

func (m *model) startConfirm(prompt string, action func()) {
	m.confirmPrompt = prompt
	m.confirmAction = action
	m.ui.prevMode = m.ui.mode
	m.ui.mode = modeConfirmDelete
}

func (m *model) subscribeActiveTopics() {
	if m.mqttClient == nil {
		return
	}
	for _, t := range m.topics.items {
		if t.active {
			m.mqttClient.Subscribe(t.title, 0, nil)
		}
	}
}

func (m *model) refreshConnectionItems() {
	items := []list.Item{}
	for _, p := range m.connections.manager.Profiles {
		status := m.connections.manager.Statuses[p.Name]
		detail := m.connections.manager.Errors[p.Name]
		items = append(items, connectionItem{title: p.Name, status: status, detail: detail})
	}
	m.connections.manager.ConnectionsList.SetItems(items)
}
