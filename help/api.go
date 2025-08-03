package help

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/marang/emqutiti/topics"
)

// ID is the focus identifier for the help component.
const ID = "help"

// Navigator provides navigation control for the help component.
type Navigator interface {
	SetMode(mode int) tea.Cmd
	PreviousMode() int
	Width() int
	Height() int
}

// Focusable re-exports the topics.Focusable interface.
type Focusable = topics.Focusable
