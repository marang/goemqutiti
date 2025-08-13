package importer

import (
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/marang/emqutiti/importer/steps"
	"github.com/marang/emqutiti/ui"
)

// Model runs an interactive import wizard.
type Model struct {
	current steps.Step
	Base    *steps.Base
}

// New creates a new wizard. A non-empty path pre-fills the file field.
func New(client steps.Publisher, path string) *Model {
	b := steps.NewBase(client, path)
	return &Model{current: steps.NewFileStep(b), Base: b}
}

func (m *Model) Init() tea.Cmd { return textinput.Blink }

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if km, ok := msg.(tea.KeyMsg); ok && km.Type == tea.KeyCtrlD {
		return tea.Quit
	}
	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		m.Base.Width = v.Width
		m.Base.Height = v.Height
		m.Base.Progress.Width = v.Width - 4
		m.Base.History.SetSize(v.Width-2, m.Base.HistoryHeight())
		return nil
	case progress.FrameMsg:
		nm, cmd := m.Base.Progress.Update(msg)
		m.Base.Progress = nm.(progress.Model)
		return cmd
	}
	next, cmd := m.current.Update(msg)
	m.current = next
	return cmd
}

// Focus satisfies Component.
func (m *Model) Focus() tea.Cmd { return textinput.Blink }

// Blur satisfies Component.
func (m *Model) Blur() {}

// View renders the wizard at the current step.
func (m *Model) View() string {
	header := m.stepsView()
	bw := m.Base.Width - 2
	if bw <= 0 {
		return header
	}
	wrap := m.Base.Width - 4
	body := m.current.View(bw, wrap)
	return header + "\n\n" + body
}

func (m *Model) stepsView() string {
	var parts []string
	for i, name := range steps.Names {
		st := ui.BlurredStyle
		if i == m.Base.Current {
			st = ui.FocusedStyle
		}
		parts = append(parts, st.Render(name))
	}
	return strings.Join(parts, " > ")
}
