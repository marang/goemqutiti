package emqutiti

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/emqutiti/ui"
)

type payloadsComponent struct {
	m     *model
	items []payloadItem
	list  list.Model
}

func newPayloadsComponent(m *model) *payloadsComponent {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.DisableQuitKeybindings()
	l.SetShowTitle(false)
	return &payloadsComponent{m: m, list: l}
}

func (p *payloadsComponent) Init() tea.Cmd { return nil }

func (p *payloadsComponent) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			return tea.Quit
		case "esc":
			return p.m.setMode(modeClient)
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
			return p.m.connections.ListenStatus()
		case "enter":
			i := p.list.Index()
			if i >= 0 {
				items := p.list.Items()
				if i < len(items) {
					pi := items[i].(payloadItem)
					p.m.topics.setTopic(pi.topic)
					p.m.message.setPayload(pi.payload)
					return p.m.setMode(modeClient)
				}
			}
		}
	}
	p.list, cmd = p.list.Update(msg)
	return tea.Batch(cmd, p.m.connections.ListenStatus())
}

func (p *payloadsComponent) View() string {
	m := p.m
	m.ui.elemPos = map[string]int{}
	m.ui.elemPos[idPayloadList] = 1
	listView := p.list.View()
	help := ui.InfoStyle.Render("[enter] load  [del] delete  [esc] back")
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	focused := m.ui.focusOrder[m.ui.focusIndex] == idPayloadList
	view := ui.LegendBox(content, "Payloads", m.ui.width-2, 0, ui.ColBlue, focused, -1)
	return m.overlayHelp(view)
}

func (p *payloadsComponent) Focus() tea.Cmd { return nil }

func (p *payloadsComponent) Blur() {}

// Focusables exposes focusable elements for the payloads component.
func (p *payloadsComponent) Focusables() map[string]Focusable {
	return map[string]Focusable{idPayloadList: &nullFocusable{}}
}

func (p *payloadsComponent) Add(topic, payload string) {
	pi := payloadItem{topic: topic, payload: payload}
	p.items = append(p.items, pi)
	items := append(p.list.Items(), pi)
	p.list.SetItems(items)
}

func (p *payloadsComponent) Items() []payloadItem { return p.items }

func (p *payloadsComponent) SetItems(plds []payloadItem) {
	p.items = plds
	items := make([]list.Item, len(plds))
	for i, pld := range plds {
		items[i] = pld
	}
	p.list.SetItems(items)
}

func (p *payloadsComponent) Clear() { p.SetItems([]payloadItem{}) }
