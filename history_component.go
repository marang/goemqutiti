package emqutiti

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/emqutiti/ui"
)

type historyAppender interface {
	Add(Message)
}

type historyQuerier interface {
	Search(archived bool, topics []string, start, end time.Time, payload string) []Message
}

type historyStore interface {
	historyAppender
	historyQuerier
}

type historyState struct {
	list            list.Model
	items           []historyItem
	store           *HistoryStore
	selectionAnchor int
	showArchived    bool
	filterForm      *historyFilterForm
	filterQuery     string
	detail          viewport.Model
	detailItem      historyItem
}

// updateHistoryList updates the history list when focused.
func (m *model) updateHistoryList(msg tea.Msg) tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] != idHistory {
		return nil
	}
	var cmd tea.Cmd
	m.history.list, cmd = m.history.list.Update(msg)
	return cmd
}

// updateHistoryDetail handles input when viewing a long history payload.
func (m *model) updateHistoryDetail(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			cmd := m.setMode(modeClient)
			return cmd
		case "ctrl+d":
			return tea.Quit
		}
	}
	m.history.detail, cmd = m.history.detail.Update(msg)
	return cmd
}

// updateHistoryFilter handles the history filter form interaction.
func (m *model) updateHistoryFilter(msg tea.Msg) tea.Cmd {
	if m.history.filterForm == nil {
		return nil
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
			return cmd
		case "enter":
			q := m.history.filterForm.query()
			topics, start, end, payload := parseHistoryQuery(q)
			var msgs []Message
			if m.history.showArchived {
				msgs = m.history.store.Search(true, topics, start, end, payload)
			} else {
				msgs = m.history.store.Search(false, topics, start, end, payload)
			}
			var items []list.Item
			m.history.items, items = messagesToHistoryItems(msgs)
			m.history.list.SetItems(items)
			m.history.list.FilterInput.SetValue("")
			m.history.list.SetFilterState(list.Unfiltered)
			m.history.filterQuery = q
			m.history.filterForm = nil
			cmd := tea.Batch(m.setMode(m.previousMode()), m.setFocus(idHistory))
			return cmd
		}
	}
	f, cmd := m.history.filterForm.Update(msg)
	m.history.filterForm = &f
	return cmd
}

// viewHistoryDetail renders the full payload of a history message.
func (m *model) viewHistoryDetail() string {
	m.ui.elemPos = map[string]int{}
	lines := strings.Split(m.history.detail.View(), "\n")
	help := ui.InfoStyle.Render("[esc] back")
	lines = append(lines, help)
	content := strings.Join(lines, "\n")
	sp := -1.0
	if m.history.detail.Height < lipgloss.Height(content) {
		sp = m.history.detail.ScrollPercent()
	}
	view := ui.LegendBox(content, "Message", m.ui.width-2, m.ui.height-2, ui.ColGreen, true, sp)
	return m.overlayHelp(view)
}

// viewHistoryFilter displays the history filter form.
func (m *model) viewHistoryFilter() string {
	m.ui.elemPos = map[string]int{}
	if m.history.filterForm == nil {
		return ""
	}
	content := lipgloss.NewStyle().Padding(1, 2).Render(m.history.filterForm.View())
	box := ui.LegendBox(content, "Filter", m.ui.width/2, 0, ui.ColBlue, true, -1)
	return lipgloss.Place(m.ui.width, m.ui.height, lipgloss.Center, lipgloss.Center, box)
}
