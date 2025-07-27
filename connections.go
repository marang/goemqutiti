package main

import (
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
	Name                string `toml:"name"`
	Schema              string `toml:"schema"`
	Host                string `toml:"host"`
	Port                int    `toml:"port"`
	ClientID            string `toml:"client_id"`
	Username            string `toml:"username"`
	Password            string `toml:"password"`
	SSL                 bool   `toml:"ssl_tls"`
	MQTTVersion         string `toml:"mqtt_version"`
	ConnectTimeout      int    `toml:"connect_timeout"`
	KeepAlive           int    `toml:"keep_alive"`
	QoS                 int    `toml:"qos"`
	AutoReconnect       bool   `toml:"auto_reconnect"`
	ReconnectPeriod     int    `toml:"reconnect_period"`
	CleanStart          bool   `toml:"clean_start"`
	SessionExpiry       int    `toml:"session_expiry_interval"`
	ReceiveMaximum      int    `toml:"receive_maximum"`
	MaximumPacketSize   int    `toml:"maximum_packet_size"`
	TopicAliasMaximum   int    `toml:"topic_alias_maximum"`
	RequestResponseInfo bool   `toml:"request_response_info"`
	RequestProblemInfo  bool   `toml:"request_problem_info"`
	LastWillEnabled     bool   `toml:"last_will_enabled"`
	LastWillTopic       string `toml:"last_will_topic"`
	LastWillQos         int    `toml:"last_will_qos"`
	LastWillRetain      bool   `toml:"last_will_retain"`
	LastWillPayload     string `toml:"last_will_payload"`
	RandomIDSuffix      bool   `toml:"random_id_suffix"`
}

// Connections manages the state and logic for handling broker profiles.
type Connections struct {
	ConnectionsList    list.Model
	TextInput          textinput.Model
	DefaultProfileName string            `toml:"default_profile"`
	Profiles           []Profile         `toml:"profiles"`
	Statuses           map[string]string // connection status by name
	Focused            bool              // Indicates if the broker manager is focused
}

// NewConnectionsModel initializes a new ConnectionsModel with default values.
func NewConnectionsModel() Connections {
	connectionList := list.New([]list.Item{}, connectionDelegate{}, 0, 0)
	// Ensure items are visible by setting a reasonable default size
	connectionList.SetSize(30, 10)
	connectionList.DisableQuitKeybindings()

	return Connections{
		ConnectionsList: connectionList,
		TextInput:       textinput.New(),
		Profiles:        []Profile{},
		Statuses:        make(map[string]string),
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
func (m *Connections) AddConnection(p Profile) {
	plain := p.Password
	keyref := "keyring:emqutiti-" + p.Name + "/" + p.Username
	p.Password = keyref
	m.Profiles = append(m.Profiles, p)
	if m.Statuses == nil {
		m.Statuses = make(map[string]string)
	}
	m.Statuses[p.Name] = "disconnected"
	m.saveConfigToFile()
	m.savePasswordToKeyring(p.Name, p.Username, plain)
}

// EditConnection updates an existing connection and saves changes to config.toml and keyring.
func (m *Connections) EditConnection(index int, p Profile) {
	if index >= 0 && index < len(m.Profiles) {
		plain := p.Password
		oldName := m.Profiles[index].Name
		p.Password = "keyring:emqutiti-" + p.Name + "/" + p.Username
		m.Profiles[index] = p
		if oldName != p.Name {
			if status, ok := m.Statuses[oldName]; ok {
				delete(m.Statuses, oldName)
				m.Statuses[p.Name] = status
			}
		}
		m.saveConfigToFile()
		m.savePasswordToKeyring(p.Name, p.Username, plain)
	}
}

// DeleteConnection removes a connection from the list and updates config.toml.
func (m *Connections) DeleteConnection(index int) {
	if index >= 0 && index < len(m.Profiles) {
		name := m.Profiles[index].Name
		m.Profiles = append(m.Profiles[:index], m.Profiles[index+1:]...)
		delete(m.Statuses, name)
		m.saveConfigToFile()
	}
}

// saveConfigToFile writes the current connections to the config.toml file using BurntSushi/toml.
func (m *Connections) saveConfigToFile() {
	saved := loadState()
	cfg := userConfig{
		DefaultProfileName: m.DefaultProfileName,
		Profiles:           m.Profiles,
		Saved:              make(map[string]persistedConn),
	}
	for k, v := range saved {
		var topics []persistedTopic
		for _, t := range v.Topics {
			topics = append(topics, persistedTopic{Title: t.title, Active: t.active})
		}
		var payloads []persistedPayload
		for _, p := range v.Payloads {
			payloads = append(payloads, persistedPayload{Topic: p.topic, Payload: p.payload})
		}
		cfg.Saved[k] = persistedConn{Topics: topics, Payloads: payloads}
	}
	writeConfig(cfg)
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
				// Do not fail if the keyring entry is missing
				fmt.Println("Warning:", err)
				connections.Profiles[i].Password = ""
				continue
			}
			// Update the password in the profile
			connections.Profiles[i].Password = keyringPassword
		}
	}

	return &connections, nil
}

// LoadProfiles updates c with profiles from the config file. It logs errors but
// leaves c unchanged on failure.
func (c *Connections) LoadProfiles(filePath string) {
	loaded, err := LoadFromConfig(filePath)
	if err != nil {
		fmt.Println("Warning:", err)
		return
	}
	c.DefaultProfileName = loaded.DefaultProfileName
	c.Profiles = loaded.Profiles
	statuses := make(map[string]string)
	for _, p := range c.Profiles {
		if st, ok := c.Statuses[p.Name]; ok {
			statuses[p.Name] = st
		} else {
			statuses[p.Name] = "disconnected"
		}
	}
	c.Statuses = statuses
}
