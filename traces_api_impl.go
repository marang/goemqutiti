package emqutiti

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/marang/emqutiti/connections"
	"github.com/marang/emqutiti/traces"

	"github.com/marang/emqutiti/constants"
)

func (m *model) tracesStore() traces.Store { return traces.FileStore{} }

func (m *model) SetModeClient() tea.Cmd      { return m.SetMode(constants.ModeClient) }
func (m *model) SetModeTracer() tea.Cmd      { return m.SetMode(constants.ModeTracer) }
func (m *model) SetModeEditTrace() tea.Cmd   { return m.SetMode(constants.ModeEditTrace) }
func (m *model) SetModeViewTrace() tea.Cmd   { return m.SetMode(constants.ModeViewTrace) }
func (m *model) SetModeTraceFilter() tea.Cmd { return m.SetMode(constants.ModeTraceFilter) }

func (m *model) Profiles() []connections.Profile { return m.connections.Manager.Profiles }

func (m *model) ActiveConnection() string { return m.connections.Active }

func (m *model) SubscribedTopics() []string {
	var topics []string
	for _, tp := range m.topics.Items {
		if tp.Subscribed {
			topics = append(topics, tp.Name)
		}
	}
	return topics
}

func (m *model) LogHistory(topic, payload, kind, text string) {
	m.history.Append(topic, payload, kind, text)
}

func (m *model) TraceHeight() int { return m.layout.trace.height }

func (m *model) SetTraceHeight(h int) { m.layout.trace.height = h }

func (m *model) NewClient(p connections.Profile) (traces.Client, error) {
	return NewMQTTClient(p, nil)
}

var _ traces.API = (*model)(nil)
