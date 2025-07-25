package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/zalando/go-keyring"
)

type Profile struct {
	Name     string `toml:"name"`
	Broker   string `toml:"broker"`
	ClientID string `toml:"client_id"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

// Connections manages the state and logic for handling connections.
type Connections struct {
	ConnectionsList    list.Model
	TextInput          textinput.Model
	DefaultProfileName string    `toml:"default_profile"`
	Profiles           []Profile `toml:"profiles"`
	Focused            bool      // Indicates if the connection manager is focused
}

// NewConnectionsModel initializes a new ConnectionsModel with default values.
func NewConnectionsModel() Connections {
	connectionList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	connectionList.Title = "Connections"

	return Connections{
		ConnectionsList: connectionList,
		TextInput:       textinput.New(),
		Profiles:        []Profile{},
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
		if m.Focused {
			// Handle key events when focused
			switch msg.String() {
			case "a": // Add new connection
				m.TextInput.Focus()
				fmt.Println("Add new connection")
			case "e": // Edit selected connection
				fmt.Println("Edit selected connection")
			case "d": // Delete selected connection
				fmt.Println("Delete selected connection")
			}
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
func (m *Connections) AddConnection(name, broker, clientID, username, password string) {
	m.Profiles = append(m.Profiles, Profile{Name: name, Broker: broker, ClientID: clientID, Username: username, Password: "keyring:emqutiti-" + name + "/" + username})
	m.saveConfigToFile()
	m.savePasswordToKeyring(name, username, password)
}

// EditConnection updates an existing connection and saves changes to config.toml and keyring.
func (m *Connections) EditConnection(index int, name, broker, clientID, username, password string) {
	if index >= 0 && index < len(m.Profiles) {
		m.Profiles[index] = Profile{Name: name, Broker: broker, ClientID: clientID, Username: username, Password: "keyring:emqutiti-" + name + "/" + username}
		m.saveConfigToFile()
		m.savePasswordToKeyring(name, username, password)
	}
}

// DeleteConnection removes a connection from the list and updates config.toml.
func (m *Connections) DeleteConnection(index int) {
	if index >= 0 && index < len(m.Profiles) {
		m.Profiles = append(m.Profiles[:index], m.Profiles[index+1:]...)
		m.saveConfigToFile()
	}
}

// saveConfigToFile writes the current connections to the config.toml file using BurntSushi/toml.
func (m *Connections) saveConfigToFile() {
	config := map[string]Profile{}
	for _, conn := range m.Profiles {
		config[conn.Name] = Profile{
			Name:     conn.Name,
			Broker:   conn.Broker,
			ClientID: conn.ClientID,
			Username: conn.Username,
			Password: "keyring:emqutiti-" + conn.Name + "/" + conn.Username,
		}
	}

	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(config); err != nil {
		fmt.Println("Error encoding TOML:", err)
		return
	}

	filePath, _ := DefaultUserConfigFile()
	os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		fmt.Println("Error writing config file:", err)
	}
}

// savePasswordToKeyring stores the password in the keyring.
func (m *Connections) savePasswordToKeyring(service, username, password string) {
	err := keyring.Set("emqutiti-"+service, username, password)
	if err != nil {
		fmt.Println("Error saving password to keyring:", err)
	}
}

// DefaultUserConfigFile try to load config from ~/.emqutiti/config.toml
func DefaultUserConfigFile() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".emqutiti", "config.toml"), nil
}

func retrievePasswordFromKeyring(password string) (string, error) {
	if !strings.HasPrefix(password, "keyring:") {
		return "", fmt.Errorf("password does not reference keyring")
	}
	parts := strings.SplitN(password, ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid keyring reference: %s", password)
	}
	serviceUsername := strings.SplitN(parts[1], "/", 2)
	if len(serviceUsername) != 2 {
		return "", fmt.Errorf("invalid keyring format: %s", parts[1])
	}

	// Retrieve the password from the keyring
	keyringPassword, err := keyring.Get(serviceUsername[0], serviceUsername[1])
	if err != nil {
		return "", fmt.Errorf("failed to retrieve password from keyring for %s/%s: %w", serviceUsername[0], serviceUsername[1], err)
	}

	return keyringPassword, nil
}

// LoadConfig loads the configuration from a TOML file and retrieves keyring-stored passwords.
func LoadFromConfig(filePath string) (*Connections, error) {
	var err error

	// Step 1: Determine the config file path if not provided
	if filePath == "" {
		if filePath, err = DefaultUserConfigFile(); err != nil {
			return nil, err
		}
	}

	// Step 2: Decode the TOML file into the Config struct
	var connections Connections
	if _, err := toml.DecodeFile(filePath, &connections); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	// Step 3: Process each profile to handle keyring references
	for i := range connections.Profiles {
		password := connections.Profiles[i].Password

		// Check if the password references the keyring
		if strings.HasPrefix(password, "keyring:") {
			keyringPassword, err := retrievePasswordFromKeyring(password)
			if err != nil {
				return nil, err
			}
			// Update the password in the profile
			connections.Profiles[i].Password = keyringPassword
		}
	}

	return &connections, nil
}
