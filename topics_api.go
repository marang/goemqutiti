package emqutiti

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type TopicsAPI interface {
	navigator
	SetFocus(id string) tea.Cmd
	FocusedID() string
	StartConfirm(prompt, info string, returnFocus func() tea.Cmd, action func() tea.Cmd, cancel func())
	RemoveTopic(index int) tea.Cmd
	RebuildActiveTopicList()
	ToggleTopic(index int) tea.Cmd
	IndexForPane(pane, idx int) int
	SubscribedItems() []list.Item
	UnsubscribedItems() []list.Item
	ResetElemPos()
	SetElemPos(id string, pos int)
	OverlayHelp(view string) string
	ListenStatus() tea.Cmd
	SetActivePane(idx int)
}

type topicsModel struct{ *model }

func (m *model) topicsAPI() TopicsAPI { return &topicsModel{m} }

func (t *topicsModel) SetFocus(id string) tea.Cmd { return t.setFocus(id) }
func (t *topicsModel) FocusedID() string          { return t.ui.focusOrder[t.ui.focusIndex] }
func (t *topicsModel) StartConfirm(prompt, info string, returnFocus func() tea.Cmd, action func() tea.Cmd, cancel func()) {
	t.startConfirm(prompt, info, returnFocus, action, cancel)
}
func (t *topicsModel) RemoveTopic(index int) tea.Cmd  { return t.removeTopic(index) }
func (t *topicsModel) RebuildActiveTopicList()        { t.rebuildActiveTopicList() }
func (t *topicsModel) ToggleTopic(index int) tea.Cmd  { return t.toggleTopic(index) }
func (t *topicsModel) IndexForPane(pane, idx int) int { return t.indexForPane(pane, idx) }
func (t *topicsModel) SubscribedItems() []list.Item   { return t.subscribedItems() }
func (t *topicsModel) UnsubscribedItems() []list.Item { return t.unsubscribedItems() }
func (t *topicsModel) ResetElemPos()                  { t.ui.elemPos = map[string]int{} }
func (t *topicsModel) SetElemPos(id string, pos int)  { t.ui.elemPos[id] = pos }
func (t *topicsModel) OverlayHelp(view string) string { return t.overlayHelp(view) }
func (t *topicsModel) ListenStatus() tea.Cmd          { return t.connections.ListenStatus() }
func (t *topicsModel) SetActivePane(idx int)          { t.setActivePane(idx) }
