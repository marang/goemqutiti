package emqutiti

import (
	"fmt"
	"log"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/emqutiti/history"
)

// handleToggleArchiveKey toggles between active and archived history.
func (m *model) handleToggleArchiveKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idHistory && m.history.store != nil {
		m.history.showArchived = !m.history.showArchived
		var msgs []history.Message
		if m.history.showArchived {
			msgs = m.history.store.Search(true, nil, time.Time{}, time.Time{}, "")
		} else {
			msgs = m.history.store.Search(false, nil, time.Time{}, time.Time{}, "")
		}
		var items []list.Item
		m.history.items, items = messagesToHistoryItems(msgs)
		m.history.list.SetItems(items)
		if len(items) > 0 {
			m.history.list.Select(len(items) - 1)
		} else {
			m.history.list.Select(-1)
		}
	}
	return nil
}

// handleArchiveKey archives selected history messages.
func (m *model) handleArchiveKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idHistory && !m.history.showArchived {
		if len(m.history.items) == 0 {
			return nil
		}
		archived := false
		for i := len(m.history.items) - 1; i >= 0; i-- {
			it := m.history.items[i]
			if it.IsSelected != nil && *it.IsSelected {
				key := fmt.Sprintf("%s/%020d", it.Topic, it.Timestamp.UnixNano())
				if m.history.store != nil {
					if err := m.history.store.Archive(key); err != nil {
						msg := fmt.Sprintf("Failed to archive message: %v", err)
						log.Println(msg)
						m.history.Append("", msg, "log", msg)
						continue
					}
				}
				m.history.items = append(m.history.items[:i], m.history.items[i+1:]...)
				archived = true
			}
		}
		if !archived {
			idx := m.history.list.Index()
			if idx >= 0 && idx < len(m.history.items) {
				it := m.history.items[idx]
				key := fmt.Sprintf("%s/%020d", it.Topic, it.Timestamp.UnixNano())
				if m.history.store != nil {
					if err := m.history.store.Archive(key); err != nil {
						msg := fmt.Sprintf("Failed to archive message: %v", err)
						log.Println(msg)
						m.history.Append("", msg, "log", msg)
					} else {
						m.history.items = append(m.history.items[:idx], m.history.items[idx+1:]...)
					}
				} else {
					m.history.items = append(m.history.items[:idx], m.history.items[idx+1:]...)
				}
			}
		}
		items := make([]list.Item, len(m.history.items))
		for i, it := range m.history.items {
			it.IsSelected = nil
			m.history.items[i] = it
			items[i] = it
		}
		m.history.list.SetItems(items)
		if len(m.history.items) == 0 {
			m.history.list.Select(-1)
		} else if m.history.list.Index() >= len(m.history.items) {
			m.history.list.Select(len(m.history.items) - 1)
		}
		m.history.selectionAnchor = -1
	}
	return nil
}

// handleDeleteHistoryKey deletes selected history messages.
func (m *model) handleDeleteHistoryKey() tea.Cmd {
	if len(m.history.items) == 0 {
		return nil
	}
	hasSelection := false
	for i := range m.history.items {
		if m.history.items[i].IsSelected != nil && *m.history.items[i].IsSelected {
			v := true
			m.history.items[i].IsMarkedForDeletion = &v
			hasSelection = true
		}
	}
	if !hasSelection {
		idx := m.history.list.Index()
		if idx >= 0 && idx < len(m.history.items) {
			v := true
			m.history.items[idx].IsMarkedForDeletion = &v
		}
	}
	rf := func() tea.Cmd { return m.setFocus(m.ui.focusOrder[m.ui.focusIndex]) }
	m.startConfirm("Delete selected messages? [y/n]", "", rf, func() tea.Cmd {
		for i := len(m.history.items) - 1; i >= 0; i-- {
			it := m.history.items[i]
			if it.IsMarkedForDeletion != nil && *it.IsMarkedForDeletion {
				key := fmt.Sprintf("%s/%020d", it.Topic, it.Timestamp.UnixNano())
				if m.history.store != nil {
					if err := m.history.store.Delete(key); err != nil {
						msg := fmt.Sprintf("Failed to delete message: %v", err)
						log.Println(msg)
						m.history.Append("", msg, "log", msg)
						continue
					}
				}
				m.history.items = append(m.history.items[:i], m.history.items[i+1:]...)
			}
		}
		items := make([]list.Item, len(m.history.items))
		for i, it := range m.history.items {
			it.IsSelected = nil
			it.IsMarkedForDeletion = nil
			m.history.items[i] = it
			items[i] = it
		}
		m.history.list.SetItems(items)
		if len(m.history.items) == 0 {
			m.history.list.Select(-1)
		} else if m.history.list.Index() >= len(m.history.items) {
			m.history.list.Select(len(m.history.items) - 1)
		}
		m.history.selectionAnchor = -1
		return nil
	}, func() {
		for i := range m.history.items {
			m.history.items[i].IsMarkedForDeletion = nil
		}
	})
	return nil
}
