package help

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/marang/emqutiti/constants"
	"github.com/marang/emqutiti/focus"
)

// ID is the focus identifier for the help component.
const ID = "help"

// Navigator provides navigation control for the help component.
type Navigator interface {
	SetMode(mode constants.AppMode) tea.Cmd
	PreviousMode() constants.AppMode
	Width() int
	Height() int
}

// Focusable re-exports the focus.Focusable interface.
type Focusable = focus.Focusable
