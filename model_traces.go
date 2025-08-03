package emqutiti

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// forceStartTrace launches the tracer at index without checking existing data.
func (t *tracesComponent) forceStartTrace(index int) {
	item := t.items[index]
	p, err := LoadProfile(item.cfg.Profile, "")
	if err != nil {
		t.api.LogHistory("", err.Error(), "log", err.Error())
		return
	}
	if p.FromEnv {
		ApplyEnvVars(p)
	} else if env := os.Getenv("EMQUTITI_DEFAULT_PASSWORD"); env != "" {
		p.Password = env
	}
	client, err := NewMQTTClient(*p, nil)
	if err != nil {
		t.api.LogHistory("", err.Error(), "log", err.Error())
		return
	}
	tr := newTracer(item.cfg, client)
	if err := tr.Start(); err != nil {
		t.api.LogHistory("", err.Error(), "log", err.Error())
		client.Disconnect()
		return
	}
	item.tracer = tr
	addTrace(item.cfg)
}

// startTrace starts the tracer at index, prompting if data already exists.
func (t *tracesComponent) startTrace(index int) {
	if index < 0 || index >= len(t.items) {
		return
	}
	item := t.items[index]
	if !item.cfg.End.IsZero() && time.Now().After(item.cfg.End) {
		t.api.LogHistory("", fmt.Sprintf("trace '%s' already finished", item.key), "log", fmt.Sprintf("trace '%s' already finished", item.key))
		return
	}
	exists, err := tracerHasData(item.cfg.Profile, item.key)
	if err == nil && exists {
		rf := func() tea.Cmd { return t.api.SetFocus(t.api.FocusedID()) }
		t.api.StartConfirm(fmt.Sprintf("Overwrite trace '%s'? [y/n]", item.key), "existing trace data will be removed", rf, func() tea.Cmd {
			tracerClearData(item.cfg.Profile, item.key)
			t.forceStartTrace(index)
			return nil
		}, nil)
		return
	}
	t.forceStartTrace(index)
}

// stopTrace stops a running tracer at the given index.
func (t *tracesComponent) stopTrace(index int) {
	if index < 0 || index >= len(t.items) {
		return
	}
	if tr := t.items[index].tracer; tr != nil {
		tr.Stop()
	}
}

// anyTraceRunning reports whether any tracer is currently active or planned.
func (t *tracesComponent) anyTraceRunning() bool {
	for i := range t.items {
		if tr := t.items[i].tracer; tr != nil && (tr.Running() || tr.Planned()) {
			return true
		}
	}
	return false
}

// traceIndex returns the index of the trace with the given key or -1.
func (t *tracesComponent) traceIndex(key string) int {
	for i, it := range t.items {
		if it.key == key {
			return i
		}
	}
	return -1
}

// savePlannedTraces persists trace configurations for later sessions.
func (t *tracesComponent) savePlannedTraces() {
	data := map[string]TracerConfig{}
	for _, it := range t.items {
		if it.tracer != nil {
			data[it.key] = it.tracer.Config()
		} else {
			data[it.key] = it.cfg
		}
	}
	saveTraces(data)
}

// loadTraceMessages loads messages for the trace at index and shows them.
func (t *tracesComponent) loadTraceMessages(index int) {
	if index < 0 || index >= len(t.items) {
		return
	}
	it := t.items[index]
	msgs, err := tracerMessages(it.cfg.Profile, it.key)
	if err != nil {
		t.api.LogHistory("", err.Error(), "log", err.Error())
		return
	}
	items := make([]list.Item, len(msgs))
	for i, mmsg := range msgs {
		items[i] = traceMsgItem{idx: i + 1, msg: mmsg}
	}
	t.view.SetItems(items)
	t.view.SetSize(t.api.Width()-4, t.api.TraceHeight())
	t.viewKey = it.key
	_ = t.api.SetMode(modeViewTrace)
}
