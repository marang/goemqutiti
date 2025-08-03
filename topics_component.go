package emqutiti

import (
	"fmt"
	"sort"

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

type topicToggleMsg struct {
	topic      string
	subscribed bool
}

// topicsComponent implements the Component interface for topic management.
type topicsComponent struct {
	*topicsState
	api topicsModel
}

func newTopicsComponent(api topicsModel) *topicsComponent {
	ts := initTopics()
	tc := &topicsComponent{topicsState: &ts, api: api}
	tc.panes.subscribed.m = tc
	tc.panes.unsubscribed.m = tc
	return tc
}

func (c *topicsComponent) Init() tea.Cmd { return nil }

// Update manages the topics list UI.
func (c *topicsComponent) Update(msg tea.Msg) tea.Cmd {
	var cmd, fcmd, tcmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			return tea.Quit
		case "esc":
			cmd := c.api.SetMode(modeClient)
			return cmd
		case "left":
			if c.panes.active == 1 {
				fcmd = c.api.SetFocus(idTopicsEnabled)
			}
		case "right":
			if c.panes.active == 0 {
				fcmd = c.api.SetFocus(idTopicsDisabled)
			}
		case "delete":
			i := c.selected
			if i >= 0 && i < len(c.items) {
				name := c.items[i].title
				rf := func() tea.Cmd { return c.api.SetFocus(c.api.FocusedID()) }
				c.api.StartConfirm(fmt.Sprintf("Delete topic '%s'? [y/n]", name), "", rf, func() tea.Cmd {
					cmd := c.RemoveTopic(i)
					c.RebuildActiveTopicList()
					return cmd
				}, nil)
				return c.api.ListenStatus()
			}
		case "enter", " ":
			i := c.selected
			if i >= 0 && i < len(c.items) {
				tcmd = c.ToggleTopic(i)
			}
		}
	}
	c.list, cmd = c.list.Update(msg)
	if c.panes.active == 0 {
		c.panes.subscribed.sel = c.list.Index()
		c.panes.subscribed.page = c.list.Paginator.Page
	} else {
		c.panes.unsubscribed.sel = c.list.Index()
		c.panes.unsubscribed.page = c.list.Paginator.Page
	}
	c.selected = c.IndexForPane(c.panes.active, c.list.Index())
	return tea.Batch(fcmd, tcmd, cmd, c.api.ListenStatus())
}

// View displays the topic manager list.
func (c *topicsComponent) View() string {
	c.api.ResetElemPos()
	c.api.SetElemPos(idTopicsEnabled, 1)
	c.api.SetElemPos(idTopicsDisabled, 1)
	help := ui.InfoStyle.Render("[space] toggle  [del] delete  [esc] back")
	activeView := c.list.View()
	var left, right string
	if c.panes.active == 0 {
		other := list.New(c.UnsubscribedItems(), list.NewDefaultDelegate(), c.list.Width(), c.list.Height())
		other.DisableQuitKeybindings()
		other.SetShowTitle(false)
		other.Paginator.Page = c.panes.unsubscribed.page
		other.Select(c.panes.unsubscribed.sel)
		left = ui.LegendBox(activeView, "Enabled", c.api.Width()/2-2, 0, ui.ColBlue, c.api.FocusedID() == idTopicsEnabled, -1)
		right = ui.LegendBox(other.View(), "Disabled", c.api.Width()/2-2, 0, ui.ColBlue, false, -1)
	} else {
		other := list.New(c.SubscribedItems(), list.NewDefaultDelegate(), c.list.Width(), c.list.Height())
		other.DisableQuitKeybindings()
		other.SetShowTitle(false)
		other.Paginator.Page = c.panes.subscribed.page
		other.Select(c.panes.subscribed.sel)
		left = ui.LegendBox(other.View(), "Enabled", c.api.Width()/2-2, 0, ui.ColBlue, false, -1)
		right = ui.LegendBox(activeView, "Disabled", c.api.Width()/2-2, 0, ui.ColBlue, c.api.FocusedID() == idTopicsDisabled, -1)
	}
	panes := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	content := lipgloss.JoinVertical(lipgloss.Left, panes, help)
	return c.api.OverlayHelp(content)
}

func (c *topicsComponent) Focus() tea.Cmd { return nil }

func (c *topicsComponent) Blur() {}

// UpdateInput routes messages to the topic text input.
func (c *topicsComponent) UpdateInput(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	c.input, cmd = c.input.Update(msg)
	return cmd
}

// Focusables exposes focusable elements for the topics component.
func (c *topicsComponent) Focusables() map[string]Focusable {
	return map[string]Focusable{
		idTopicsEnabled:  &c.panes.subscribed,
		idTopicsDisabled: &c.panes.unsubscribed,
		idTopic:          adapt(&c.input),
	}
}

// HasTopic reports whether the given topic already exists in the list.
func (c *topicsComponent) HasTopic(topic string) bool {
	for _, t := range c.items {
		if t.title == topic {
			return true
		}
	}
	return false
}

// SortTopics orders the topic list with active topics first and keeps selection.
func (c *topicsComponent) SortTopics() {
	if len(c.items) == 0 {
		return
	}
	sel := ""
	if c.selected >= 0 && c.selected < len(c.items) {
		sel = c.items[c.selected].title
	}
	sort.SliceStable(c.items, func(i, j int) bool {
		if c.items[i].subscribed != c.items[j].subscribed {
			return c.items[i].subscribed && !c.items[j].subscribed
		}
		return c.items[i].title < c.items[j].title
	})
	if sel != "" {
		for i, t := range c.items {
			if t.title == sel {
				c.selected = i
				break
			}
		}
	}
}

// SubscribedItems returns topics currently subscribed.
func (c *topicsComponent) SubscribedItems() []list.Item {
	var out []list.Item
	for _, t := range c.items {
		if t.subscribed {
			out = append(out, t)
		}
	}
	return out
}

// UnsubscribedItems returns topics that are not subscribed.
func (c *topicsComponent) UnsubscribedItems() []list.Item {
	var out []list.Item
	for _, t := range c.items {
		if !t.subscribed {
			out = append(out, t)
		}
	}
	return out
}

// IndexForPane converts a pane list index to a global topics index.
func (c *topicsComponent) IndexForPane(pane, idx int) int {
	count := -1
	for i, t := range c.items {
		if (pane == 0 && t.subscribed) || (pane == 1 && !t.subscribed) {
			count++
			if count == idx {
				return i
			}
		}
	}
	return -1
}

// RebuildActiveTopicList updates the active list to show the current pane.
func (c *topicsComponent) RebuildActiveTopicList() {
	if c.panes.active == 0 {
		items := c.SubscribedItems()
		if c.panes.subscribed.sel >= len(items) {
			c.panes.subscribed.sel = len(items) - 1
		}
		if c.panes.subscribed.sel < 0 && len(items) > 0 {
			c.panes.subscribed.sel = 0
		}
		c.list.SetItems(items)
		if len(items) > 0 {
			c.list.Select(c.panes.subscribed.sel)
		}
		c.list.Paginator.Page = c.panes.subscribed.page
		c.selected = c.IndexForPane(0, c.panes.subscribed.sel)
	} else {
		items := c.UnsubscribedItems()
		if c.panes.unsubscribed.sel >= len(items) {
			c.panes.unsubscribed.sel = len(items) - 1
		}
		if c.panes.unsubscribed.sel < 0 && len(items) > 0 {
			c.panes.unsubscribed.sel = 0
		}
		c.list.SetItems(items)
		if len(items) > 0 {
			c.list.Select(c.panes.unsubscribed.sel)
		}
		c.list.Paginator.Page = c.panes.unsubscribed.page
		c.selected = c.IndexForPane(1, c.panes.unsubscribed.sel)
	}
}

// SetActivePane switches focus to the given pane index and rebuilds the list.
func (c *topicsComponent) SetActivePane(idx int) {
	if idx == c.panes.active {
		return
	}
	if c.panes.active == 0 {
		c.panes.subscribed.sel = c.list.Index()
		c.panes.subscribed.page = c.list.Paginator.Page
	} else {
		c.panes.unsubscribed.sel = c.list.Index()
		c.panes.unsubscribed.page = c.list.Paginator.Page
	}
	c.panes.active = idx
	c.RebuildActiveTopicList()
}

// ToggleTopic toggles the subscription state of the topic at index and emits an event.
func (c *topicsComponent) ToggleTopic(index int) tea.Cmd {
	if index < 0 || index >= len(c.items) {
		return nil
	}
	t := &c.items[index]
	t.subscribed = !t.subscribed
	c.SortTopics()
	c.RebuildActiveTopicList()
	topic := t.title
	sub := t.subscribed
	return func() tea.Msg { return topicToggleMsg{topic: topic, subscribed: sub} }
}

// RemoveTopic deletes the topic at index and emits an unsubscribe event.
func (c *topicsComponent) RemoveTopic(index int) tea.Cmd {
	if index < 0 || index >= len(c.items) {
		return nil
	}
	topic := c.items[index].title
	c.items = append(c.items[:index], c.items[index+1:]...)
	if len(c.items) == 0 {
		c.selected = -1
	} else if c.selected >= len(c.items) {
		c.selected = len(c.items) - 1
	}
	c.SortTopics()
	c.RebuildActiveTopicList()
	return func() tea.Msg { return topicToggleMsg{topic: topic, subscribed: false} }
}

// TopicAtPosition returns the index of the topic chip at the provided coordinates or -1.
func (c *topicsComponent) TopicAtPosition(x, y int) int {
	for i, b := range c.chipBounds {
		if x >= b.xPos && x < b.xPos+b.width && y >= b.yPos && y < b.yPos+b.height {
			return i
		}
	}
	return -1
}

func (c *topicsComponent) SetSelected(i int) { c.selected = i }
func (c *topicsComponent) Selected() int     { return c.selected }

var _ TopicsAPI = (*topicsComponent)(nil)
