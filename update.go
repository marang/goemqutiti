package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/goemqutiti/config"
)

type statusMessage string

type connectResult struct {
	client  *MQTTClient
	profile Profile
	err     error
}

// connectBroker attempts to connect to the given profile and reports status on the channel.
func connectBroker(p Profile, ch chan string) tea.Cmd {
	return func() tea.Msg {
		if ch != nil {
			brokerURL := fmt.Sprintf("%s://%s:%d", p.Schema, p.Host, p.Port)
			ch <- fmt.Sprintf("Connecting to %s", brokerURL)
		}
		client, err := NewMQTTClient(p, ch)
		return connectResult{client: client, profile: p, err: err}
	}
}

// listenMessages waits for incoming MQTT messages on the provided channel.
func listenMessages(ch chan MQTTMessage) tea.Cmd {
	return func() tea.Msg {
		if ch == nil {
			return nil
		}
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

// listenStatus retrieves status updates from the status channel.
func listenStatus(ch chan string) tea.Cmd {
	return func() tea.Msg {
		if ch == nil {
			return nil
		}
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return statusMessage(msg)
	}
}

// flushStatus drains all pending messages from the status channel.
func flushStatus(ch chan string) {
	if ch == nil {
		return
	}
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}

// saveCurrent persists topics and payloads for the active connection.
func (m *model) saveCurrent() {
	if m.connections.active == "" {
		return
	}
	m.connections.saved[m.connections.active] = connectionData{Topics: m.topics.items, Payloads: m.message.payloads}
	saveState(m.connections.saved)
}

// restoreState loads saved state for the named connection.
func (m *model) restoreState(name string) {
	if data, ok := m.connections.saved[name]; ok {
		m.topics.items = data.Topics
		m.message.payloads = data.Payloads
		m.sortTopics()
		m.rebuildActiveTopicList()
	} else {
		m.topics.items = []topicItem{}
		m.message.payloads = []payloadItem{}
	}
}

// appendHistory stores a message in the history list and optional store.
func (m *model) appendHistory(topic, payload, kind, logText string) {
	ts := time.Now()
	text := payload
	if kind == "log" {
		text = logText
	}
	hi := historyItem{timestamp: ts, topic: topic, payload: text, kind: kind, archived: false}
	if m.history.store != nil {
		m.history.store.Add(Message{Timestamp: ts, Topic: topic, Payload: payload, Kind: kind, Archived: false})
	}
	if !m.history.showArchived {
		if m.history.filterQuery != "" {
			topics, start, end, pf := parseHistoryQuery(m.history.filterQuery)
			var msgs []Message
			if m.history.showArchived {
				msgs = m.history.store.SearchArchived(topics, start, end, pf)
			} else {
				msgs = m.history.store.Search(topics, start, end, pf)
			}
			items := make([]list.Item, len(msgs))
			m.history.items = make([]historyItem, len(msgs))
			for i, mmsg := range msgs {
				hi := historyItem{timestamp: mmsg.Timestamp, topic: mmsg.Topic, payload: mmsg.Payload, kind: mmsg.Kind, archived: mmsg.Archived}
				items[i] = hi
				m.history.items[i] = hi
			}
			m.history.list.SetItems(items)
			m.history.list.Select(len(items) - 1)
		} else {
			m.history.items = append(m.history.items, hi)
			items := make([]list.Item, len(m.history.items))
			for i, it := range m.history.items {
				items[i] = it
			}
			m.history.list.SetItems(items)
			m.history.list.Select(len(items) - 1)
		}
	}
}

// setFocus moves focus to the given element id.
func (m *model) setFocus(id string) tea.Cmd {
	for i, name := range m.ui.focusOrder {
		if name == id {
			m.focus.Set(i)
			m.ui.focusIndex = m.focus.Index()
			break
		}
	}
	m.scrollToFocused()
	return nil
}

// focusFromMouse determines which element was clicked and focuses it.
func (m *model) focusFromMouse(y int) tea.Cmd {
	cy := y + m.ui.viewport.YOffset - 1
	chosen := ""
	maxPos := -1
	for _, id := range m.ui.focusOrder {
		if pos, ok := m.ui.elemPos[id]; ok && cy >= pos && pos > maxPos {
			chosen = id
			maxPos = pos
		}
	}
	if chosen != "" {
		if chosen != m.ui.focusOrder[m.ui.focusIndex] {
			return m.setFocus(chosen)
		}
		return nil
	}
	if len(m.ui.focusOrder) > 0 && m.ui.focusOrder[m.ui.focusIndex] != m.ui.focusOrder[0] {
		return m.setFocus(m.ui.focusOrder[0])
	}
	return nil
}

// scrollToFocused ensures the focused element is visible in the viewport.
func (m *model) scrollToFocused() {
	if len(m.ui.focusOrder) == 0 {
		return
	}
	id := m.ui.focusOrder[m.ui.focusIndex]
	pos, ok := m.ui.elemPos[id]
	if !ok {
		return
	}
	offset := pos - 1
	if offset < 0 {
		offset = 0
	}
	if offset < m.ui.viewport.YOffset {
		m.ui.viewport.SetYOffset(offset)
	} else if offset >= m.ui.viewport.YOffset+m.ui.viewport.Height {
		m.ui.viewport.SetYOffset(offset - m.ui.viewport.Height + 1)
	}
}

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
				msgs := idx.Search(nil, time.Time{}, time.Time{}, "")
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
						config.ApplyEnvVars(&p)
					} else if env := os.Getenv("MQTT_PASSWORD"); env != "" {
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
					config.ApplyEnvVars(&p)
				} else if env := os.Getenv("MQTT_PASSWORD"); env != "" {
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

// updateForm handles the add/edit connection form.
func (m model) updateForm(msg tea.Msg) (model, tea.Cmd) {
	if m.connections.form == nil {
		return m, nil
	}
	var cmd tea.Cmd
	switch msg.(type) {
	case tea.WindowSizeMsg, tea.MouseMsg:
		m.connections.manager.ConnectionsList, _ = m.connections.manager.ConnectionsList.Update(msg)
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			return m, tea.Quit
		case "esc":
			cmd := m.setMode(modeConnections)
			m.connections.form = nil
			return m, cmd
		case "enter":
			p := m.connections.form.Profile()
			if m.connections.form.index >= 0 {
				m.connections.manager.EditConnection(m.connections.form.index, p)
			} else {
				m.connections.manager.AddConnection(p)
			}
			m.refreshConnectionItems()
			cmd := m.setMode(modeConnections)
			m.connections.form = nil
			return m, cmd
		}
	}
	f, cmd := m.connections.form.Update(msg)
	m.connections.form = &f
	return m, tea.Batch(cmd, listenStatus(m.connections.statusChan))
}

// updateConfirmDelete processes confirmation dialog key presses.
func (m *model) updateConfirmDelete(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			return *m, tea.Quit
		case "y":
			if m.confirmAction != nil {
				m.confirmAction()
				m.confirmAction = nil
			}
			if m.confirmCancel != nil {
				m.confirmCancel = nil
			}
			cmd := m.setMode(m.previousMode())
			cmds := []tea.Cmd{cmd, listenStatus(m.connections.statusChan)}
			if m.confirmReturnFocus != "" {
				cmds = append(cmds, m.setFocus(m.confirmReturnFocus))
				m.confirmReturnFocus = ""
			} else {
				m.scrollToFocused()
			}
			return *m, tea.Batch(cmds...)
		case "n", "esc":
			if m.confirmCancel != nil {
				m.confirmCancel()
				m.confirmCancel = nil
			}
			cmd := m.setMode(m.previousMode())
			cmds := []tea.Cmd{cmd, listenStatus(m.connections.statusChan)}
			if m.confirmReturnFocus != "" {
				cmds = append(cmds, m.setFocus(m.confirmReturnFocus))
				m.confirmReturnFocus = ""
			} else {
				m.scrollToFocused()
			}
			return *m, tea.Batch(cmds...)
		}
	}
	return *m, listenStatus(m.connections.statusChan)
}

// updateTopics manages the topics list UI.
func (m model) updateTopics(msg tea.Msg) (model, tea.Cmd) {
	var cmd, fcmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			return m, tea.Quit
		case "esc":
			cmd := m.setMode(modeClient)
			return m, cmd
		case "left":
			if m.topics.panes.active == 1 {
				fcmd = m.setFocus(idTopicsEnabled)
			}
		case "right":
			if m.topics.panes.active == 0 {
				fcmd = m.setFocus(idTopicsDisabled)
			}
		case "delete":
			i := m.topics.selected
			if i >= 0 && i < len(m.topics.items) {
				name := m.topics.items[i].title
				m.confirmReturnFocus = m.ui.focusOrder[m.ui.focusIndex]
				m.startConfirm(fmt.Sprintf("Delete topic '%s'? [y/n]", name), "", func() {
					m.removeTopic(i)
					m.rebuildActiveTopicList()
				})
				return m, listenStatus(m.connections.statusChan)
			}
		case "enter", " ":
			i := m.topics.selected
			if i >= 0 && i < len(m.topics.items) {
				m.toggleTopic(i)
				m.rebuildActiveTopicList()
			}
		}
	}
	m.topics.list, cmd = m.topics.list.Update(msg)
	if m.topics.panes.active == 0 {
		m.topics.panes.subscribed.sel = m.topics.list.Index()
		m.topics.panes.subscribed.page = m.topics.list.Paginator.Page
	} else {
		m.topics.panes.unsubscribed.sel = m.topics.list.Index()
		m.topics.panes.unsubscribed.page = m.topics.list.Paginator.Page
	}
	m.topics.selected = m.indexForPane(m.topics.panes.active, m.topics.list.Index())
	return m, tea.Batch(fcmd, cmd, listenStatus(m.connections.statusChan))
}

// updatePayloads manages the stored payloads list.
func (m model) updatePayloads(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			return m, tea.Quit
		case "esc":
			cmd := m.setMode(modeClient)
			return m, cmd
		case "delete":
			i := m.message.list.Index()
			if i >= 0 {
				items := m.message.list.Items()
				if i < len(items) {
					m.message.payloads = append(m.message.payloads[:i], m.message.payloads[i+1:]...)
					items = append(items[:i], items[i+1:]...)
					m.message.list.SetItems(items)
				}
			}
			return m, listenStatus(m.connections.statusChan)
		case "enter":
			i := m.message.list.Index()
			if i >= 0 {
				items := m.message.list.Items()
				if i < len(items) {
					pi := items[i].(payloadItem)
					m.topics.input.SetValue(pi.topic)
					m.message.input.SetValue(pi.payload)
					cmd := m.setMode(modeClient)
					return m, cmd
				}
			}
		}
	}
	m.message.list, cmd = m.message.list.Update(msg)
	return m, tea.Batch(cmd, listenStatus(m.connections.statusChan))
}

// updateSelectionRange selects history entries from the anchor to idx.
func (m *model) updateSelectionRange(idx int) {
	start := m.history.selectionAnchor
	end := idx
	if start > end {
		start, end = end, start
	}
	for i := range m.history.items {
		m.history.items[i].isSelected = nil
	}
	for i := start; i <= end && i < len(m.history.items); i++ {
		v := true
		m.history.items[i].isSelected = &v
	}
}

// Update routes messages based on the current mode.
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ui.width = msg.Width
		m.ui.height = msg.Height
		m.connections.manager.ConnectionsList.SetSize(msg.Width-4, msg.Height-6)
		// textinput.View() renders the prompt and cursor in addition
		// to the configured width. Reduce the width slightly so the
		// surrounding box stays within the terminal boundaries.
		m.topics.input.Width = msg.Width - 7
		m.message.input.SetWidth(msg.Width - 4)
		m.message.input.SetHeight(m.layout.message.height)
		if m.layout.history.height == 0 {
			m.layout.history.height = (msg.Height-1)/3 + 10
		}
		m.history.list.SetSize(msg.Width-4, m.layout.history.height)
		if m.layout.trace.height == 0 {
			m.layout.trace.height = msg.Height - 6
		}
		m.traces.view.SetSize(msg.Width-4, m.layout.trace.height)
		m.traces.list.SetSize(msg.Width-4, msg.Height-4)
		m.help.vp.Width = msg.Width - 4
		m.help.vp.Height = msg.Height - 4
		m.ui.viewport.Width = msg.Width
		// Reserve two lines for the info header at the top of the view.
		m.ui.viewport.Height = msg.Height - 2
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+up", "ctrl+k":
			m.ui.viewport.ScrollUp(1)
			return m, nil
		case "ctrl+down", "ctrl+j":
			m.ui.viewport.ScrollDown(1)
			return m, nil
		case "tab":
			if m.currentMode() == modeHistoryFilter {
				nm, cmd := m.updateHistoryFilter(msg)
				*m = nm
				return m, cmd
			}
			if len(m.ui.focusOrder) > 0 {
				m.focus.Next()
				m.ui.focusIndex = m.focus.Index()
				id := m.ui.focusOrder[m.ui.focusIndex]
				m.setFocus(id)
				if id == idTopics {
					if len(m.topics.items) > 0 {
						m.topics.selected = 0
						m.ensureTopicVisible()
					} else {
						m.topics.selected = -1
					}
				}
				return m, nil
			}
		case "shift+tab":
			if m.currentMode() == modeHistoryFilter {
				nm, cmd := m.updateHistoryFilter(msg)
				*m = nm
				return m, cmd
			}
			if len(m.ui.focusOrder) > 0 {
				m.focus.Prev()
				m.ui.focusIndex = m.focus.Index()
				id := m.ui.focusOrder[m.ui.focusIndex]
				m.setFocus(id)
				if id == idTopics {
					if len(m.topics.items) > 0 {
						m.topics.selected = 0
						m.ensureTopicVisible()
					} else {
						m.topics.selected = -1
					}
				}
				return m, nil
			}
		}
		if m.currentMode() != modeHistoryFilter &&
			(msg.String() == "enter" || msg.String() == " " || msg.String() == "space") &&
			m.help.Focused() {
			cmd := m.setMode(modeHelp)
			return m, cmd
		}
	}

	switch m.currentMode() {
	case modeClient:
		cmd := m.updateClient(msg)
		return m, cmd
	case modeConnections:
		nm, cmd := m.updateConnections(msg)
		*m = nm
		return m, cmd
	case modeEditConnection:
		nm, cmd := m.updateForm(msg)
		*m = nm
		return m, cmd
	case modeConfirmDelete:
		nm, cmd := m.updateConfirmDelete(msg)
		*m = nm
		return m, cmd
	case modeTopics:
		nm, cmd := m.updateTopics(msg)
		*m = nm
		return m, cmd
	case modePayloads:
		nm, cmd := m.updatePayloads(msg)
		*m = nm
		return m, cmd
	case modeTracer:
		nm, cmd := m.updateTraces(msg)
		*m = nm
		return m, cmd
	case modeEditTrace:
		nm, cmd := m.updateTraceForm(msg)
		*m = nm
		return m, cmd
	case modeViewTrace:
		nm, cmd := m.updateTraceView(msg)
		*m = nm
		return m, cmd
	case modeImporter:
		nm, cmd := m.updateImporter(msg)
		*m = nm
		return m, cmd
	case modeHistoryFilter:
		nm, cmd := m.updateHistoryFilter(msg)
		*m = nm
		return m, cmd
	case modeHelp:
		nm, cmd := m.updateHelp(msg)
		*m = nm
		return m, cmd
	default:
		return m, nil
	}
}
