package confirm

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/emqutiti/ui"
)

type Component struct {
	nav    Navigator
	status StatusListener

	prompt      string
	info        string
	action      func() tea.Cmd
	cancel      func()
	returnFocus func() tea.Cmd
	focused     bool
}

func NewComponent(nav Navigator, status StatusListener, returnFocus func() tea.Cmd, action func() tea.Cmd, cancel func()) *Component {
	return &Component{nav: nav, status: status, returnFocus: returnFocus, action: action, cancel: cancel}
}

func (c *Component) Init() tea.Cmd { return nil }

func (c *Component) Start(prompt, info string) {
	c.prompt = prompt
	c.info = info
	_ = c.nav.SetConfirmMode()
}

func (c *Component) Update(msg tea.Msg) tea.Cmd {
	switch t := msg.(type) {
	case tea.KeyMsg:
		switch t.String() {
		case "ctrl+d":
			return tea.Quit
		case "y":
			var acmd tea.Cmd
			if c.action != nil {
				acmd = c.action()
				c.action = nil
			}
			if c.cancel != nil {
				c.cancel = nil
			}
			cmd := c.nav.SetPreviousMode()
			cmds := []tea.Cmd{cmd, c.status.ListenStatus()}
			if acmd != nil {
				cmds = append(cmds, acmd)
			}
			if c.returnFocus != nil {
				cmds = append(cmds, c.returnFocus())
				c.returnFocus = nil
			} else {
				c.nav.ScrollToFocused()
			}
			return tea.Batch(cmds...)
		case "n", "esc":
			if c.cancel != nil {
				c.cancel()
				c.cancel = nil
			}
			cmd := c.nav.SetPreviousMode()
			cmds := []tea.Cmd{cmd, c.status.ListenStatus()}
			if c.returnFocus != nil {
				cmds = append(cmds, c.returnFocus())
				c.returnFocus = nil
			} else {
				c.nav.ScrollToFocused()
			}
			return tea.Batch(cmds...)
		}
	}
	return c.status.ListenStatus()
}

func (c *Component) View() string {
	content := c.prompt
	if c.info != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, c.prompt, c.info)
	}
	content = lipgloss.NewStyle().Padding(1, 2).Render(content)
	box := ui.LegendBox(content, "Confirm", c.nav.Width()/2, 0, ui.ColBlue, true, -1)
	return lipgloss.Place(c.nav.Width(), c.nav.Height(), lipgloss.Center, lipgloss.Center, box)
}

func (c *Component) Focus() tea.Cmd {
	c.focused = true
	return nil
}

func (c *Component) Blur() { c.focused = false }

func (c *Component) Focused() bool { return c.focused }
