package emqutiti

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/emqutiti/connections"
	"github.com/marang/emqutiti/history"
)

// connectionsModel wraps model to satisfy the connections.API interface.
type connectionsModel struct{ *model }

func (m *model) connectionsAPI() connections.API { return &connectionsModel{m} }

// Manager returns the underlying Connections manager.
func (c *connectionsModel) Manager() *connections.Connections { return &c.connections.Manager }

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
	for _, t := range c.topics.Items {
		if t.Subscribed {
			c.mqttClient.Subscribe(t.Name, 0, nil)
		}
	}
}
func (c *connectionsModel) ConnectionMessage() string       { return c.connections.Connection }
func (c *connectionsModel) SetConnectionMessage(msg string) { c.connections.Connection = msg }
func (c *connectionsModel) Active() string                  { return c.connections.Active }
func (c *connectionsModel) BeginAdd() {
	f := connections.NewForm(connections.Profile{}, -1)
	c.connections.Form = &f
}
func (c *connectionsModel) BeginEdit(index int) {
	if index >= 0 && index < len(c.connections.Manager.Profiles) {
		f := connections.NewForm(c.connections.Manager.Profiles[index], index)
		c.connections.Form = &f
	}
}
func (c *connectionsModel) BeginDelete(index int) {
	if index < 0 || index >= len(c.connections.Manager.Profiles) {
		return
	}
	name := c.connections.Manager.Profiles[index].Name
	info := "This also deletes history and traces"
	rf := func() tea.Cmd { return c.setFocus(c.ui.focusOrder[c.ui.focusIndex]) }
	c.startConfirm(
		fmt.Sprintf("Delete broker '%s'? [y/n]", name),
		info,
		rf,
		func() tea.Cmd {
			c.connections.Manager.DeleteConnection(index)
			c.RefreshConnectionItems()
			return nil
		},
		nil,
	)
}
func (c *connectionsModel) Connect(p connections.Profile) tea.Cmd {
	c.connections.FlushStatus()
	if p.FromEnv {
		connections.ApplyEnvVars(&p)
	} else if env := os.Getenv("EMQUTITI_DEFAULT_PASSWORD"); env != "" {
		p.Password = env
	}
	c.connections.SetConnecting(p.Name)
	brokerURL := fmt.Sprintf("%s://%s:%d", p.Schema, p.Host, p.Port)
	c.connections.Connection = "Connecting to " + brokerURL
	c.RefreshConnectionItems()
	return connectBroker(p, c.connections.SendStatus)
}
func (c *connectionsModel) HandleConnectResult(msg connections.ConnectResult) {
	profile := msg.Profile()
	brokerURL := fmt.Sprintf("%s://%s:%d", profile.Schema, profile.Host, profile.Port)
	if err := msg.Err(); err != nil {
		c.connections.SetDisconnected(profile.Name, fmt.Sprintf("Failed to connect to %s: %v", brokerURL, err))
		c.connections.Connection = fmt.Sprintf("Failed to connect to %s: %v", brokerURL, err)
		c.RefreshConnectionItems()
		return
	}
	c.mqttClient = msg.Client().(*MQTTClient)
	c.connections.Active = profile.Name
	if st := c.history.Store(); st != nil {
		st.Close()
	}
	if idx, err := history.OpenStore(profile.Name); err == nil {
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
	ts, ps := c.connections.RestoreState(profile.Name)
	c.topics.SetSnapshot(ts)
	c.payloads.SetSnapshot(ps)
	c.topics.SortTopics()
	c.topics.RebuildActiveTopicList()
	c.SubscribeActiveTopics()
	c.connections.Connection = "Connected to " + brokerURL
	c.connections.SetConnected(profile.Name)
	c.RefreshConnectionItems()
}
func (c *connectionsModel) DisconnectActive() {
	if c.mqttClient != nil {
		c.mqttClient.Disconnect()
		c.connections.SetDisconnected(c.connections.Active, "")
		c.RefreshConnectionItems()
		c.connections.Connection = ""
		c.connections.Active = ""
		c.mqttClient = nil
	}
}
func (c *connectionsModel) ResizeTraces(width, height int) { c.traces.list.SetSize(width, height) }
func (c *connectionsModel) ResetElemPos()                  { c.ui.elemPos = map[string]int{} }
func (c *connectionsModel) SetElemPos(id string, pos int)  { c.ui.elemPos[id] = pos }
func (c *connectionsModel) OverlayHelp(view string) string { return c.overlayHelp(view) }

var _ connections.API = (*connectionsModel)(nil)
