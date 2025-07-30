package main

import "github.com/charmbracelet/bubbles/list"

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
	m.setMode(modeConfirmDelete)
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
func (m *model) setMode(mode appMode) {
	if m.ui.mode == modeHelp && mode != modeHelp {
		m.help.Blur()
	}
	m.ui.mode = mode
	m.ui.focusIndex = 0
	if mode == modeHelp {
		m.ui.focusOrder = []string{idHelp}
		m.help.Focus()
	} else {
		m.ui.focusOrder = defaultFocusOrder
	}
}

// computeFocusOrder filters the default focus list based on elements present in
// the current view.
func (m *model) computeFocusOrder() {
	m.ui.focusOrder = m.ui.focusOrder[:0]
	for _, id := range defaultFocusOrder {
		if _, ok := m.ui.elemPos[id]; ok {
			m.ui.focusOrder = append(m.ui.focusOrder, id)
		}
	}
	if len(m.ui.focusOrder) == 0 {
		m.ui.focusOrder = []string{idHelp}
	}
	if m.ui.focusIndex >= len(m.ui.focusOrder) {
		m.ui.focusIndex = 0
	}
}
