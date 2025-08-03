package emqutiti

import tea "github.com/charmbracelet/bubbletea"

// tracesComponent wraps the existing trace functionality so that it satisfies
// the Component interface. It delegates update and view logic to the model
// but allows tests to supply custom implementations of trace persistence.
type tracesComponent struct {
	m     *model
	store TraceStore
}

// newTracesComponent creates a tracesComponent using the provided navigator
// and TraceStore implementation. The navigator is expected to be the model
// itself, which holds the tracing state.
func newTracesComponent(nav navigator, store TraceStore) *tracesComponent {
	return &tracesComponent{m: nav.(*model), store: store}
}

func (t *tracesComponent) Init() tea.Cmd { return nil }

func (t *tracesComponent) Update(msg tea.Msg) tea.Cmd { return t.m.updateTraces(msg) }

func (t *tracesComponent) View() string { return t.m.viewTraces() }

func (t *tracesComponent) Focus() tea.Cmd { return nil }

func (t *tracesComponent) Blur() {}

// Focusables satisfies the FocusableSet interface but the trace list focusable
// is provided by the model's base Focusables so this returns an empty map.
func (t *tracesComponent) Focusables() map[string]Focusable { return map[string]Focusable{} }
