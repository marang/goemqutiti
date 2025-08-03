package emqutiti

import (
	"fmt"

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

// RefreshConnectionItems rebuilds the connections list to show status
// information.
func (c *connectionsState) RefreshConnectionItems() {
	items := []list.Item{}
	for _, p := range c.manager.Profiles {
		status := c.manager.Statuses[p.Name]
		detail := c.manager.Errors[p.Name]
		items = append(items, connectionItem{title: p.Name, status: status, detail: detail})
	}
	c.manager.ConnectionsList.SetItems(items)
}

// connectionsComponent implements the Component interface for managing brokers.
type connectionsComponent struct {
	nav navigator
	api ConnectionsAPI
}

func newConnectionsComponent(nav navigator, api ConnectionsAPI) *connectionsComponent {
	return &connectionsComponent{nav: nav, api: api}
}

func (c *connectionsComponent) Init() tea.Cmd { return nil }

// Update processes input when the connections view is active.
func (c *connectionsComponent) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case connectResult:
		c.api.HandleConnectResult(msg)
		if msg.err == nil {
			cmd = c.nav.SetMode(modeClient)
			return tea.Batch(cmd, c.api.ListenStatus())
		}
		return c.api.ListenStatus()
	case tea.KeyMsg:
		mgr := c.api.Manager()
		if mgr.ConnectionsList.FilterState() == list.Filtering {
			switch msg.String() {
			case "enter":
				i := mgr.ConnectionsList.Index()
				if i >= 0 && i < len(mgr.Profiles) {
					p := mgr.Profiles[i]
					if p.Name == c.api.Active() && mgr.Statuses[p.Name] == "connected" {
						brokerURL := fmt.Sprintf("%s://%s:%d", p.Schema, p.Host, p.Port)
						c.api.SetConnectionMessage("Connected to " + brokerURL)
						c.api.SetConnected(p.Name)
						c.api.RefreshConnectionItems()
						return c.nav.SetMode(modeClient)
					}
					return c.api.Connect(p)
				}
			}
			break
		}
		switch msg.String() {
		case "ctrl+d":
			return tea.Quit
		case "ctrl+r":
			c.api.ResizeTraces(c.nav.Width()-4, c.nav.Height()-4)
			return c.nav.SetMode(modeTracer)
		case "a":
			c.api.BeginAdd()
			return c.nav.SetMode(modeEditConnection)
		case "e":
			i := mgr.ConnectionsList.Index()
			if i >= 0 && i < len(mgr.Profiles) {
				c.api.BeginEdit(i)
				return c.nav.SetMode(modeEditConnection)
			}
		case "enter":
			i := mgr.ConnectionsList.Index()
			if i >= 0 && i < len(mgr.Profiles) {
				p := mgr.Profiles[i]
				if p.Name == c.api.Active() && mgr.Statuses[p.Name] == "connected" {
					brokerURL := fmt.Sprintf("%s://%s:%d", p.Schema, p.Host, p.Port)
					c.api.SetConnectionMessage("Connected to " + brokerURL)
					c.api.SetConnected(p.Name)
					c.api.RefreshConnectionItems()
					return c.nav.SetMode(modeClient)
				}
				return c.api.Connect(p)
			}
		case "delete":
			i := mgr.ConnectionsList.Index()
			if i >= 0 {
				c.api.BeginDelete(i)
				return c.api.ListenStatus()
			}
		case "x":
			c.api.DisconnectActive()
		}
	}
	mgr := c.api.Manager()
	mgr.ConnectionsList, cmd = mgr.ConnectionsList.Update(msg)
	return tea.Batch(cmd, c.api.ListenStatus())
}

// View renders the connections UI component.
func (c *connectionsComponent) View() string {
	c.api.ResetElemPos()
	c.api.SetElemPos(idConnList, 1)
	listView := c.api.Manager().ConnectionsList.View()
	help := ui.InfoStyle.Render("[enter] connect/open client  [x] disconnect  [a]dd [e]dit [del] delete  Ctrl+R traces")
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	view := ui.LegendBox(content, "Brokers", c.nav.Width()-2, 0, ui.ColBlue, true, -1)
	return c.api.OverlayHelp(view)
}

// Focus marks the component as focused.
func (c *connectionsComponent) Focus() tea.Cmd {
	c.api.Manager().Focused = true
	return nil
}

// Blur marks the component as blurred.
func (c *connectionsComponent) Blur() { c.api.Manager().Focused = false }

// Focusables exposes focusable elements for the connections component.
func (c *connectionsComponent) Focusables() map[string]Focusable {
	return map[string]Focusable{idConnList: &nullFocusable{}}
}
