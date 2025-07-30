package main

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// topicAtPosition returns the index of the topic chip located at the
// provided coordinates, or -1 if none exists.

func (m *model) topicAtPosition(x, y int) int {
	for i, b := range m.topics.chipBounds {
		if x >= b.xPos && x < b.xPos+b.width && y >= b.yPos && y < b.yPos+b.height {
			return i
		}
	}
	return -1
}

// historyIndexAt converts a Y coordinate into an index within the history list.
func (m *model) historyIndexAt(y int) int {
	rel := y - (m.ui.elemPos[idHistory] + 1) + m.ui.viewport.YOffset
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

// startConfirm displays a confirmation dialog and runs the action on accept.
func (m *model) startConfirm(prompt, info string, action func()) {
	m.confirmPrompt = prompt
	m.confirmInfo = info
	m.confirmAction = action
	m.ui.prevMode = m.ui.mode
	_ = m.setMode(modeConfirmDelete)
}

// subscribeActiveTopics subscribes the MQTT client to all currently active topics.
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

// refreshConnectionItems rebuilds the connections list to show status information.
func (m *model) refreshConnectionItems() {
	items := []list.Item{}
	for _, p := range m.connections.manager.Profiles {
		status := m.connections.manager.Statuses[p.Name]
		detail := m.connections.manager.Errors[p.Name]
		items = append(items, connectionItem{title: p.Name, status: status, detail: detail})
	}
	m.connections.manager.ConnectionsList.SetItems(items)
}

// setMode updates the current mode and focus order.
func (m *model) setMode(mode appMode) tea.Cmd {
	if len(m.ui.focusOrder) > 0 {
		if f, ok := m.focusMap[m.ui.focusOrder[m.ui.focusIndex]]; ok && f != nil {
			f.Blur()
		}
	}
	m.ui.mode = mode
	order, ok := focusByMode[mode]
	if !ok || len(order) == 0 {
		order = []string{idHelp}
	}
	m.ui.focusOrder = append([]string(nil), order...)
	m.ui.focusIndex = 0
	if f, ok := m.focusMap[m.ui.focusOrder[0]]; ok && f != nil {
		return f.Focus()
	}
	return nil
}
