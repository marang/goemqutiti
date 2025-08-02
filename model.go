package emqutiti

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type boxConfig struct {
	height int
}

type layoutConfig struct {
	message boxConfig
	history boxConfig
	topics  boxConfig
	trace   boxConfig
}

type helpState struct {
	vp      viewport.Model
	focused bool
}

func (h *helpState) Focus() tea.Cmd {
	h.focused = true
	return nil
}

func (h *helpState) Blur() { h.focused = false }

func (h helpState) Focused() bool { return h.focused }

func (h helpState) View() string { return "" }

// uiState groups general UI information such as current focus and layout.
type uiState struct {
	focusIndex int            // index of the currently focused element
	modeStack  []appMode      // mode stack, index 0 is current
	width      int            // terminal width
	height     int            // terminal height
	viewport   viewport.Model // scrolling container for the main view
	elemPos    map[string]int // cached Y positions of each box
	focusOrder []string       // order of focusable elements
}

type model struct {
	mqttClient *MQTTClient

	connections  connectionsState
	history      historyState
	topics       topicsState
	message      messageState
	traces       tracesState
	help         helpState
	importWizard *ImportWizard

	ui uiState

	confirmPrompt      string
	confirmInfo        string
	confirmAction      func()
	confirmCancel      func()
	confirmReturnFocus string

	layout layoutConfig

	focusables map[string]Focusable
	focus      *FocusMap
}
