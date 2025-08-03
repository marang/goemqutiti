package emqutiti

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// TopicsAPI exposes topic management behavior to the rest of the application.
type TopicsAPI interface {
	HasTopic(string) bool
	SortTopics()
	RebuildActiveTopicList()
	ToggleTopic(index int) tea.Cmd
	RemoveTopic(index int) tea.Cmd
	IndexForPane(pane, idx int) int
	SubscribedItems() []list.Item
	UnsubscribedItems() []list.Item
	TopicAtPosition(x, y int) int
	SetActivePane(idx int)
	SetSelected(int)
	Selected() int
}

// topicsModel defines the dependencies topicsComponent requires from the model.
type topicsModel interface {
	navigator
	SetFocus(id string) tea.Cmd
	FocusedID() string
	StartConfirm(prompt, info string, returnFocus func() tea.Cmd, action func() tea.Cmd, cancel func())
	ResetElemPos()
	SetElemPos(id string, pos int)
	OverlayHelp(view string) string
	ListenStatus() tea.Cmd
}

var _ topicsModel = (*model)(nil)
