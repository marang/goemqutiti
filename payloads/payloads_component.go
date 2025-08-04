package payloads

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/emqutiti/focus"
	"github.com/marang/emqutiti/ui"
)

// LoadMsg requests that the model load a payload for editing.
type LoadMsg struct{ Topic, Payload string }

type Component struct {
	m      Model
	status StatusListener
	items  []Item
	list   list.Model
}

// New creates a payload management component.
func New(m Model, s StatusListener) *Component {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.DisableQuitKeybindings()
	l.SetShowTitle(false)
	return &Component{m: m, status: s, list: l}
}

func (p *Component) Init() tea.Cmd { return nil }

func (p *Component) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			return tea.Quit
		case "esc":
			return p.m.SetClientMode()
		case "delete":
			i := p.list.Index()
			if i >= 0 {
				items := p.list.Items()
				if i < len(items) {
					p.items = append(p.items[:i], p.items[i+1:]...)
					items = append(items[:i], items[i+1:]...)
					p.list.SetItems(items)
				}
			}
			return p.status.ListenStatus()
		case "enter":
			i := p.list.Index()
			if i >= 0 {
				items := p.list.Items()
				if i < len(items) {
					pi := items[i].(Item)
					return tea.Batch(
						func() tea.Msg { return LoadMsg{Topic: pi.Topic, Payload: pi.Payload} },
						p.m.SetClientMode(),
						p.status.ListenStatus(),
					)
				}
			}
		}
	}
	p.list, cmd = p.list.Update(msg)
	return tea.Batch(cmd, p.status.ListenStatus())
}

func (p *Component) View() string {
	m := p.m
	m.ResetElemPos()
	m.SetElemPos(IDList, 1)
	listView := p.list.View()
	help := ui.InfoStyle.Render("[enter] load  [del] delete  [esc] back")
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	focused := m.FocusedID() == IDList
	view := ui.LegendBox(content, "Payloads", m.Width()-2, 0, ui.ColBlue, focused, -1)
	return m.OverlayHelp(view)
}

func (p *Component) Focus() tea.Cmd { return nil }

func (p *Component) Blur() {}

// Focusables exposes focusable elements for the payloads component.
func (p *Component) Focusables() map[string]focus.Focusable {
	return map[string]focus.Focusable{IDList: &nullFocusable{}}
}

func (p *Component) Add(topic, payload string) {
	pi := Item{Topic: topic, Payload: payload}
	p.items = append(p.items, pi)
	items := append(p.list.Items(), pi)
	p.list.SetItems(items)
}

func (p *Component) Items() []Item { return p.items }

func (p *Component) SetItems(plds []Item) {
	p.items = plds
	items := make([]list.Item, len(plds))
	for i, pld := range plds {
		items[i] = pld
	}
	p.list.SetItems(items)
}

func (p *Component) Clear() { p.SetItems([]Item{}) }

// List exposes the underlying list model.
func (p *Component) List() *list.Model { return &p.list }

type nullFocusable struct{ focused bool }

func (n *nullFocusable) Focus()          { n.focused = true }
func (n *nullFocusable) Blur()           { n.focused = false }
func (n *nullFocusable) IsFocused() bool { return n.focused }
func (n *nullFocusable) View() string    { return "" }
