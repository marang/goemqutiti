package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type traceTickMsg struct{}

func traceTicker() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg { return traceTickMsg{} })
}

func (m model) updateTraces(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case traceTickMsg:
		// just refresh
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			m.savePlannedTraces()
			return m, tea.Quit
		case "esc":
			m.savePlannedTraces()
			m.ui.mode = modeClient
		case "enter":
			i := m.traces.list.Index()
			if i >= 0 && i < len(m.traces.items) {
				it := &m.traces.items[i]
				if it.tracer != nil && (it.tracer.Running() || it.tracer.Planned()) {
					m.stopTrace(i)
				} else {
					m.startTrace(i)
				}
			}
		}
	}
	m.traces.list, cmd = m.traces.list.Update(msg)
	if m.anyTraceRunning() {
		return m, tea.Batch(cmd, traceTicker())
	}
	return m, cmd
}
