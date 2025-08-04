package topics

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/marang/emqutiti/ui"
)

func initTopics() state {
	ti := textinput.New()
	ti.Placeholder = "Enter Topic"
	ti.CharLimit = 32
	ti.Prompt = "> "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(ui.ColGray)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(ui.ColGray)
	ti.Cursor.Style = ui.CursorStyle
	ti.TextStyle = ui.FocusedStyle
	ti.Width = 0
	topicsList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	topicsList.DisableQuitKeybindings()
	topicsList.SetShowTitle(false)
	ts := state{
		Input: ti,
		Items: []Item{},
		list:  topicsList,
		panes: topicsPanes{
			subscribed:   paneState{sel: 0, page: 0, index: 0},
			unsubscribed: paneState{sel: 0, page: 0, index: 1},
			active:       0,
		},
		selected:   -1,
		ChipBounds: []ChipBound{},
		VP:         viewport.New(0, 0),
	}
	return ts
}
