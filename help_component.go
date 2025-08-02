package emqutiti

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/emqutiti/ui"
)

type helpComponent struct {
	nav     navigator
	vp      viewport.Model
	focused bool
}

func newHelpComponent(nav navigator) *helpComponent {
	return &helpComponent{
		nav: nav,
		vp:  viewport.New(0, 0),
	}
}

func (h *helpComponent) Init() tea.Cmd { return nil }

func (h *helpComponent) Update(msg tea.Msg) tea.Cmd {
	switch t := msg.(type) {
	case tea.KeyMsg:
		switch t.String() {
		case "esc":
			return h.nav.SetMode(h.nav.PreviousMode())
		case "ctrl+d":
			return tea.Quit
		}
	}
	var cmd tea.Cmd
	h.vp, cmd = h.vp.Update(msg)
	return cmd
}

func (h *helpComponent) View() string {
	h.vp.SetContent(helpText)
	content := h.vp.View()
	sp := -1.0
	if h.vp.Height < lipgloss.Height(content) {
		sp = h.vp.ScrollPercent()
	}
	return ui.LegendBox(content, "Help", h.nav.Width()-2, h.nav.Height()-2, ui.ColGreen, true, sp)
}

func (h *helpComponent) Focus() tea.Cmd {
	h.focused = true
	return nil
}

func (h *helpComponent) Blur() { h.focused = false }

func (h *helpComponent) Focused() bool { return h.focused }

// Focusables exposes focusable elements for the help component.
func (h *helpComponent) Focusables() map[string]Focusable {
	return map[string]Focusable{idHelp: adapt(h)}
}
