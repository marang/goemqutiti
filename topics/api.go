package topics

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// API exposes topic management behavior to the rest of the application.
type API interface {
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
	Snapshot() []TopicSnapshot
	SetSnapshot([]TopicSnapshot)
}

// Model defines the dependencies Component requires from the host application.
type Model interface {
	ShowClient() tea.Cmd
	SetFocus(id string) tea.Cmd
	FocusedID() string
	StartConfirm(prompt, info string, returnFocus func() tea.Cmd, action func() tea.Cmd, cancel func())
	ResetElemPos()
	SetElemPos(id string, pos int)
	OverlayHelp(view string) string
	ListenStatus() tea.Cmd
	Width() int
}
