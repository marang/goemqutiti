package emqutiti

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/emqutiti/importer"
)

// Component defines a screen or feature that can participate in the Tea update
// and view cycle. Implementations handle their own initialization, state
// updates, rendering and focus management.
type Component interface {
	Init() tea.Cmd
	Update(tea.Msg) tea.Cmd
	View() string
	Focus() tea.Cmd
	Blur()
}

// component is a lightweight adapter that allows plain functions to satisfy
// the Component interface. Missing functions simply result in no-ops.
type component struct {
	init   func() tea.Cmd
	update func(tea.Msg) tea.Cmd
	view   func() string
	focus  func() tea.Cmd
	blur   func()
}

func (c component) Init() tea.Cmd {
	if c.init != nil {
		return c.init()
	}
	return nil
}

func (c component) Update(msg tea.Msg) tea.Cmd {
	if c.update != nil {
		return c.update(msg)
	}
	return nil
}

func (c component) View() string {
	if c.view != nil {
		return c.view()
	}
	return ""
}

func (c component) Focus() tea.Cmd {
	if c.focus != nil {
		return c.focus()
	}
	return nil
}

func (c component) Blur() {
	if c.blur != nil {
		c.blur()
	}
}

type boxConfig struct {
	height int
}

type layoutConfig struct {
	message boxConfig
	history boxConfig
	topics  boxConfig
	trace   boxConfig
}

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

	connections connectionsState
	history     historyState
	topics      topicsState
	message     messageState
	traces      tracesState
	help        *helpComponent
	importer    *importer.Model

	ui uiState

	confirm *confirmComponent

	layout layoutConfig

	// components maps each application mode to its corresponding component
	// implementation. These components handle mode-specific update and view
	// logic which the model delegates to at runtime.
	components map[appMode]Component

	focusables map[string]Focusable
	focus      *FocusMap
}
