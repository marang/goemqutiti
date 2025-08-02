package emqutiti

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"time"
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
	m.confirm.start(prompt, info, action)
}

// startHistoryFilter opens the history filter form.
func (m *model) startHistoryFilter() tea.Cmd {
	var topics []string
	for _, t := range m.topics.items {
		topics = append(topics, t.title)
	}
	var topic, payload string
	var start, end time.Time
	if m.history.filterQuery != "" {
		ts, s, e, p := parseHistoryQuery(m.history.filterQuery)
		if len(ts) > 0 {
			topic = ts[0]
		}
		start, end, payload = s, e, p
	} else {
		end = time.Now()
		start = end.Add(-time.Hour)
	}
	hf := newHistoryFilterForm(topics, topic, payload, start, end)
	m.history.filterForm = &hf
	return m.setMode(modeHistoryFilter)
}

// subscribeActiveTopics subscribes the MQTT client to all currently active topics.
func (m *model) subscribeActiveTopics() {
	if m.mqttClient == nil {
		return
	}
	for _, t := range m.topics.items {
		if t.subscribed {
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
	if m.focus != nil && len(m.focus.items) > 0 {
		m.focus.items[m.focus.focusIndex].Blur()
	}
	// push mode to stack
	if len(m.ui.modeStack) == 0 || m.ui.modeStack[0] != mode {
		m.ui.modeStack = append([]appMode{mode}, m.ui.modeStack...)
	} else {
		m.ui.modeStack[0] = mode
	}
	// remove any other occurrences of this mode to keep order meaningful
	for i := 1; i < len(m.ui.modeStack); {
		if m.ui.modeStack[i] == mode {
			m.ui.modeStack = append(m.ui.modeStack[:i], m.ui.modeStack[i+1:]...)
		} else {
			i++
		}
	}
	if len(m.ui.modeStack) > 10 {
		m.ui.modeStack = m.ui.modeStack[:10]
	}
	order, ok := focusByMode[mode]
	if !ok || len(order) == 0 {
		order = []string{idHelp}
	}
	m.ui.focusOrder = append([]string(nil), order...)
	items := make([]Focusable, len(order))
	for i, id := range order {
		items[i] = m.focusables[id]
	}
	m.focus = NewFocusMap(items)
	m.ui.focusIndex = m.focus.Index()
	m.help.Blur()
	return nil
}

// currentMode returns the active application mode.
func (m *model) currentMode() appMode {
	if len(m.ui.modeStack) == 0 {
		return modeClient
	}
	return m.ui.modeStack[0]
}

// previousMode returns the last mode before the current one.
func (m *model) previousMode() appMode {
	if len(m.ui.modeStack) > 1 {
		return m.ui.modeStack[1]
	}
	return m.currentMode()
}
