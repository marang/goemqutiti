package help

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/emqutiti/constants"
	"github.com/marang/emqutiti/focus"
	"github.com/marang/emqutiti/ui"
)

// Component renders the application help screen.
type Component struct {
	nav     Navigator
	width   *int
	height  *int
	elemPos *map[string]int
	vp      viewport.Model
	focused bool
}

// New creates a help Component.
func New(nav Navigator, width, height *int, elemPos *map[string]int) *Component {
	return &Component{
		nav:     nav,
		width:   width,
		height:  height,
		elemPos: elemPos,
		vp:      viewport.New(0, 0),
	}
}

// Init implements tea.Model.
func (h *Component) Init() tea.Cmd { return nil }

// Update handles key events and viewport updates.
func (h *Component) Update(msg tea.Msg) tea.Cmd {
	switch t := msg.(type) {
	case tea.KeyMsg:
		switch t.String() {
		case constants.KeyEsc:
			return h.nav.SetMode(h.nav.PreviousMode())
		case constants.KeyCtrlD:
			return tea.Quit
		}
	}
	var cmd tea.Cmd
	h.vp, cmd = h.vp.Update(msg)
	return cmd
}

// View renders the help content.
func (h *Component) View() string {
	*h.elemPos = map[string]int{}
	h.vp.SetContent(helpText)
	content := h.vp.View()
	sp := -1.0
	if h.vp.Height < lipgloss.Height(content) {
		sp = h.vp.ScrollPercent()
	}
	return ui.LegendBox(content, "Help", *h.width-2, *h.height-2, ui.ColGreen, true, sp)
}

// SetSize configures the viewport dimensions.
func (h *Component) SetSize(width, height int) {
	h.vp.Width = width - 4
	h.vp.Height = height - 4
}

// Focus marks the component as focused.
func (h *Component) Focus() tea.Cmd {
	h.focused = true
	return nil
}

// Blur removes focus from the component.
func (h *Component) Blur() { h.focused = false }

// Focused reports whether the component is focused.
func (h *Component) Focused() bool { return h.focused }

// Focusables exposes focusable elements for the help component.
func (h *Component) Focusables() map[string]focus.Focusable {
	return map[string]focus.Focusable{ID: focus.Adapt(h)}
}
