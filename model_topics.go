package emqutiti

import (
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type topicItem struct {
	title      string
	subscribed bool
}

func (t topicItem) FilterValue() string { return t.title }
func (t topicItem) Title() string       { return t.title }
func (t topicItem) Description() string {
	if t.subscribed {
		return "subscribed"
	}
	return "unsubscribed"
}

type chipBound struct {
	xPos, yPos    int
	width, height int
}

type paneState struct {
	sel     int
	page    int
	m       *model
	index   int
	focused bool
}

func (p *paneState) Focus() {
	p.focused = true
	if p.m != nil {
		p.m.setActivePane(p.index)
	}
}

func (p *paneState) Blur() { p.focused = false }

func (p paneState) IsFocused() bool { return p.focused }

func (p paneState) View() string { return "" }

type topicsPanes struct {
	subscribed   paneState
	unsubscribed paneState
	active       int
}

type topicToggleMsg struct {
	topic      string
	subscribed bool
}

// hasTopic reports whether the given topic already exists in the list.
func (m *model) hasTopic(topic string) bool {
	for _, t := range m.topics.items {
		if t.title == topic {
			return true
		}
	}
	return false
}

// sortTopics orders the topic list with active topics first and keeps selection.
func (m *model) sortTopics() {
	if len(m.topics.items) == 0 {
		return
	}
	sel := ""
	if m.topics.selected >= 0 && m.topics.selected < len(m.topics.items) {
		sel = m.topics.items[m.topics.selected].title
	}
	sort.SliceStable(m.topics.items, func(i, j int) bool {
		if m.topics.items[i].subscribed != m.topics.items[j].subscribed {
			return m.topics.items[i].subscribed && !m.topics.items[j].subscribed
		}
		return m.topics.items[i].title < m.topics.items[j].title
	})
	if sel != "" {
		for i, t := range m.topics.items {
			if t.title == sel {
				m.topics.selected = i
				break
			}
		}
	}
}

// subscribedItems returns topics currently subscribed.
func (m *model) subscribedItems() []list.Item {
	var out []list.Item
	for _, t := range m.topics.items {
		if t.subscribed {
			out = append(out, t)
		}
	}
	return out
}

// unsubscribedItems returns topics that are not subscribed.
func (m *model) unsubscribedItems() []list.Item {
	var out []list.Item
	for _, t := range m.topics.items {
		if !t.subscribed {
			out = append(out, t)
		}
	}
	return out
}

// indexForPane converts a pane list index to a global topics index.
func (m *model) indexForPane(pane, idx int) int {
	count := -1
	for i, t := range m.topics.items {
		if (pane == 0 && t.subscribed) || (pane == 1 && !t.subscribed) {
			count++
			if count == idx {
				return i
			}
		}
	}
	return -1
}

// rebuildActiveTopicList updates the active list to show the current pane.
func (m *model) rebuildActiveTopicList() {
	if m.topics.panes.active == 0 {
		items := m.subscribedItems()
		if m.topics.panes.subscribed.sel >= len(items) {
			m.topics.panes.subscribed.sel = len(items) - 1
		}
		if m.topics.panes.subscribed.sel < 0 && len(items) > 0 {
			m.topics.panes.subscribed.sel = 0
		}
		m.topics.list.SetItems(items)
		if len(items) > 0 {
			m.topics.list.Select(m.topics.panes.subscribed.sel)
		}
		m.topics.list.Paginator.Page = m.topics.panes.subscribed.page
		m.topics.selected = m.indexForPane(0, m.topics.panes.subscribed.sel)
	} else {
		items := m.unsubscribedItems()
		if m.topics.panes.unsubscribed.sel >= len(items) {
			m.topics.panes.unsubscribed.sel = len(items) - 1
		}
		if m.topics.panes.unsubscribed.sel < 0 && len(items) > 0 {
			m.topics.panes.unsubscribed.sel = 0
		}
		m.topics.list.SetItems(items)
		if len(items) > 0 {
			m.topics.list.Select(m.topics.panes.unsubscribed.sel)
		}
		m.topics.list.Paginator.Page = m.topics.panes.unsubscribed.page
		m.topics.selected = m.indexForPane(1, m.topics.panes.unsubscribed.sel)
	}
}

// setActivePane switches focus to the given pane index and rebuilds the list.
func (m *model) setActivePane(idx int) {
	if idx == m.topics.panes.active {
		return
	}
	if m.topics.panes.active == 0 {
		m.topics.panes.subscribed.sel = m.topics.list.Index()
		m.topics.panes.subscribed.page = m.topics.list.Paginator.Page
	} else {
		m.topics.panes.unsubscribed.sel = m.topics.list.Index()
		m.topics.panes.unsubscribed.page = m.topics.list.Paginator.Page
	}
	m.topics.panes.active = idx
	m.rebuildActiveTopicList()
}

// toggleTopic toggles the subscription state of the topic at index and emits an event.
func (m *model) toggleTopic(index int) tea.Cmd {
	if index < 0 || index >= len(m.topics.items) {
		return nil
	}
	t := &m.topics.items[index]
	t.subscribed = !t.subscribed
	m.sortTopics()
	m.rebuildActiveTopicList()
	topic := t.title
	sub := t.subscribed
	return func() tea.Msg { return topicToggleMsg{topic: topic, subscribed: sub} }
}

// removeTopic deletes the topic at index and emits an unsubscribe event.
func (m *model) removeTopic(index int) tea.Cmd {
	if index < 0 || index >= len(m.topics.items) {
		return nil
	}
	topic := m.topics.items[index].title
	m.topics.items = append(m.topics.items[:index], m.topics.items[index+1:]...)
	if len(m.topics.items) == 0 {
		m.topics.selected = -1
	} else if m.topics.selected >= len(m.topics.items) {
		m.topics.selected = len(m.topics.items) - 1
	}
	m.sortTopics()
	m.rebuildActiveTopicList()
	return func() tea.Msg { return topicToggleMsg{topic: topic, subscribed: false} }
}
