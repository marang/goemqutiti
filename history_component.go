package emqutiti

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/emqutiti/ui"
)

type historyState struct {
	list            list.Model
	items           []historyItem
	store           HistoryStore
	selectionAnchor int
	showArchived    bool
	filterForm      *historyFilterForm
	filterQuery     string
	detail          viewport.Model
	detailItem      historyItem
}

// historyComponent provides a Component implementation for browsing and
// filtering message history. It holds its own state while delegating cross
// component interactions back to the root model.
type historyComponent struct {
	*historyState
	m *model
}

func newHistoryComponent(m *model, hs historyState) *historyComponent {
	return &historyComponent{historyState: &hs, m: m}
}

func (h *historyComponent) Init() tea.Cmd { return nil }

// Update updates the history list when it has focus in client mode.
func (h *historyComponent) Update(msg tea.Msg) tea.Cmd {
	if h.m.ui.focusOrder[h.m.ui.focusIndex] != idHistory {
		return nil
	}
	var cmd tea.Cmd
	h.list, cmd = h.list.Update(msg)
	return cmd
}

func (h *historyComponent) View() string { return "" }

func (h *historyComponent) Focus() tea.Cmd { return nil }

func (h *historyComponent) Blur() {}

// UpdateDetail handles input when viewing a long history payload.
func (h *historyComponent) UpdateDetail(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			cmd := h.m.setMode(modeClient)
			return cmd
		case "ctrl+d":
			return tea.Quit
		}
	}
	h.detail, cmd = h.detail.Update(msg)
	return cmd
}

// UpdateFilter handles the history filter form interaction.
func (h *historyComponent) UpdateFilter(msg tea.Msg) tea.Cmd {
	if h.filterForm == nil {
		return nil
	}
	switch t := msg.(type) {
	case tea.KeyMsg:
		switch t.String() {
		case "esc":
			h.filterForm = nil
			if len(h.m.ui.modeStack) > 0 {
				h.m.ui.modeStack = h.m.ui.modeStack[1:]
			}
			if len(h.m.ui.modeStack) > 0 && h.m.ui.modeStack[0] == modeHelp {
				h.m.ui.modeStack = h.m.ui.modeStack[1:]
			}
			cmd := tea.Batch(h.m.setMode(h.m.currentMode()), h.m.setFocus(idHistory))
			return cmd
		case "enter":
			q := h.filterForm.query()
			topics, start, end, payload := parseHistoryQuery(q)
			var msgs []Message
			if h.showArchived {
				msgs = h.store.Search(true, topics, start, end, payload)
			} else {
				msgs = h.store.Search(false, topics, start, end, payload)
			}
			var items []list.Item
			h.items, items = messagesToHistoryItems(msgs)
			h.list.SetItems(items)
			h.list.FilterInput.SetValue("")
			h.list.SetFilterState(list.Unfiltered)
			h.filterQuery = q
			h.filterForm = nil
			cmd := tea.Batch(h.m.setMode(h.m.previousMode()), h.m.setFocus(idHistory))
			return cmd
		}
	}
	f, cmd := h.filterForm.Update(msg)
	h.filterForm = &f
	return cmd
}

// ViewDetail renders the full payload of a history message.
func (h *historyComponent) ViewDetail() string {
	h.m.ui.elemPos = map[string]int{}
	lines := strings.Split(h.detail.View(), "\n")
	help := ui.InfoStyle.Render("[esc] back")
	lines = append(lines, help)
	content := strings.Join(lines, "\n")
	sp := -1.0
	if h.detail.Height < lipgloss.Height(content) {
		sp = h.detail.ScrollPercent()
	}
	view := ui.LegendBox(content, "Message", h.m.ui.width-2, h.m.ui.height-2, ui.ColGreen, true, sp)
	return h.m.overlayHelp(view)
}

// ViewFilter displays the history filter form.
func (h *historyComponent) ViewFilter() string {
	h.m.ui.elemPos = map[string]int{}
	if h.filterForm == nil {
		return ""
	}
	content := lipgloss.NewStyle().Padding(1, 2).Render(h.filterForm.View())
	box := ui.LegendBox(content, "Filter", h.m.ui.width/2, 0, ui.ColBlue, true, -1)
	return lipgloss.Place(h.m.ui.width, h.m.ui.height, lipgloss.Center, lipgloss.Center, box)
}

// Focusables exposes focusable elements for history component. The history list
// itself is managed by the root model, so this returns an empty map.
func (h *historyComponent) Focusables() map[string]Focusable { return map[string]Focusable{} }
