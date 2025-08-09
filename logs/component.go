package logs

import (
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/emqutiti/constants"
	"github.com/marang/emqutiti/ui"
)

// Navigator abstracts navigation needed by the log view.
type Navigator interface {
	SetMode(constants.AppMode) tea.Cmd
	PreviousMode() constants.AppMode
	Width() int
	Height() int
}

// Component renders application logs and implements io.Writer.
type Component struct {
	nav     Navigator
	width   *int
	height  *int
	elemPos *map[string]int
	view    ui.HistoryView
	lines   []string
	focused bool
}

// New creates a log Component sized for the current window.
func New(nav Navigator, width, height *int, elemPos *map[string]int) *Component {
	lv := &Component{
		nav:     nav,
		width:   width,
		height:  height,
		elemPos: elemPos,
		view:    ui.NewHistoryView(*width-2, *height-2),
	}
	return lv
}

// Init implements tea.Model.
func (c *Component) Init() tea.Cmd { return nil }

// Update handles key presses and forwards other messages to the viewport.
func (c *Component) Update(msg tea.Msg) tea.Cmd {
	switch t := msg.(type) {
	case tea.KeyMsg:
		switch t.String() {
		case constants.KeyEsc:
			c.Blur()
			return c.nav.SetMode(c.nav.PreviousMode())
		case constants.KeyCtrlD:
			return tea.Quit
		}
	}
	return c.view.Update(msg)
}

// View renders the log lines inside a LegendBox.
func (c *Component) View() string {
	*c.elemPos = map[string]int{}
	content := c.view.View()
	sp := -1.0
	if c.view.ScrollPercent() >= 0 {
		sp = c.view.ScrollPercent()
	}
	return ui.LegendBox(content, "Logs", *c.width-2, *c.height-2, ui.ColGreen, true, sp)
}

// Focus marks the component as focused and scrolls to the latest line.
func (c *Component) Focus() tea.Cmd {
	c.focused = true
	c.view.GotoBottom()
	return nil
}

// Blur removes focus from the component.
func (c *Component) Blur() { c.focused = false }

// Focused reports whether the component is focused.
func (c *Component) Focused() bool { return c.focused }

// SetSize configures the viewport dimensions.
func (c *Component) SetSize(width, height int) {
	c.view.SetSize(width-2, height-2)
}

// Write appends log data to the view, satisfying io.Writer.
func (c *Component) Write(p []byte) (int, error) {
	if c == nil {
		return len(p), nil
	}
	s := strings.TrimRight(string(p), "\n")
	if s != "" {
		c.lines = append(c.lines, s)
		c.view.SetLines(c.lines)
		if c.focused {
			c.view.GotoBottom()
		}
	}
	return len(p), nil
}

// Lines returns a copy of the logged lines.
func (c *Component) Lines() []string { return append([]string(nil), c.lines...) }

var _ io.Writer = (*Component)(nil)
