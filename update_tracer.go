package main

import (
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/emqutiti/ui"
)

type traceTickMsg struct{}

// traceTicker schedules periodic refresh events while traces run.
func traceTicker() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg { return traceTickMsg{} })
}

// updateTraces manages the traces list and responds to key presses.
func (m *model) updateTraces(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case traceTickMsg:
		// just refresh
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			m.savePlannedTraces()
			return tea.Quit
		case "esc":
			m.savePlannedTraces()
			cmd := m.setMode(modeClient)
			return cmd
		case "a":
			opts := make([]string, len(m.connections.manager.Profiles))
			for i, p := range m.connections.manager.Profiles {
				opts[i] = p.Name
			}
			topics := []string{}
			for _, t := range m.topics.items {
				if t.subscribed {
					topics = append(topics, t.title)
				}
			}
			f := newTraceForm(opts, m.connections.active, topics)
			m.traces.form = &f
			cmd := m.setMode(modeEditTrace)
			return tea.Batch(cmd, textinput.Blink)
		case "enter":
			i := m.traces.list.Index()
			if i >= 0 && i < len(m.traces.items) {
				it := m.traces.items[i]
				if it.tracer != nil && (it.tracer.Running() || it.tracer.Planned()) {
					m.stopTrace(i)
				} else {
					m.startTrace(i)
				}
			}
		case "v":
			i := m.traces.list.Index()
			if i >= 0 && i < len(m.traces.items) {
				m.loadTraceMessages(i)
				return nil
			}
		case "delete":
			i := m.traces.list.Index()
			if i >= 0 && i < len(m.traces.items) {
				key := m.traces.items[i].key
				m.stopTrace(i)
				m.traces.items = append(m.traces.items[:i], m.traces.items[i+1:]...)
				items := []list.Item{}
				for _, it := range m.traces.items {
					items = append(items, it)
				}
				m.traces.list.SetItems(items)
				removeTrace(key)
			}
			if m.anyTraceRunning() {
				return traceTicker()
			}
			return nil
		}
	}
	m.traces.list, cmd = m.traces.list.Update(msg)
	if m.anyTraceRunning() {
		return tea.Batch(cmd, traceTicker())
	}
	return cmd
}

// updateTraceForm handles input for the new trace form.
func (m *model) updateTraceForm(msg tea.Msg) tea.Cmd {
	if m.traces.form == nil {
		return nil
	}
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			return tea.Quit
		case "esc":
			m.traces.form = nil
			cmd := m.setMode(modeTracer)
			return cmd
		case "enter":
			cfg := m.traces.form.Config()
			if cfg.Key == "" || len(cfg.Topics) == 0 || cfg.Profile == "" {
				m.traces.form.errMsg = "key, profile and topics required"
				return nil
			}
			if cfg.Start.IsZero() {
				cfg.Start = time.Now().Round(time.Second)
				if tf, ok := m.traces.form.Fields[idxTraceStart].(*ui.TextField); ok {
					tf.SetValue(cfg.Start.Format(time.RFC3339))
				}
			}
			if cfg.End.IsZero() {
				cfg.End = cfg.Start.Add(time.Hour)
				if tf, ok := m.traces.form.Fields[idxTraceEnd].(*ui.TextField); ok {
					tf.SetValue(cfg.End.Format(time.RFC3339))
				}
			}
			if m.traceIndex(cfg.Key) >= 0 {
				m.traces.form.errMsg = "trace key exists"
				return nil
			}
			p, err := LoadProfile(cfg.Profile, "")
			if err != nil {
				m.traces.form.errMsg = err.Error()
				return nil
			}
			if p.FromEnv {
				ApplyEnvVars(p)
			} else if env := os.Getenv("EMQUTITI_DEFAULT_PASSWORD"); env != "" {
				p.Password = env
			}
			client, err := NewMQTTClient(*p, nil)
			if err != nil {
				m.traces.form.errMsg = err.Error()
				return nil
			}
			client.Disconnect()
			newItem := &traceItem{key: cfg.Key, cfg: cfg}
			m.traces.items = append(m.traces.items, newItem)
			items := m.traces.list.Items()
			items = append(items, newItem)
			m.traces.list.SetItems(items)
			addTrace(cfg)
			m.traces.form = nil
			cmd := m.setMode(modeTracer)
			return cmd
		}
	}
	f, cmd := m.traces.form.Update(msg)
	m.traces.form = &f
	return cmd
}

// updateTraceView displays messages captured for a trace.
func (m *model) updateTraceView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			cmd := m.setMode(modeTracer)
			return cmd
		case "ctrl+d":
			return tea.Quit
		case "ctrl+shift+up":
			if m.layout.trace.height > 1 {
				m.layout.trace.height--
				m.traces.view.SetSize(m.ui.width-4, m.layout.trace.height)
			}
		case "ctrl+shift+down":
			m.layout.trace.height++
			m.traces.view.SetSize(m.ui.width-4, m.layout.trace.height)
		}
	}
	m.traces.view, cmd = m.traces.view.Update(msg)
	return cmd
}
