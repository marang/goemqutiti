package emqutiti

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/emqutiti/history"
)

// handleCopyKey copies selected or current history items to the clipboard.
func (m *model) handleCopyKey() tea.Cmd {
	var idxs []int
	hitems := m.history.Items()
	for i, it := range hitems {
		if it.IsSelected != nil && *it.IsSelected {
			idxs = append(idxs, i)
		}
	}
	if len(idxs) > 0 {
		sort.Ints(idxs)
		var parts []string
		items := m.history.List().Items()
		for _, i := range idxs {
			if i >= 0 && i < len(items) {
				hi := items[i].(history.Item)
				txt := hi.Payload
				if hi.Kind != "log" {
					txt = fmt.Sprintf("%s: %s", hi.Topic, hi.Payload)
				}
				parts = append(parts, txt)
			}
		}
		if err := clipboard.WriteAll(strings.Join(parts, "\n")); err != nil {
			m.history.Append("", err.Error(), "log", err.Error())
		}
	} else if len(m.history.List().Items()) > 0 {
		idx := m.history.List().Index()
		if idx >= 0 {
			hi := m.history.List().Items()[idx].(history.Item)
			text := hi.Payload
			if hi.Kind != "log" {
				text = fmt.Sprintf("%s: %s", hi.Topic, hi.Payload)
			}
			if err := clipboard.WriteAll(text); err != nil {
				m.history.Append("", err.Error(), "log", err.Error())
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
	m.history.SetFilterQuery("")
	m.history.List().FilterInput.SetValue("")
	m.history.List().SetFilterState(list.Unfiltered)
	var msgs []history.Message
	if m.history.ShowArchived() {
		msgs = m.history.Store().Search(true, nil, time.Time{}, time.Time{}, "")
	} else {
		msgs = m.history.Store().Search(false, nil, time.Time{}, time.Time{}, "")
	}
	hitems, items := messagesToHistoryItems(msgs)
	m.history.SetItems(hitems)
	m.history.List().SetItems(items)
	return nil
}

// handleHistoryViewKey opens a detail view for long history payloads.
func (m *model) handleHistoryViewKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] != idHistory {
		return nil
	}
	idx := m.history.List().Index()
	if idx < 0 || idx >= len(m.history.List().Items()) {
		return nil
	}
	hi := m.history.List().Items()[idx].(history.Item)
	if utf8.RuneCountInString(hi.Payload) <= historyPreviewLimit {
		return nil
	}
	m.history.SetDetailItem(hi)
	m.history.Detail().SetContent(hi.Payload)
	m.history.Detail().SetYOffset(0)
	return m.setMode(modeHistoryDetail)
}
