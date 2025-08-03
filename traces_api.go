package emqutiti

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/marang/emqutiti/connections"
)

// TracesAPI exposes model interactions required by tracesComponent.
type TracesAPI interface {
	navigator
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
}

// TraceStore defines persistence and messaging operations for traces.
type TraceStore interface {
	LoadTraces() map[string]TracerConfig
	SaveTraces(map[string]TracerConfig)
	AddTrace(TracerConfig)
	RemoveTrace(string)
	Messages(profile, key string) ([]TracerMessage, error)
	HasData(profile, key string) (bool, error)
	ClearData(profile, key string) error
	LoadCounts(profile, key string, topics []string) (map[string]int, error)
}

// fileTraceStore implements TraceStore using on-disk state and the
// tracer message database. It provides the default application
// behaviour but can be replaced for testing.
type fileTraceStore struct{}

func (fileTraceStore) LoadTraces() map[string]TracerConfig     { return loadTraces() }
func (fileTraceStore) SaveTraces(data map[string]TracerConfig) { saveTraces(data) }
func (fileTraceStore) AddTrace(cfg TracerConfig)               { addTrace(cfg) }
func (fileTraceStore) RemoveTrace(key string)                  { removeTrace(key) }
func (fileTraceStore) Messages(profile, key string) ([]TracerMessage, error) {
	return tracerMessages(profile, key)
}
func (fileTraceStore) HasData(profile, key string) (bool, error) {
	return tracerHasData(profile, key)
}
func (fileTraceStore) ClearData(profile, key string) error {
	return tracerClearData(profile, key)
}
func (fileTraceStore) LoadCounts(profile, key string, topics []string) (map[string]int, error) {
	return tracerLoadCounts(profile, key, topics)
}

// tracesStore exposes the default TraceStore implementation.
func (m *model) tracesStore() TraceStore { return fileTraceStore{} }

// Profiles returns available connection profiles.
func (m *model) Profiles() []connections.Profile { return m.connections.Manager.Profiles }

// ActiveConnection returns the key of the active connection.
func (m *model) ActiveConnection() string { return m.connections.Active }

// SubscribedTopics lists currently subscribed topic names.
func (m *model) SubscribedTopics() []string {
	var topics []string
	for _, tp := range m.topics.Items {
		if tp.Subscribed {
			topics = append(topics, tp.Name)
		}
	}
	return topics
}

// LogHistory forwards messages to the history component.
func (m *model) LogHistory(topic, payload, kind, text string) {
	m.history.Append(topic, payload, kind, text)
}

// TraceHeight returns the current trace panel height.
func (m *model) TraceHeight() int { return m.layout.trace.height }

// SetTraceHeight updates the trace panel height.
func (m *model) SetTraceHeight(h int) { m.layout.trace.height = h }

var _ TracesAPI = (*model)(nil)
