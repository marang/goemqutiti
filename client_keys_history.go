package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// handleCopyKey copies selected or current history items to the clipboard.
func (m *model) handleCopyKey() tea.Cmd {
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

// handleHistoryFilterKey opens the history filter when focused on history.
func (m *model) handleHistoryFilterKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idHistory {
		return m.startHistoryFilter()
	}
	return nil
}

// handleClearFilterKey clears any active history filters.
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
	var items []list.Item
	m.history.items, items = messagesToHistoryItems(msgs)
	m.history.list.SetItems(items)
	return nil
}

// handleHistoryViewKey opens a detail view for long history payloads.
func (m *model) handleHistoryViewKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] != idHistory {
		return nil
	}
	idx := m.history.list.Index()
	if idx < 0 || idx >= len(m.history.list.Items()) {
		return nil
	}
	hi := m.history.list.Items()[idx].(historyItem)
	if utf8.RuneCountInString(hi.payload) <= historyPreviewLimit {
		return nil
	}
	m.history.detailItem = hi
	m.history.detail.SetContent(hi.payload)
	m.history.detail.SetYOffset(0)
	return m.setMode(modeHistoryDetail)
}
