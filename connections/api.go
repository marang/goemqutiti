package connections

import (
	tea "github.com/charmbracelet/bubbletea"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Client defines the MQTT functions used by the connections package.
type Client interface {
	Subscribe(topic string, qos byte, callback mqtt.MessageHandler) error
	Disconnect()
}

// ConnectResult represents the outcome of a connection attempt.
type ConnectResult interface {
	Client() Client
	Profile() Profile
	Err() error
}

// ConnectionStatusManager exposes methods to update connection status
// information.
type ConnectionStatusManager interface {
	SetConnecting(name string)
	SetConnected(name string)
	SetDisconnected(name, detail string)
}

// API defines the methods used by the connections component to interact
// with the application without depending on concrete model types.
type API interface {
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
	HandleConnectResult(msg ConnectResult)
	DisconnectActive()
	ResizeTraces(width, height int)
	ResetElemPos()
	SetElemPos(id string, pos int)
	OverlayHelp(view string) string
}

// Navigator exposes navigation helpers required by the component.
type Navigator interface {
	SetMode(mode int) tea.Cmd
	Width() int
	Height() int
}

// Focusable represents a UI element that can gain or lose focus.
type Focusable interface {
	Focus()
	Blur()
	IsFocused() bool
	View() string
}
