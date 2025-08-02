package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *model) handleCopyKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] != idHistory {
		return nil
	}
	var idxs []int
	for i, it := range m.history.items {
		if it.isSelected != nil && *it.isSelected {
			idxs = append(idxs, i)
		}
	}
	if len(idxs) > 0 {
		sort.Ints(idxs)
		var parts []string
		items := m.history.list.Items()
		for _, i := range idxs {
			if i >= 0 && i < len(items) {
				hi := items[i].(historyItem)
				txt := hi.payload
				if hi.kind != "log" {
					txt = fmt.Sprintf("%s: %s", hi.topic, hi.payload)
				}
				parts = append(parts, txt)
			}
		}
		if err := clipboard.WriteAll(strings.Join(parts, "\n")); err != nil {
			m.appendHistory("", err.Error(), "log", err.Error())
		}
	} else if len(m.history.list.Items()) > 0 {
		idx := m.history.list.Index()
		if idx >= 0 {
			hi := m.history.list.Items()[idx].(historyItem)
			text := hi.payload
			if hi.kind != "log" {
				text = fmt.Sprintf("%s: %s", hi.topic, hi.payload)
			}
			if err := clipboard.WriteAll(text); err != nil {
				m.appendHistory("", err.Error(), "log", err.Error())
			}
		}
	}
	return nil
}

func (m *model) handleHistoryFilterKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] != idHistory {
		return nil
	}
	return m.startHistoryFilter()
}

func (m *model) handleClearFilterKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] != idHistory {
		return nil
	}
	m.history.filterQuery = ""
	m.history.list.FilterInput.SetValue("")
	m.history.list.SetFilterState(list.Unfiltered)
	var msgs []Message
	if m.history.showArchived {
		msgs = m.history.store.SearchArchived(nil, time.Time{}, time.Time{}, "")
	} else {
		msgs = m.history.store.Search(nil, time.Time{}, time.Time{}, "")
	}
	m.history.items = make([]historyItem, len(msgs))
	items := make([]list.Item, len(msgs))
	for i, mm := range msgs {
		hi := historyItem{timestamp: mm.Timestamp, topic: mm.Topic, payload: mm.Payload, kind: mm.Kind, archived: mm.Archived}
		m.history.items[i] = hi
		items[i] = hi
	}
	m.history.list.SetItems(items)
	return nil
}

func (m *model) handleSpaceKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] != idHistory || m.history.showArchived {
		return nil
	}
	idx := m.history.list.Index()
	if idx >= 0 && idx < len(m.history.items) {
		if m.history.items[idx].isSelected != nil && *m.history.items[idx].isSelected {
			m.history.items[idx].isSelected = nil
		} else {
			v := true
			m.history.items[idx].isSelected = &v
		}
		m.history.selectionAnchor = idx
	}
	return nil
}

func (m *model) handleShiftUpKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] != idHistory || m.history.showArchived {
		return nil
	}
	if m.history.selectionAnchor == -1 {
		m.history.selectionAnchor = m.history.list.Index()
		if m.history.selectionAnchor >= 0 && m.history.selectionAnchor < len(m.history.items) {
			v := true
			m.history.items[m.history.selectionAnchor].isSelected = &v
		}
	}
	if m.history.list.Index() > 0 {
		m.history.list.CursorUp()
		idx := m.history.list.Index()
		m.updateSelectionRange(idx)
	}
	return nil
}

func (m *model) handleShiftDownKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] != idHistory || m.history.showArchived {
		return nil
	}
	if m.history.selectionAnchor == -1 {
		m.history.selectionAnchor = m.history.list.Index()
		if m.history.selectionAnchor >= 0 && m.history.selectionAnchor < len(m.history.items) {
			v := true
			m.history.items[m.history.selectionAnchor].isSelected = &v
		}
	}
	if m.history.list.Index() < len(m.history.list.Items())-1 {
		m.history.list.CursorDown()
		idx := m.history.list.Index()
		m.updateSelectionRange(idx)
	}
	return nil
}

func (m *model) handleSelectAllKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] != idHistory || m.history.showArchived {
		return nil
	}
	allSelected := true
	for i := range m.history.items {
		if m.history.items[i].isSelected == nil || !*m.history.items[i].isSelected {
			allSelected = false
			break
		}
	}
	if allSelected {
		for i := range m.history.items {
			m.history.items[i].isSelected = nil
		}
		m.history.selectionAnchor = -1
	} else {
		for i := range m.history.items {
			v := true
			m.history.items[i].isSelected = &v
		}
		if len(m.history.items) > 0 {
			m.history.selectionAnchor = 0
		}
	}
	return nil
}

func (m *model) handleToggleArchiveKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] != idHistory || m.history.store == nil {
		return nil
	}
	m.history.showArchived = !m.history.showArchived
	var msgs []Message
	if m.history.showArchived {
		msgs = m.history.store.SearchArchived(nil, time.Time{}, time.Time{}, "")
	} else {
		msgs = m.history.store.Search(nil, time.Time{}, time.Time{}, "")
	}
	m.history.items = make([]historyItem, len(msgs))
	items := make([]list.Item, len(msgs))
	for i, mm := range msgs {
		hi := historyItem{timestamp: mm.Timestamp, topic: mm.Topic, payload: mm.Payload, kind: mm.Kind, archived: mm.Archived}
		m.history.items[i] = hi
		items[i] = hi
	}
	m.history.list.SetItems(items)
	if len(items) > 0 {
		m.history.list.Select(len(items) - 1)
	} else {
		m.history.list.Select(-1)
	}
	return nil
}

func (m *model) handleHistoryScrollKeys(msg tea.KeyMsg) tea.Cmd {
	// keep current selection and anchor
	return nil
}

func (m *model) handleArchiveKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] != idHistory || m.history.showArchived {
		return nil
	}
	if len(m.history.items) == 0 {
		return nil
	}
	archived := false
	for i := len(m.history.items) - 1; i >= 0; i-- {
		it := m.history.items[i]
		if it.isSelected != nil && *it.isSelected {
			key := fmt.Sprintf("%s/%020d", it.topic, it.timestamp.UnixNano())
			if m.history.store != nil {
				_ = m.history.store.Archive(key)
			}
			m.history.items = append(m.history.items[:i], m.history.items[i+1:]...)
			archived = true
		}
	}
	if !archived {
		idx := m.history.list.Index()
		if idx >= 0 && idx < len(m.history.items) {
			it := m.history.items[idx]
			key := fmt.Sprintf("%s/%020d", it.topic, it.timestamp.UnixNano())
			if m.history.store != nil {
				_ = m.history.store.Archive(key)
			}
			m.history.items = append(m.history.items[:idx], m.history.items[idx+1:]...)
		}
	}
	items := make([]list.Item, len(m.history.items))
	for i, it := range m.history.items {
		it.isSelected = nil
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
}

func (m *model) handleHistoryDeleteKey() tea.Cmd {
	if len(m.history.items) == 0 {
		return nil
	}
	hasSelection := false
	for i := range m.history.items {
		if m.history.items[i].isSelected != nil && *m.history.items[i].isSelected {
			v := true
			m.history.items[i].isMarkedForDeletion = &v
			hasSelection = true
		}
	}
	if !hasSelection {
		idx := m.history.list.Index()
		if idx >= 0 && idx < len(m.history.items) {
			v := true
			m.history.items[idx].isMarkedForDeletion = &v
		}
	}
	m.confirmReturnFocus = m.ui.focusOrder[m.ui.focusIndex]
	m.startConfirm("Delete selected messages? [y/n]", "", func() {
		for i := len(m.history.items) - 1; i >= 0; i-- {
			it := m.history.items[i]
			if it.isMarkedForDeletion != nil && *it.isMarkedForDeletion {
				key := fmt.Sprintf("%s/%020d", it.topic, it.timestamp.UnixNano())
				if m.history.store != nil {
					_ = m.history.store.Delete(key)
				}
				m.history.items = append(m.history.items[:i], m.history.items[i+1:]...)
			}
		}
		items := make([]list.Item, len(m.history.items))
		for i, it := range m.history.items {
			it.isSelected = nil
			it.isMarkedForDeletion = nil
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
	})
	m.confirmCancel = func() {
		for i := range m.history.items {
			m.history.items[i].isMarkedForDeletion = nil
		}
	}
	return nil
}
