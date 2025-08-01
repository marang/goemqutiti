package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Connections manages the state and logic for handling broker profiles.
type Connections struct {
	ConnectionsList    list.Model
	TextInput          textinput.Model
	DefaultProfileName string            `toml:"default_profile"`
	Profiles           []Profile         `toml:"profiles"`
	Statuses           map[string]string // connection status by name
	Errors             map[string]string // last connection error message
	Focused            bool              // Indicates if the broker manager is focused
}

// NewConnectionsModel initializes a new ConnectionsModel with default values.
func NewConnectionsModel() Connections {
	connectionList := list.New([]list.Item{}, connectionDelegate{}, 0, 0)
	// Ensure items are visible by setting a reasonable default size
	connectionList.SetSize(30, 10)
	connectionList.DisableQuitKeybindings()
	connectionList.SetShowTitle(false)

	return Connections{
		ConnectionsList: connectionList,
		TextInput:       textinput.New(),
		Profiles:        []Profile{},
		Statuses:        make(map[string]string),
		Errors:          make(map[string]string),
		Focused:         false,
	}
}

// Init initializes the Bubble Tea model.
func (m Connections) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages and updates the model accordingly.
func (m Connections) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.Focused && msg.String() == "a" {
			m.TextInput.Focus()
		}
	}
	m.ConnectionsList, cmd = m.ConnectionsList.Update(msg)
	return m, cmd
}

// View renders the UI components.
func (m Connections) View() string {
	return m.ConnectionsList.View()
}

// AddConnection adds a new connection to the list and saves it to config.toml and keyring.
func (m *Connections) AddConnection(p Profile) {
	if m.Statuses == nil {
		m.Statuses = make(map[string]string)
	}
	m.Statuses[p.Name] = "disconnected"
	if m.Errors == nil {
		m.Errors = make(map[string]string)
	}
	m.Errors[p.Name] = ""
	persistProfileChange(&m.Profiles, m.DefaultProfileName, p, -1)
	m.refreshList()
}

// EditConnection updates an existing connection and saves changes to config.toml and keyring.
func (m *Connections) EditConnection(index int, p Profile) {
	if index >= 0 && index < len(m.Profiles) {
		oldName := m.Profiles[index].Name
		if oldName != p.Name {
			if status, ok := m.Statuses[oldName]; ok {
				delete(m.Statuses, oldName)
				m.Statuses[p.Name] = status
			}
			if errMsg, ok := m.Errors[oldName]; ok {
				delete(m.Errors, oldName)
				m.Errors[p.Name] = errMsg
			}
		}
		persistProfileChange(&m.Profiles, m.DefaultProfileName, p, index)
		m.refreshList()
	}
}

// DeleteConnection removes a connection from the list and updates config.toml.
func (m *Connections) DeleteConnection(index int) {
	if index >= 0 && index < len(m.Profiles) {
		name := m.Profiles[index].Name
		m.Profiles = append(m.Profiles[:index], m.Profiles[index+1:]...)
		delete(m.Statuses, name)
		delete(m.Errors, name)
		// Persist removal so the connection no longer appears after a restart
		saveConfig(m.Profiles, m.DefaultProfileName)
		deleteProfileData(name)
		m.refreshList()
	}
}

// refreshList rebuilds the list items from the current profiles.
func (m *Connections) refreshList() {
	items := []list.Item{}
	for _, p := range m.Profiles {
		status := m.Statuses[p.Name]
		detail := m.Errors[p.Name]
		items = append(items, connectionItem{title: p.Name, status: status, detail: detail})
	}
	m.ConnectionsList.SetItems(items)
}

// LoadFromConfig loads connection profiles from the config file.
func LoadFromConfig(filePath string) (*Connections, error) {
	cfg, err := LoadConfig(filePath)
	if err != nil {
		return nil, err
	}
	return &Connections{DefaultProfileName: cfg.DefaultProfile, Profiles: cfg.Profiles}, nil
}

// LoadProfiles updates c with profiles from the config file. It logs errors but
// leaves c unchanged on failure.
func (c *Connections) LoadProfiles(filePath string) error {
	loaded, err := LoadFromConfig(filePath)
	if err != nil {
		fmt.Println("Warning:", err)
		return err
	}
	c.DefaultProfileName = loaded.DefaultProfileName
	c.Profiles = loaded.Profiles
	statuses := make(map[string]string)
	errors := make(map[string]string)
	for _, p := range c.Profiles {
		if st, ok := c.Statuses[p.Name]; ok {
			statuses[p.Name] = st
		} else {
			statuses[p.Name] = "disconnected"
		}
		if errMsg, ok := c.Errors[p.Name]; ok {
			errors[p.Name] = errMsg
		} else {
			errors[p.Name] = ""
		}
	}
	c.Statuses = statuses
	c.Errors = errors
	return nil
}
