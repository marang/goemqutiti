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
	if m.ui.focusOrder[m.ui.focusIndex] == idHistory && m.history.Store() != nil {
		m.history.SetShowArchived(!m.history.ShowArchived())
		var msgs []history.Message
		if m.history.ShowArchived() {
			msgs = m.history.Store().Search(true, nil, time.Time{}, time.Time{}, "")
		} else {
			msgs = m.history.Store().Search(false, nil, time.Time{}, time.Time{}, "")
		}
		hitems, items := history.MessagesToItems(msgs)
		m.history.SetItems(hitems)
		m.history.List().SetItems(items)
		if len(items) > 0 {
			m.history.List().Select(len(items) - 1)
		} else {
			m.history.List().Select(-1)
		}
	}
	return nil
}

// handleArchiveKey archives selected history messages.
func (m *model) handleArchiveKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idHistory && !m.history.ShowArchived() {
		hitems := m.history.Items()
		if len(hitems) == 0 {
			return nil
		}
		archived := false
		for i := len(hitems) - 1; i >= 0; i-- {
			it := hitems[i]
			if it.IsSelected != nil && *it.IsSelected {
				key := fmt.Sprintf("%s/%020d", it.Topic, it.Timestamp.UnixNano())
				if st := m.history.Store(); st != nil {
					if err := st.Archive(key); err != nil {
						msg := fmt.Sprintf("Failed to archive message: %v", err)
						log.Println(msg)
						m.history.Append("", msg, "log", false, msg)
						continue
					}
				}
				hitems = append(hitems[:i], hitems[i+1:]...)
				archived = true
			}
		}
		if !archived {
			idx := m.history.List().Index()
			if idx >= 0 && idx < len(hitems) {
				it := hitems[idx]
				key := fmt.Sprintf("%s/%020d", it.Topic, it.Timestamp.UnixNano())
				if st := m.history.Store(); st != nil {
					if err := st.Archive(key); err != nil {
						msg := fmt.Sprintf("Failed to archive message: %v", err)
						log.Println(msg)
						m.history.Append("", msg, "log", false, msg)
					} else {
						hitems = append(hitems[:idx], hitems[idx+1:]...)
					}
				} else {
					hitems = append(hitems[:idx], hitems[idx+1:]...)
				}
			}
		}
		items := make([]list.Item, len(hitems))
		for i, it := range hitems {
			it.IsSelected = nil
			hitems[i] = it
			items[i] = it
		}
		m.history.SetItems(hitems)
		m.history.List().SetItems(items)
		if len(hitems) == 0 {
			m.history.List().Select(-1)
		} else if m.history.List().Index() >= len(hitems) {
			m.history.List().Select(len(hitems) - 1)
		}
		m.history.SetSelectionAnchor(-1)
	}
	return nil
}

// handleDeleteHistoryKey deletes selected history messages.
func (m *model) handleDeleteHistoryKey() tea.Cmd {
	hitems := m.history.Items()
	if len(hitems) == 0 {
		return nil
	}
	hasSelection := false
	for i := range hitems {
		if hitems[i].IsSelected != nil && *hitems[i].IsSelected {
			v := true
			hitems[i].IsMarkedForDeletion = &v
			hasSelection = true
		}
	}
	if !hasSelection {
		idx := m.history.List().Index()
		if idx >= 0 && idx < len(hitems) {
			v := true
			hitems[idx].IsMarkedForDeletion = &v
		}
	}
	m.history.SetItems(hitems)
	rf := func() tea.Cmd { return m.SetFocus(m.ui.focusOrder[m.ui.focusIndex]) }
	m.StartConfirm("Delete selected messages? [y/n]", "", rf, func() tea.Cmd {
		hitems := m.history.Items()
		for i := len(hitems) - 1; i >= 0; i-- {
			it := hitems[i]
			if it.IsMarkedForDeletion != nil && *it.IsMarkedForDeletion {
				key := fmt.Sprintf("%s/%020d", it.Topic, it.Timestamp.UnixNano())
				if st := m.history.Store(); st != nil {
					if err := st.Delete(key); err != nil {
						msg := fmt.Sprintf("Failed to delete message: %v", err)
						log.Println(msg)
						m.history.Append("", msg, "log", false, msg)
						continue
					}
				}
				hitems = append(hitems[:i], hitems[i+1:]...)
			}
		}
		items := make([]list.Item, len(hitems))
		for i, it := range hitems {
			it.IsSelected = nil
			it.IsMarkedForDeletion = nil
			hitems[i] = it
			items[i] = it
		}
		m.history.SetItems(hitems)
		m.history.List().SetItems(items)
		if len(hitems) == 0 {
			m.history.List().Select(-1)
		} else if m.history.List().Index() >= len(hitems) {
			m.history.List().Select(len(hitems) - 1)
		}
		m.history.SetSelectionAnchor(-1)
		return nil
	}, func() {
		hitems := m.history.Items()
		for i := range hitems {
			hitems[i].IsMarkedForDeletion = nil
		}
		m.history.SetItems(hitems)
	})
	return nil
}
