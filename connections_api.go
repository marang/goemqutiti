package emqutiti

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/emqutiti/history"
)

// ConnectionsAPI defines the methods used by components to manage connection profiles
// and status updates without depending on the model implementation.
type ConnectionsAPI interface {
	ConnectionStatusManager
	Manager() *Connections
	ListenStatus() tea.Cmd
	SendStatus(string)
	FlushStatus()
	RefreshConnectionItems()
	SubscribeActiveTopics()
	ConnectionMessage() string
	SetConnectionMessage(string)
	Active() string
	BeginAdd()
	BeginEdit(index int)
	BeginDelete(index int)
	Connect(p Profile) tea.Cmd
	HandleConnectResult(msg connectResult)
	DisconnectActive()
	ResizeTraces(width, height int)
	ResetElemPos()
	SetElemPos(id string, pos int)
	OverlayHelp(view string) string
}

// connectionsModel wraps model to satisfy the ConnectionsAPI interface.
type connectionsModel struct{ *model }

func (m *model) connectionsAPI() ConnectionsAPI { return &connectionsModel{m} }

// Manager returns the underlying Connections manager.
func (c *connectionsModel) Manager() *Connections { return &c.connections.manager }

func (c *connectionsModel) SetConnecting(name string) { c.connections.SetConnecting(name) }
func (c *connectionsModel) SetConnected(name string)  { c.connections.SetConnected(name) }
func (c *connectionsModel) SetDisconnected(name, detail string) {
	c.connections.SetDisconnected(name, detail)
}

func (c *connectionsModel) ListenStatus() tea.Cmd { return c.connections.ListenStatus() }
func (c *connectionsModel) SendStatus(msg string) { c.connections.SendStatus(msg) }
func (c *connectionsModel) FlushStatus()          { c.connections.FlushStatus() }

func (c *connectionsModel) RefreshConnectionItems() { c.connections.RefreshConnectionItems() }
func (c *connectionsModel) SubscribeActiveTopics() {
	if c.mqttClient == nil {
		return
	}
	for _, t := range c.topics.items {
		if t.subscribed {
			c.mqttClient.Subscribe(t.title, 0, nil)
		}
	}
}
func (c *connectionsModel) ConnectionMessage() string       { return c.connections.connection }
func (c *connectionsModel) SetConnectionMessage(msg string) { c.connections.connection = msg }
func (c *connectionsModel) Active() string                  { return c.connections.active }
func (c *connectionsModel) BeginAdd() {
	f := newConnectionForm(Profile{}, -1)
	c.connections.form = &f
}
func (c *connectionsModel) BeginEdit(index int) {
	if index >= 0 && index < len(c.connections.manager.Profiles) {
		f := newConnectionForm(c.connections.manager.Profiles[index], index)
		c.connections.form = &f
	}
}
func (c *connectionsModel) BeginDelete(index int) {
	if index < 0 || index >= len(c.connections.manager.Profiles) {
		return
	}
	name := c.connections.manager.Profiles[index].Name
	info := "This also deletes history and traces"
	rf := func() tea.Cmd { return c.setFocus(c.ui.focusOrder[c.ui.focusIndex]) }
	c.startConfirm(
		fmt.Sprintf("Delete broker '%s'? [y/n]", name),
		info,
		rf,
		func() tea.Cmd {
			c.connections.manager.DeleteConnection(index)
			c.connections.manager.refreshList()
			c.RefreshConnectionItems()
			return nil
		},
		nil,
	)
}
func (c *connectionsModel) Connect(p Profile) tea.Cmd {
	c.connections.FlushStatus()
	if p.FromEnv {
		ApplyEnvVars(&p)
	} else if env := os.Getenv("EMQUTITI_DEFAULT_PASSWORD"); env != "" {
		p.Password = env
	}
	c.connections.SetConnecting(p.Name)
	brokerURL := fmt.Sprintf("%s://%s:%d", p.Schema, p.Host, p.Port)
	c.connections.connection = "Connecting to " + brokerURL
	c.RefreshConnectionItems()
	return connectBroker(p, c.connections.SendStatus)
}
func (c *connectionsModel) HandleConnectResult(msg connectResult) {
	brokerURL := fmt.Sprintf("%s://%s:%d", msg.profile.Schema, msg.profile.Host, msg.profile.Port)
	if msg.err != nil {
		c.connections.SetDisconnected(msg.profile.Name, fmt.Sprintf("Failed to connect to %s: %v", brokerURL, msg.err))
		c.connections.connection = fmt.Sprintf("Failed to connect to %s: %v", brokerURL, msg.err)
		c.RefreshConnectionItems()
		return
	}
	c.mqttClient = msg.client
	c.connections.active = msg.profile.Name
	if st := c.history.Store(); st != nil {
		st.Close()
	}
	if idx, err := history.OpenStore(msg.profile.Name); err == nil {
		c.history.SetStore(idx)
		msgs := idx.Search(false, nil, time.Time{}, time.Time{}, "")
		hitems := make([]history.Item, len(msgs))
		items := make([]list.Item, len(msgs))
		for i, mmsg := range msgs {
			hi := history.Item{Timestamp: mmsg.Timestamp, Topic: mmsg.Topic, Payload: mmsg.Payload, Kind: mmsg.Kind, Archived: mmsg.Archived}
			hitems[i] = hi
			items[i] = hi
		}
		c.history.SetItems(hitems)
		c.history.List().SetItems(items)
	}
	ts, ps := c.connections.RestoreState(msg.profile.Name)
	c.topics.SetSnapshot(ts)
	c.payloads.SetSnapshot(ps)
	c.topics.SortTopics()
	c.topics.RebuildActiveTopicList()
	c.SubscribeActiveTopics()
	c.connections.connection = "Connected to " + brokerURL
	c.connections.SetConnected(msg.profile.Name)
	c.RefreshConnectionItems()
}
func (c *connectionsModel) DisconnectActive() {
	if c.mqttClient != nil {
		c.mqttClient.Disconnect()
		c.connections.SetDisconnected(c.connections.active, "")
		c.RefreshConnectionItems()
		c.connections.connection = ""
		c.connections.active = ""
		c.mqttClient = nil
	}
}
func (c *connectionsModel) ResizeTraces(width, height int) { c.traces.list.SetSize(width, height) }
func (c *connectionsModel) ResetElemPos()                  { c.ui.elemPos = map[string]int{} }
func (c *connectionsModel) SetElemPos(id string, pos int)  { c.ui.elemPos[id] = pos }
func (c *connectionsModel) OverlayHelp(view string) string { return c.overlayHelp(view) }

type _ = connectionsModel // avoid unused type if not referenced
