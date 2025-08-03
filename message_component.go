package emqutiti

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/emqutiti/ui"
)

// messageState holds the textarea model for composing messages.
type messageState struct {
	input textarea.Model
}

func (m *messageState) setPayload(payload string) { m.input.SetValue(payload) }

// messageComponent implements Component for the message editor.
type messageComponent struct {
	*messageState
	m MessageModel
}

func newMessageComponent(m MessageModel, ms messageState) *messageComponent {
	return &messageComponent{messageState: &ms, m: m}
}

func (c *messageComponent) Init() tea.Cmd { return nil }

// Update handles textarea updates when editing messages.
func (c *messageComponent) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	c.input, cmd = c.input.Update(msg)
	return cmd
}

// View renders the message editor box.
func (c *messageComponent) View() string {
	msgContent := c.input.View()
	msgLines := c.input.LineCount()
	msgHeight := c.m.MessageHeight()
	msgSP := -1.0
	if msgLines > msgHeight {
		off := c.input.Line() - msgHeight + 1
		if off < 0 {
			off = 0
		}
		maxOff := msgLines - msgHeight
		if off > maxOff {
			off = maxOff
		}
		if maxOff > 0 {
			msgSP = float64(off) / float64(maxOff)
		}
	}
	focused := c.m.FocusedID() == idMessage
	return ui.LegendBox(msgContent, "Message (Ctrl+S publishes)", c.m.Width()-2, msgHeight, ui.ColBlue, focused, msgSP)
}

func (c *messageComponent) Focus() tea.Cmd { return c.input.Focus() }

func (c *messageComponent) Blur() { c.input.Blur() }

// Focusables exposes focusable elements for the message component.
func (c *messageComponent) Focusables() map[string]Focusable {
	return map[string]Focusable{idMessage: adapt(&c.input)}
}
