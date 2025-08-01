package main

import (
	"time"

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
			cmd := m.setMode(m.currentMode())
			return m, cmd
		case "enter":
			q := m.history.filterForm.query()
			topics, start, end, _ := parseHistoryQuery(q)
			if start.IsZero() && end.IsZero() {
				end = time.Now()
				start = end.Add(-time.Hour)
			}
			var msgs []Message
			if m.history.showArchived {
				msgs = m.history.store.SearchArchived(topics, start, end, "")
			} else {
				msgs = m.history.store.Search(topics, start, end, "")
			}
			items := make([]list.Item, len(msgs))
			for i, mmsg := range msgs {
				items[i] = historyItem{timestamp: mmsg.Timestamp, topic: mmsg.Topic, payload: mmsg.Payload, kind: mmsg.Kind, archived: mmsg.Archived}
			}
			m.history.list.SetItems(items)
			m.history.list.FilterInput.SetValue(q)
			m.history.list.SetFilterState(list.FilterApplied)
			m.history.filterForm = nil
			cmd := m.setMode(m.previousMode())
			return m, cmd
		}
	}
	f, cmd := m.history.filterForm.Update(msg)
	m.history.filterForm = &f
	return m, cmd
}
