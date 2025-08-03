package emqutiti

import (
	tea "github.com/charmbracelet/bubbletea"
	"time"

	"github.com/marang/emqutiti/confirm"
	"github.com/marang/emqutiti/history"
)

// historyIndexAt converts a Y coordinate into an index within the history list.
func (m *model) historyIndexAt(y int) int {
	rel := y - (m.ui.elemPos[idHistory] + 1) + m.ui.viewport.YOffset
	if rel < 0 {
		return -1
	}
	h := 2 // historyDelegate height
	idx := rel / h
	lst := m.history.List()
	start := lst.Paginator.Page * lst.Paginator.PerPage
	i := start + idx
	if i >= len(lst.Items()) || i < 0 {
		return -1
	}
	return i
}

// FocusedID returns the identifier of the currently focused element.
func (m *model) FocusedID() string { return m.ui.focusOrder[m.ui.focusIndex] }

// ListenStatus proxies connection status updates for components.
func (m *model) ListenStatus() tea.Cmd { return m.connections.ListenStatus() }

// ScrollToFocused ensures the focused element is visible.
func (m *model) ScrollToFocused() { m.scrollToFocused() }

// ResetElemPos clears cached element positions.
func (m *model) ResetElemPos() { m.ui.elemPos = map[string]int{} }

// SetElemPos records the position of a UI element.
func (m *model) SetElemPos(id string, pos int) { m.ui.elemPos[id] = pos }

// SetFocus delegates focus changes to the model's focus manager.
func (m *model) SetFocus(id string) tea.Cmd { return m.setFocus(id) }

// StartConfirm displays a confirmation dialog and runs the action on accept.
func (m *model) StartConfirm(prompt, info string, returnFocus func() tea.Cmd, action func() tea.Cmd, cancel func()) {
	m.startConfirm(prompt, info, returnFocus, action, cancel)
}

// startConfirm displays a confirmation dialog and runs the action on accept.
func (m *model) startConfirm(prompt, info string, returnFocus func() tea.Cmd, action func() tea.Cmd, cancel func()) {
	m.confirm = confirm.NewComponent(m, m, returnFocus, action, cancel)
	m.confirm.Start(prompt, info)
	m.components[modeConfirmDelete] = m.confirm
}

// startHistoryFilter opens the history filter form.
func (m *model) startHistoryFilter() tea.Cmd {
	var topics []string
	for _, t := range m.topics.Items {
		topics = append(topics, t.Name)
	}
	var topic, payload string
	var start, end time.Time
	if m.history.FilterQuery() != "" {
		ts, s, e, p := parseHistoryQuery(m.history.FilterQuery())
		if len(ts) > 0 {
			topic = ts[0]
		}
		start, end, payload = s, e, p
	} else {
		end = time.Now()
		start = end.Add(-time.Hour)
	}
	hf := history.NewFilterForm(topics, topic, payload, start, end)
	m.history.SetFilterForm(&hf)
	return m.setMode(modeHistoryFilter)
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

// SetMode exposes setMode to satisfy the navigator interface.
func (m *model) SetMode(mode appMode) tea.Cmd { return m.setMode(mode) }

// PreviousMode exposes previousMode to satisfy the navigator interface.
func (m *model) PreviousMode() appMode { return m.previousMode() }

// SetConfirmMode switches to the confirmation screen.
func (m *model) SetConfirmMode() tea.Cmd { return m.setMode(modeConfirmDelete) }

// SetPreviousMode returns to the prior screen.
func (m *model) SetPreviousMode() tea.Cmd { return m.setMode(m.previousMode()) }

// CurrentMode exposes currentMode to satisfy component interfaces.
func (m *model) CurrentMode() appMode { return m.currentMode() }

// Width returns the current UI width.
func (m *model) Width() int { return m.ui.width }

// MessageHeight returns the configured message box height.
func (m *model) MessageHeight() int { return m.layout.message.height }

// Height returns the current UI height.
func (m *model) Height() int { return m.ui.height }

// SetClientMode switches to the main client screen.
func (m *model) SetClientMode() tea.Cmd { return m.setMode(modeClient) }
