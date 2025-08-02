package emqutiti

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/emqutiti/ui"
)

// topicsState holds topic list state shared across components.
type topicsState struct {
	input      textinput.Model
	items      []topicItem
	list       list.Model
	panes      topicsPanes
	selected   int
	chipBounds []chipBound
	vp         viewport.Model
}

func (t *topicsState) setTopic(topic string) { t.input.SetValue(topic) }

// topicsComponent implements the Component interface for topic management.
type topicsComponent struct{ m *model }

func newTopicsComponent(m *model) *topicsComponent { return &topicsComponent{m: m} }

func (c *topicsComponent) Init() tea.Cmd { return nil }

// Update manages the topics list UI.
func (c *topicsComponent) Update(msg tea.Msg) tea.Cmd {
	m := c.m
	var cmd, fcmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			return tea.Quit
		case "esc":
			cmd := m.setMode(modeClient)
			return cmd
		case "left":
			if m.topics.panes.active == 1 {
				fcmd = m.setFocus(idTopicsEnabled)
			}
		case "right":
			if m.topics.panes.active == 0 {
				fcmd = m.setFocus(idTopicsDisabled)
			}
		case "delete":
			i := m.topics.selected
			if i >= 0 && i < len(m.topics.items) {
				name := m.topics.items[i].title
				m.confirm.returnFocus = m.ui.focusOrder[m.ui.focusIndex]
				m.startConfirm(fmt.Sprintf("Delete topic '%s'? [y/n]", name), "", func() {
					m.removeTopic(i)
					m.rebuildActiveTopicList()
				})
				return m.connections.ListenStatus()
			}
		case "enter", " ":
			i := m.topics.selected
			if i >= 0 && i < len(m.topics.items) {
				m.toggleTopic(i)
				m.rebuildActiveTopicList()
			}
		}
	}
	m.topics.list, cmd = m.topics.list.Update(msg)
	if m.topics.panes.active == 0 {
		m.topics.panes.subscribed.sel = m.topics.list.Index()
		m.topics.panes.subscribed.page = m.topics.list.Paginator.Page
	} else {
		m.topics.panes.unsubscribed.sel = m.topics.list.Index()
		m.topics.panes.unsubscribed.page = m.topics.list.Paginator.Page
	}
	m.topics.selected = m.indexForPane(m.topics.panes.active, m.topics.list.Index())
	return tea.Batch(fcmd, cmd, m.connections.ListenStatus())
}

// View displays the topic manager list.
func (c *topicsComponent) View() string {
	m := c.m
	m.ui.elemPos = map[string]int{}
	m.ui.elemPos[idTopicsEnabled] = 1
	m.ui.elemPos[idTopicsDisabled] = 1
	help := ui.InfoStyle.Render("[space] toggle  [del] delete  [esc] back")
	activeView := m.topics.list.View()
	var left, right string
	if m.topics.panes.active == 0 {
		other := list.New(m.unsubscribedItems(), list.NewDefaultDelegate(), m.topics.list.Width(), m.topics.list.Height())
		other.DisableQuitKeybindings()
		other.SetShowTitle(false)
		other.Paginator.Page = m.topics.panes.unsubscribed.page
		other.Select(m.topics.panes.unsubscribed.sel)
		left = ui.LegendBox(activeView, "Enabled", m.ui.width/2-2, 0, ui.ColBlue, m.ui.focusOrder[m.ui.focusIndex] == idTopicsEnabled, -1)
		right = ui.LegendBox(other.View(), "Disabled", m.ui.width/2-2, 0, ui.ColBlue, false, -1)
	} else {
		other := list.New(m.subscribedItems(), list.NewDefaultDelegate(), m.topics.list.Width(), m.topics.list.Height())
		other.DisableQuitKeybindings()
		other.SetShowTitle(false)
		other.Paginator.Page = m.topics.panes.subscribed.page
		other.Select(m.topics.panes.subscribed.sel)
		left = ui.LegendBox(other.View(), "Enabled", m.ui.width/2-2, 0, ui.ColBlue, false, -1)
		right = ui.LegendBox(activeView, "Disabled", m.ui.width/2-2, 0, ui.ColBlue, m.ui.focusOrder[m.ui.focusIndex] == idTopicsDisabled, -1)
	}
	panes := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	content := lipgloss.JoinVertical(lipgloss.Left, panes, help)
	return m.overlayHelp(content)
}

func (c *topicsComponent) Focus() tea.Cmd { return nil }

func (c *topicsComponent) Blur() {}
