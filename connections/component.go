package connections

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/emqutiti/ui"
)

const (
	modeClient         = 0
	modeConnections    = 1
	modeEditConnection = 2
	modeTracer         = 6
	idConnList         = "conn-list"
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

// State holds connection related state shared across components.
type State struct {
	Connection  string
	Active      string
	Manager     Connections
	Form        *Form
	DeleteIndex int
	StatusChan  chan string
	Saved       map[string]ConnectionSnapshot
}

// SetConnecting marks the named connection as in progress.
func (c *State) SetConnecting(name string) {
	c.Manager.Statuses[name] = "connecting"
	c.Manager.Errors[name] = ""
}

// SetConnected marks the named connection as connected.
func (c *State) SetConnected(name string) {
	c.Manager.Statuses[name] = "connected"
	c.Manager.Errors[name] = ""
}

// SetDisconnected marks the named connection as disconnected with an optional detail.
func (c *State) SetDisconnected(name, detail string) {
	c.Manager.Statuses[name] = "disconnected"
	c.Manager.Errors[name] = detail
}

// ListenStatus returns a command to receive connection status updates.
func (c *State) ListenStatus() tea.Cmd {
	return ListenStatus(c.StatusChan)
}

// SendStatus reports a status message if the channel is available.
func (c *State) SendStatus(msg string) {
	if c.StatusChan != nil {
		c.StatusChan <- msg
	}
}

// FlushStatus discards any pending status messages.
func (c *State) FlushStatus() { FlushStatus(c.StatusChan) }

// RefreshConnectionItems rebuilds the connections list to show status
// information.
func (c *State) RefreshConnectionItems() {
	c.Manager.refreshList()
}

// SaveCurrent persists topics and payloads for the active connection.
func (c *State) SaveCurrent(topics []TopicSnapshot, payloads []PayloadSnapshot) {
	if c.Active == "" {
		return
	}
	c.Saved[c.Active] = ConnectionSnapshot{Topics: topics, Payloads: payloads}
	if err := SaveState(c.Saved); err != nil {
		log.Printf("Failed to save connection state: %v", err)
	}
}

// RestoreState returns saved topics and payloads for the named connection.
func (c *State) RestoreState(name string) ([]TopicSnapshot, []PayloadSnapshot) {
	if data, ok := c.Saved[name]; ok {
		return data.Topics, data.Payloads
	}
	return []TopicSnapshot{}, []PayloadSnapshot{}
}

// Component implements the Component interface for managing brokers.
type Component struct {
	nav Navigator
	api API
}

func NewComponent(nav Navigator, api API) *Component {
	return &Component{nav: nav, api: api}
}

func (c *Component) Init() tea.Cmd { return nil }

// Update processes input when the connections view is active.
func (c *Component) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case ConnectResult:
		c.api.HandleConnectResult(msg)
		if msg.Err == nil {
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
		case "ctrl+o":
			i := mgr.ConnectionsList.Index()
			if i >= 0 {
				if mgr.DefaultProfileName == mgr.Profiles[i].Name {
					mgr.ClearDefault()
				} else {
					mgr.SetDefault(i)
				}
			}
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
func (c *Component) View() string {
	c.api.ResetElemPos()
	c.api.SetElemPos(idConnList, 1)
	listView := c.api.Manager().ConnectionsList.View()
	help := ui.InfoStyle.Render("[enter] connect/open client  [x] disconnect  [a]dd [e]dit [del] delete  Ctrl+O default  Ctrl+R traces")
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	view := ui.LegendBox(content, "Brokers", c.nav.Width()-2, 0, ui.ColBlue, true, -1)
	return c.api.OverlayHelp(view)
}

// Focus marks the component as focused.
func (c *Component) Focus() tea.Cmd {
	c.api.Manager().Focused = true
	return nil
}

// Blur marks the component as blurred.
func (c *Component) Blur() { c.api.Manager().Focused = false }

// Focusables exposes focusable elements for the connections component.
func (c *Component) Focusables() map[string]Focusable {
	return map[string]Focusable{idConnList: &nullFocusable{}}
}

// nullFocusable is a no-op focusable used for non-interactive areas.
type nullFocusable struct{ focused bool }

func (n *nullFocusable) Focus()          { n.focused = true }
func (n *nullFocusable) Blur()           { n.focused = false }
func (n *nullFocusable) IsFocused() bool { return n.focused }
func (n *nullFocusable) View() string    { return "" }
