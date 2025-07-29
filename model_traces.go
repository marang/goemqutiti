package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"

	"github.com/marang/goemqutiti/config"
)

func (m *model) forceStartTrace(index int) {
	item := m.traces.items[index]
	p, err := config.LoadProfile(item.cfg.Profile, "")
	if err != nil {
		m.appendHistory("", err.Error(), "log", err.Error())
		return
	}
	if p.FromEnv {
		config.ApplyEnvVars(p)
	} else if env := os.Getenv("MQTT_PASSWORD"); env != "" {
		p.Password = env
	}
	client, err := NewMQTTClient(*p, nil)
	if err != nil {
		m.appendHistory("", err.Error(), "log", err.Error())
		return
	}
	tr := newTracer(item.cfg, client)
	if err := tr.Start(); err != nil {
		m.appendHistory("", err.Error(), "log", err.Error())
		client.Disconnect()
		return
	}
	item.tracer = tr
	addTrace(item.cfg)
}

func (m *model) startTrace(index int) {
	if index < 0 || index >= len(m.traces.items) {
		return
	}
	item := m.traces.items[index]
	if !item.cfg.End.IsZero() && time.Now().After(item.cfg.End) {
		m.appendHistory("", fmt.Sprintf("trace '%s' already finished", item.key), "log", fmt.Sprintf("trace '%s' already finished", item.key))
		return
	}
	exists, err := tracerHasData(item.cfg.Profile, item.key)
	if err == nil && exists {
		m.startConfirm(fmt.Sprintf("Overwrite trace '%s'? [y/n]", item.key), "existing trace data will be removed", func() {
			tracerClearData(item.cfg.Profile, item.key)
			m.forceStartTrace(index)
		})
		return
	}
	m.forceStartTrace(index)
}

func (m *model) stopTrace(index int) {
	if index < 0 || index >= len(m.traces.items) {
		return
	}
	if tr := m.traces.items[index].tracer; tr != nil {
		tr.Stop()
	}
}

func (m *model) anyTraceRunning() bool {
	for i := range m.traces.items {
		if tr := m.traces.items[i].tracer; tr != nil && (tr.Running() || tr.Planned()) {
			return true
		}
	}
	return false
}

func (m *model) traceIndex(key string) int {
	for i, it := range m.traces.items {
		if it.key == key {
			return i
		}
	}
	return -1
}

func (m *model) savePlannedTraces() {
	data := map[string]TracerConfig{}
	for _, it := range m.traces.items {
		if it.tracer != nil {
			data[it.key] = it.tracer.Config()
		} else {
			data[it.key] = it.cfg
		}
	}
	saveTraces(data)
}
func (m *model) loadTraceMessages(index int) {
	if index < 0 || index >= len(m.traces.items) {
		return
	}
	it := m.traces.items[index]
	msgs, err := tracerMessages(it.cfg.Profile, it.key)
	if err != nil {
		m.appendHistory("", err.Error(), "log", err.Error())
		return
	}
	items := make([]list.Item, len(msgs))
	for i, mmsg := range msgs {
		items[i] = traceMsgItem{idx: i + 1, msg: mmsg}
	}
	m.traces.view.SetItems(items)
	m.traces.view.SetSize(m.ui.width-4, m.layout.trace.height)
	m.traces.viewKey = it.key
	m.ui.mode = modeViewTrace
}
