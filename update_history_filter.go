package main

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// updateHistoryFilter handles the history filter form interaction.
func (m model) updateHistoryFilter(msg tea.Msg) (model, tea.Cmd) {
	if m.history.filterForm == nil {
		return m, nil
	}
	switch t := msg.(type) {
	case tea.KeyMsg:
		switch t.String() {
		case "esc":
			m.history.filterForm = nil
			if len(m.ui.modeStack) > 0 {
				m.ui.modeStack = m.ui.modeStack[1:]
			}
			if len(m.ui.modeStack) > 0 && m.ui.modeStack[0] == modeHelp {
				m.ui.modeStack = m.ui.modeStack[1:]
			}
			cmd := tea.Batch(m.setMode(m.currentMode()), m.setFocus(idHistory))
			return m, cmd
		case "enter":
			q := m.history.filterForm.query()
			topics, start, end, payload := parseHistoryQuery(q)
			var msgs []Message
			if m.history.showArchived {
				msgs = m.history.store.SearchArchived(topics, start, end, payload)
			} else {
				msgs = m.history.store.Search(topics, start, end, payload)
			}
			var items []list.Item
			m.history.items, items = messagesToHistoryItems(msgs)
			m.history.list.SetItems(items)
			m.history.list.FilterInput.SetValue("")
			m.history.list.SetFilterState(list.Unfiltered)
			m.history.filterQuery = q
			m.history.filterForm = nil
			cmd := tea.Batch(m.setMode(m.previousMode()), m.setFocus(idHistory))
			return m, cmd
		}
	}
	f, cmd := m.history.filterForm.Update(msg)
	m.history.filterForm = &f
	return m, cmd
}
