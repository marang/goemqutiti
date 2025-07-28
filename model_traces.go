package main

import (
	"os"

	"github.com/charmbracelet/bubbles/list"

	"github.com/marang/goemqutiti/config"
	"github.com/marang/goemqutiti/history"
	"github.com/marang/goemqutiti/tracer"
)

func (m *model) startTrace(index int) {
	if index < 0 || index >= len(m.traces.items) {
		return
	}
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
	tr := tracer.New(item.cfg, client)
	if err := tr.Start(); err != nil {
		m.appendHistory("", err.Error(), "log", err.Error())
		client.Disconnect()
		return
	}
	item.tracer = tr
	addTrace(item.cfg)
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

func (m *model) savePlannedTraces() {
	data := map[string]tracer.Config{}
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
	idx, err := history.OpenTrace(it.cfg.Profile)
	if err != nil {
		m.appendHistory("", err.Error(), "log", err.Error())
		return
	}
	msgs, err := idx.TraceMessages(it.key)
	idx.Close()
	if err != nil {
		m.appendHistory("", err.Error(), "log", err.Error())
		return
	}
	items := make([]list.Item, len(msgs))
	for i, mmsg := range msgs {
		items[i] = historyItem{topic: mmsg.Topic, payload: mmsg.Payload, kind: "trace"}
	}
	m.traces.view.SetItems(items)
	m.traces.view.SetSize(m.ui.width-4, m.ui.height-4)
	m.traces.viewKey = it.key
	m.ui.mode = modeViewTrace
}
