package emqutiti

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/emqutiti/constants"
)

type navigator interface {
	SetMode(constants.AppMode) tea.Cmd
	PreviousMode() constants.AppMode
	Width() int
	Height() int
}
