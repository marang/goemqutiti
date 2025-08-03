package traces

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/marang/emqutiti/connections"
)

// IDList identifies the trace list focusable element.
const IDList = "trace-list"

// API defines interactions required by the traces component from the host model.
type API interface {
	SetModeClient() tea.Cmd
	SetModeTracer() tea.Cmd
	SetModeEditTrace() tea.Cmd
	SetModeViewTrace() tea.Cmd
	SetFocus(id string) tea.Cmd
	FocusedID() string
	ResetElemPos()
	SetElemPos(id string, pos int)
	OverlayHelp(view string) string
	StartConfirm(prompt, info string, returnFocus func() tea.Cmd, action func() tea.Cmd, cancel func())
	Profiles() []connections.Profile
	ActiveConnection() string
	SubscribedTopics() []string
	LogHistory(topic, payload, kind, text string)
	TraceHeight() int
	SetTraceHeight(int)
	Width() int
	Height() int
	NewClient(connections.Profile) (Client, error)
}

// Store defines persistence and messaging operations for traces.
type Store interface {
	LoadTraces() map[string]TracerConfig
	SaveTraces(map[string]TracerConfig)
	AddTrace(TracerConfig)
	RemoveTrace(string)
	Messages(profile, key string) ([]TracerMessage, error)
	HasData(profile, key string) (bool, error)
	ClearData(profile, key string) error
	LoadCounts(profile, key string, topics []string) (map[string]int, error)
}
