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
type topicsComponent struct {
	*topicsState
	api TopicsAPI
}

func newTopicsComponent(api TopicsAPI) *topicsComponent {
	ts := initTopics()
	ts.panes.subscribed.m = api
	ts.panes.unsubscribed.m = api
	return &topicsComponent{topicsState: &ts, api: api}
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
					cmd := c.api.RemoveTopic(i)
					c.api.RebuildActiveTopicList()
					return cmd
				}, nil)
				return c.api.ListenStatus()
			}
		case "enter", " ":
			i := c.selected
			if i >= 0 && i < len(c.items) {
				tcmd = c.api.ToggleTopic(i)
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
	c.selected = c.api.IndexForPane(c.panes.active, c.list.Index())
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
		other := list.New(c.api.UnsubscribedItems(), list.NewDefaultDelegate(), c.list.Width(), c.list.Height())
		other.DisableQuitKeybindings()
		other.SetShowTitle(false)
		other.Paginator.Page = c.panes.unsubscribed.page
		other.Select(c.panes.unsubscribed.sel)
		left = ui.LegendBox(activeView, "Enabled", c.api.Width()/2-2, 0, ui.ColBlue, c.api.FocusedID() == idTopicsEnabled, -1)
		right = ui.LegendBox(other.View(), "Disabled", c.api.Width()/2-2, 0, ui.ColBlue, false, -1)
	} else {
		other := list.New(c.api.SubscribedItems(), list.NewDefaultDelegate(), c.list.Width(), c.list.Height())
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

// Focusables exposes focusable elements for the topics component.
func (c *topicsComponent) Focusables() map[string]Focusable {
	return map[string]Focusable{
		idTopicsEnabled:  &c.panes.subscribed,
		idTopicsDisabled: &c.panes.unsubscribed,
	}
}

func (c *topicsComponent) SetSelected(i int) { c.selected = i }
func (c *topicsComponent) Selected() int     { return c.selected }
