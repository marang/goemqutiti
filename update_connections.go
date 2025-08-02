package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// updateConnections processes input when the connections view is active.
func (m model) updateConnections(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case connectResult:
		brokerURL := fmt.Sprintf("%s://%s:%d", msg.profile.Schema, msg.profile.Host, msg.profile.Port)
		if msg.err != nil {
			m.connections.manager.Statuses[msg.profile.Name] = "disconnected"
			m.connections.manager.Errors[msg.profile.Name] = fmt.Sprintf("Failed to connect to %s: %v", brokerURL, msg.err)
			m.connections.connection = fmt.Sprintf("Failed to connect to %s: %v", brokerURL, msg.err)
			m.refreshConnectionItems()
		} else {
			m.mqttClient = msg.client
			m.connections.active = msg.profile.Name
			if m.history.store != nil {
				m.history.store.Close()
			}
			if idx, err := openHistoryStore(msg.profile.Name); err == nil {
				m.history.store = idx
				msgs := idx.Search(false, nil, time.Time{}, time.Time{}, "")
				m.history.items = nil
				items := make([]list.Item, len(msgs))
				for i, mmsg := range msgs {
					items[i] = historyItem{timestamp: mmsg.Timestamp, topic: mmsg.Topic, payload: mmsg.Payload, kind: mmsg.Kind}
					m.history.items = append(m.history.items, items[i].(historyItem))
				}
				m.history.list.SetItems(items)
			}
			m.restoreState(msg.profile.Name)
			m.subscribeActiveTopics()
			m.connections.connection = "Connected to " + brokerURL
			m.connections.manager.Statuses[msg.profile.Name] = "connected"
			m.connections.manager.Errors[msg.profile.Name] = ""
			m.refreshConnectionItems()
			cmd := m.setMode(modeClient)
			return m, tea.Batch(cmd, listenStatus(m.connections.statusChan))
		}
		return m, listenStatus(m.connections.statusChan)
	case tea.KeyMsg:
		if m.connections.manager.ConnectionsList.FilterState() == list.Filtering {
			switch msg.String() {
			case "enter":
				i := m.connections.manager.ConnectionsList.Index()
				if i >= 0 && i < len(m.connections.manager.Profiles) {
					p := m.connections.manager.Profiles[i]
					if p.Name == m.connections.active && m.connections.manager.Statuses[p.Name] == "connected" {
						brokerURL := fmt.Sprintf("%s://%s:%d", p.Schema, p.Host, p.Port)
						m.connections.connection = "Connected to " + brokerURL
						m.connections.manager.Errors[p.Name] = ""
						m.refreshConnectionItems()
						cmd := m.setMode(modeClient)
						return m, cmd
					}
					flushStatus(m.connections.statusChan)
					if p.FromEnv {
						ApplyEnvVars(&p)
					} else if env := os.Getenv("EMQUTITI_DEFAULT_PASSWORD"); env != "" {
						p.Password = env
					}
					m.connections.manager.Errors[p.Name] = ""
					m.connections.manager.Statuses[p.Name] = "connecting"
					brokerURL := fmt.Sprintf("%s://%s:%d", p.Schema, p.Host, p.Port)
					m.connections.connection = "Connecting to " + brokerURL
					m.refreshConnectionItems()
					return m, connectBroker(p, m.connections.statusChan)
				}
			}
			break
		}
		switch msg.String() {
		case "ctrl+d":
			return m, tea.Quit
		case "ctrl+r":
			m.traces.list.SetSize(m.ui.width-4, m.ui.height-4)
			cmd := m.setMode(modeTracer)
			return m, cmd
		case "a":
			f := newConnectionForm(Profile{}, -1)
			m.connections.form = &f
			cmd := m.setMode(modeEditConnection)
			return m, cmd
		case "e":
			i := m.connections.manager.ConnectionsList.Index()
			if i >= 0 && i < len(m.connections.manager.Profiles) {
				f := newConnectionForm(m.connections.manager.Profiles[i], i)
				m.connections.form = &f
				cmd := m.setMode(modeEditConnection)
				return m, cmd
			}
		case "enter":
			i := m.connections.manager.ConnectionsList.Index()
			if i >= 0 && i < len(m.connections.manager.Profiles) {
				p := m.connections.manager.Profiles[i]
				if p.Name == m.connections.active && m.connections.manager.Statuses[p.Name] == "connected" {
					brokerURL := fmt.Sprintf("%s://%s:%d", p.Schema, p.Host, p.Port)
					m.connections.connection = "Connected to " + brokerURL
					m.connections.manager.Errors[p.Name] = ""
					m.refreshConnectionItems()
					cmd := m.setMode(modeClient)
					return m, cmd
				}
				flushStatus(m.connections.statusChan)
				if p.FromEnv {
					ApplyEnvVars(&p)
				} else if env := os.Getenv("EMQUTITI_DEFAULT_PASSWORD"); env != "" {
					p.Password = env
				}
				m.connections.manager.Errors[p.Name] = ""
				m.connections.manager.Statuses[p.Name] = "connecting"
				brokerURL := fmt.Sprintf("%s://%s:%d", p.Schema, p.Host, p.Port)
				m.connections.connection = "Connecting to " + brokerURL
				m.refreshConnectionItems()
				return m, connectBroker(p, m.connections.statusChan)
			}
		case "delete":
			i := m.connections.manager.ConnectionsList.Index()
			if i >= 0 {
				name := m.connections.manager.Profiles[i].Name
				info := "This also deletes history and traces"
				m.confirmReturnFocus = m.ui.focusOrder[m.ui.focusIndex]
				m.startConfirm(fmt.Sprintf("Delete broker '%s'? [y/n]", name), info, func() {
					m.connections.manager.DeleteConnection(i)
					m.connections.manager.refreshList()
					m.refreshConnectionItems()
				})
				return m, listenStatus(m.connections.statusChan)
			}
		case "x":
			if m.mqttClient != nil {
				m.mqttClient.Disconnect()
				m.connections.manager.Statuses[m.connections.active] = "disconnected"
				m.connections.manager.Errors[m.connections.active] = ""
				m.refreshConnectionItems()
				m.connections.connection = ""
				m.connections.active = ""
				m.mqttClient = nil
			}
		}
	}
	m.connections.manager.ConnectionsList, cmd = m.connections.manager.ConnectionsList.Update(msg)
	return m, tea.Batch(cmd, listenStatus(m.connections.statusChan))
}
