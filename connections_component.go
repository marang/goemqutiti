package emqutiti

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/emqutiti/ui"
)

// connectionItem represents a single broker profile in the list.
type connectionItem struct {
	title  string
	status string
	detail string
}

func (c connectionItem) FilterValue() string { return c.title }
func (c connectionItem) Title() string       { return c.title }
func (c connectionItem) Description() string {
	if c.detail != "" {
		return c.status + "\n" + c.detail
	}
	return c.status
}

// connectionData stores topics and payloads for a connection.
type connectionData struct {
	Topics   []topicItem
	Payloads []payloadItem
}

// connectionsState holds connection related state shared across components.
type connectionsState struct {
	connection  string
	active      string
	manager     Connections
	form        *connectionForm
	deleteIndex int
	statusChan  chan string
	saved       map[string]connectionData
}

// ConnectionStatusManager exposes methods to update connection status information.
type ConnectionStatusManager interface {
	SetConnecting(name string)
	SetConnected(name string)
	SetDisconnected(name, detail string)
}

var _ ConnectionStatusManager = (*connectionsState)(nil)

// SetConnecting marks the named connection as in progress.
func (c *connectionsState) SetConnecting(name string) {
	c.manager.Statuses[name] = "connecting"
	c.manager.Errors[name] = ""
}

// SetConnected marks the named connection as connected.
func (c *connectionsState) SetConnected(name string) {
	c.manager.Statuses[name] = "connected"
	c.manager.Errors[name] = ""
}

// SetDisconnected marks the named connection as disconnected with an optional detail.
func (c *connectionsState) SetDisconnected(name, detail string) {
	c.manager.Statuses[name] = "disconnected"
	c.manager.Errors[name] = detail
}

// ListenStatus returns a command to receive connection status updates.
func (c *connectionsState) ListenStatus() tea.Cmd {
	return listenStatus(c.statusChan)
}

// SendStatus reports a status message if the channel is available.
func (c *connectionsState) SendStatus(msg string) {
	if c.statusChan != nil {
		c.statusChan <- msg
	}
}

// FlushStatus discards any pending status messages.
func (c *connectionsState) FlushStatus() { flushStatus(c.statusChan) }

// connectionsComponent implements the Component interface for managing brokers.
type connectionsComponent struct{ m *model }

func newConnectionsComponent(m *model) *connectionsComponent { return &connectionsComponent{m: m} }

func (c *connectionsComponent) Init() tea.Cmd { return nil }

// Update processes input when the connections view is active.
func (c *connectionsComponent) Update(msg tea.Msg) tea.Cmd {
	m := c.m
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case connectResult:
		brokerURL := fmt.Sprintf("%s://%s:%d", msg.profile.Schema, msg.profile.Host, msg.profile.Port)
		if msg.err != nil {
			m.connections.SetDisconnected(msg.profile.Name, fmt.Sprintf("Failed to connect to %s: %v", brokerURL, msg.err))
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
			m.connections.SetConnected(msg.profile.Name)
			m.refreshConnectionItems()
			cmd := m.setMode(modeClient)
			return tea.Batch(cmd, m.connections.ListenStatus())
		}
		return m.connections.ListenStatus()
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
						m.connections.SetConnected(p.Name)
						m.refreshConnectionItems()
						cmd := m.setMode(modeClient)
						return cmd
					}
					m.connections.FlushStatus()
					if p.FromEnv {
						ApplyEnvVars(&p)
					} else if env := os.Getenv("EMQUTITI_DEFAULT_PASSWORD"); env != "" {
						p.Password = env
					}
					m.connections.SetConnecting(p.Name)
					brokerURL := fmt.Sprintf("%s://%s:%d", p.Schema, p.Host, p.Port)
					m.connections.connection = "Connecting to " + brokerURL
					m.refreshConnectionItems()
					return connectBroker(p, m.connections.SendStatus)
				}
			}
			break
		}
		switch msg.String() {
		case "ctrl+d":
			return tea.Quit
		case "ctrl+r":
			m.traces.list.SetSize(m.ui.width-4, m.ui.height-4)
			cmd := m.setMode(modeTracer)
			return cmd
		case "a":
			f := newConnectionForm(Profile{}, -1)
			m.connections.form = &f
			cmd := m.setMode(modeEditConnection)
			return cmd
		case "e":
			i := m.connections.manager.ConnectionsList.Index()
			if i >= 0 && i < len(m.connections.manager.Profiles) {
				f := newConnectionForm(m.connections.manager.Profiles[i], i)
				m.connections.form = &f
				cmd := m.setMode(modeEditConnection)
				return cmd
			}
		case "enter":
			i := m.connections.manager.ConnectionsList.Index()
			if i >= 0 && i < len(m.connections.manager.Profiles) {
				p := m.connections.manager.Profiles[i]
				if p.Name == m.connections.active && m.connections.manager.Statuses[p.Name] == "connected" {
					brokerURL := fmt.Sprintf("%s://%s:%d", p.Schema, p.Host, p.Port)
					m.connections.connection = "Connected to " + brokerURL
					m.connections.SetConnected(p.Name)
					m.refreshConnectionItems()
					cmd := m.setMode(modeClient)
					return cmd
				}
				m.connections.FlushStatus()
				if p.FromEnv {
					ApplyEnvVars(&p)
				} else if env := os.Getenv("EMQUTITI_DEFAULT_PASSWORD"); env != "" {
					p.Password = env
				}
				m.connections.SetConnecting(p.Name)
				brokerURL := fmt.Sprintf("%s://%s:%d", p.Schema, p.Host, p.Port)
				m.connections.connection = "Connecting to " + brokerURL
				m.refreshConnectionItems()
				return connectBroker(p, m.connections.SendStatus)
			}
		case "delete":
			i := m.connections.manager.ConnectionsList.Index()
			if i >= 0 {
				name := m.connections.manager.Profiles[i].Name
				info := "This also deletes history and traces"
				m.confirm.returnFocus = m.ui.focusOrder[m.ui.focusIndex]
				m.startConfirm(fmt.Sprintf("Delete broker '%s'? [y/n]", name), info, func() tea.Cmd {
					m.connections.manager.DeleteConnection(i)
					m.connections.manager.refreshList()
					m.refreshConnectionItems()
					return nil
				})
				return m.connections.ListenStatus()
			}
		case "x":
			if m.mqttClient != nil {
				m.mqttClient.Disconnect()
				m.connections.SetDisconnected(m.connections.active, "")
				m.refreshConnectionItems()
				m.connections.connection = ""
				m.connections.active = ""
				m.mqttClient = nil
			}
		}
	}
	m.connections.manager.ConnectionsList, cmd = m.connections.manager.ConnectionsList.Update(msg)
	return tea.Batch(cmd, m.connections.ListenStatus())
}

// View renders the connections UI component.
func (c *connectionsComponent) View() string {
	m := c.m
	m.ui.elemPos = map[string]int{}
	m.ui.elemPos[idConnList] = 1
	listView := m.connections.manager.ConnectionsList.View()
	help := ui.InfoStyle.Render("[enter] connect/open client  [x] disconnect  [a]dd [e]dit [del] delete  Ctrl+R traces")
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	view := ui.LegendBox(content, "Brokers", m.ui.width-2, 0, ui.ColBlue, true, -1)
	return m.overlayHelp(view)
}

// Focus marks the component as focused.
func (c *connectionsComponent) Focus() tea.Cmd {
	c.m.connections.manager.Focused = true
	return nil
}

// Blur marks the component as blurred.
func (c *connectionsComponent) Blur() { c.m.connections.manager.Focused = false }

// Focusables exposes focusable elements for the connections component.
func (c *connectionsComponent) Focusables() map[string]Focusable {
	return map[string]Focusable{idConnList: &nullFocusable{}}
}
