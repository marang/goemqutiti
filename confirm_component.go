package emqutiti

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/emqutiti/ui"
)

type confirmComponent struct {
	m *model

	prompt string
	info   string
	action func()
	cancel func()

	returnFocus string
	focused     bool
}

func newConfirmComponent(m *model) *confirmComponent {
	return &confirmComponent{m: m}
}

func (c *confirmComponent) Init() tea.Cmd { return nil }

func (c *confirmComponent) start(prompt, info string, action func()) {
	c.prompt = prompt
	c.info = info
	c.action = action
	c.cancel = nil
	_ = c.m.setMode(modeConfirmDelete)
}

func (c *confirmComponent) Update(msg tea.Msg) tea.Cmd {
	switch t := msg.(type) {
	case tea.KeyMsg:
		switch t.String() {
		case "ctrl+d":
			return tea.Quit
		case "y":
			if c.action != nil {
				c.action()
				c.action = nil
			}
			if c.cancel != nil {
				c.cancel = nil
			}
			cmd := c.m.setMode(c.m.previousMode())
			cmds := []tea.Cmd{cmd, c.m.connections.ListenStatus()}
			if c.returnFocus != "" {
				cmds = append(cmds, c.m.setFocus(c.returnFocus))
				c.returnFocus = ""
			} else {
				c.m.scrollToFocused()
			}
			return tea.Batch(cmds...)
		case "n", "esc":
			if c.cancel != nil {
				c.cancel()
				c.cancel = nil
			}
			cmd := c.m.setMode(c.m.previousMode())
			cmds := []tea.Cmd{cmd, c.m.connections.ListenStatus()}
			if c.returnFocus != "" {
				cmds = append(cmds, c.m.setFocus(c.returnFocus))
				c.returnFocus = ""
			} else {
				c.m.scrollToFocused()
			}
			return tea.Batch(cmds...)
		}
	}
	return c.m.connections.ListenStatus()
}

func (c *confirmComponent) View() string {
	c.m.ui.elemPos = map[string]int{}
	content := c.prompt
	if c.info != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, c.prompt, c.info)
	}
	content = lipgloss.NewStyle().Padding(1, 2).Render(content)
	box := ui.LegendBox(content, "Confirm", c.m.ui.width/2, 0, ui.ColBlue, true, -1)
	return lipgloss.Place(c.m.ui.width, c.m.ui.height, lipgloss.Center, lipgloss.Center, box)
}

func (c *confirmComponent) Focus() tea.Cmd {
	c.focused = true
	return nil
}

func (c *confirmComponent) Blur() { c.focused = false }

func (c *confirmComponent) Focused() bool { return c.focused }

// Focusables exposes focusable elements for the confirm component.
func (c *confirmComponent) Focusables() map[string]Focusable { return map[string]Focusable{} }
