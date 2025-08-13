package traces

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	connections "github.com/marang/emqutiti/connections"
	"github.com/marang/emqutiti/constants"
)

type testAPI struct{ mode constants.AppMode }

func (t *testAPI) StartConfirm(string, string, func() tea.Cmd, func() tea.Cmd, func()) {}
func (t *testAPI) SetModeClient() tea.Cmd                                              { t.mode = constants.ModeClient; return nil }
func (t *testAPI) SetModeTracer() tea.Cmd                                              { t.mode = constants.ModeTracer; return nil }
func (t *testAPI) SetModeEditTrace() tea.Cmd                                           { t.mode = constants.ModeEditTrace; return nil }
func (t *testAPI) SetModeViewTrace() tea.Cmd                                           { t.mode = constants.ModeViewTrace; return nil }
func (t *testAPI) SetModeTraceFilter() tea.Cmd                                         { t.mode = constants.ModeTraceFilter; return nil }
func (t *testAPI) SetFocus(string) tea.Cmd                                             { return nil }
func (t *testAPI) FocusedID() string                                                   { return "" }
func (t *testAPI) ResetElemPos()                                                       {}
func (t *testAPI) SetElemPos(string, int)                                              {}
func (t *testAPI) OverlayHelp(v string) string                                         { return v }
func (t *testAPI) Profiles() []connections.Profile                                     { return nil }
func (t *testAPI) ActiveConnection() string                                            { return "" }
func (t *testAPI) SubscribedTopics() []string                                          { return nil }
func (t *testAPI) LogHistory(string, string, string, bool, string)                     {}
func (t *testAPI) TraceHeight() int                                                    { return 0 }
func (t *testAPI) SetTraceHeight(int)                                                  {}
func (t *testAPI) Width() int                                                          { return 80 }
func (t *testAPI) Height() int                                                         { return 24 }
func (t *testAPI) NewClient(connections.Profile) (Client, error)                       { return nil, nil }

type noopStore struct{}

func (noopStore) LoadTraces() map[string]TracerConfig                         { return nil }
func (noopStore) SaveTraces(map[string]TracerConfig) error                    { return nil }
func (noopStore) AddTrace(TracerConfig) error                                 { return nil }
func (noopStore) RemoveTrace(string) error                                    { return nil }
func (noopStore) Messages(string, string) ([]TracerMessage, error)            { return nil, nil }
func (noopStore) HasData(string, string) (bool, error)                        { return false, nil }
func (noopStore) ClearData(string, string) error                              { return nil }
func (noopStore) LoadCounts(string, string, []string) (map[string]int, error) { return nil, nil }

func TestEscSetsClientMode(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	api := &testAPI{}
	c := NewComponent(api, State{}, &noopStore{})
	c.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if api.mode != constants.ModeClient {
		t.Fatalf("expected mode %v, got %v", constants.ModeClient, api.mode)
	}
}
