package topics

import tea "github.com/charmbracelet/bubbletea"

type Focusable interface {
	Focus()
	Blur()
	IsFocused() bool
	View() string
}

type teaFocusable interface {
	Focus() tea.Cmd
	Blur()
	Focused() bool
	View() string
}

func adapt(f teaFocusable) Focusable { return focusAdapter{f} }

type focusAdapter struct{ f teaFocusable }

func (a focusAdapter) Focus()          { _ = a.f.Focus() }
func (a focusAdapter) Blur()           { a.f.Blur() }
func (a focusAdapter) IsFocused() bool { return a.f.Focused() }
func (a focusAdapter) View() string    { return a.f.View() }

const (
	idTopicsEnabled  = "topics-enabled"
	idTopicsDisabled = "topics-disabled"
	idTopic          = "topic"
)

// Item represents a topic and its subscription state.
type Item struct {
	Name       string
	Subscribed bool
}

func (t Item) FilterValue() string { return t.Name }
func (t Item) Title() string       { return t.Name }
func (t Item) Description() string {
	if t.Subscribed {
		return "subscribed"
	}
	return "unsubscribed"
}

type ChipBound struct {
	XPos, YPos    int
	Width, Height int
}

type paneManager interface{ SetActivePane(int) }

type paneState struct {
	sel     int
	page    int
	m       paneManager
	index   int
	focused bool
}

func (p *paneState) Focus() {
	p.focused = true
	if p.m != nil {
		p.m.SetActivePane(p.index)
	}
}

func (p *paneState) Blur() { p.focused = false }

func (p paneState) IsFocused() bool { return p.focused }

func (p paneState) View() string { return "" }

type topicsPanes struct {
	subscribed   paneState
	unsubscribed paneState
	active       int
}
