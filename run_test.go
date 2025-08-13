package emqutiti

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	cfg "github.com/marang/emqutiti/cmd"
	connections "github.com/marang/emqutiti/connections"
	"github.com/marang/emqutiti/help"
	"github.com/marang/emqutiti/history"
	"github.com/marang/emqutiti/importer"
	"github.com/marang/emqutiti/proxy"
	"github.com/marang/emqutiti/traces"
)

type stubTraceStore struct {
	traces.Store
	hasData    bool
	hasDataErr error
	addCfg     traces.TracerConfig
	addErr     error
}

func (s *stubTraceStore) LoadTraces() map[string]traces.TracerConfig      { return nil }
func (s *stubTraceStore) SaveTraces(map[string]traces.TracerConfig) error { return nil }
func (s *stubTraceStore) AddTrace(cfg traces.TracerConfig) error {
	s.addCfg = cfg
	return s.addErr
}
func (s *stubTraceStore) RemoveTrace(string) error                                { return nil }
func (s *stubTraceStore) Messages(string, string) ([]traces.TracerMessage, error) { return nil, nil }
func (s *stubTraceStore) HasData(string, string) (bool, error)                    { return s.hasData, s.hasDataErr }
func (s *stubTraceStore) ClearData(string, string) error                          { return nil }
func (s *stubTraceStore) LoadCounts(string, string, []string) (map[string]int, error) {
	return nil, nil
}

type stubProgram struct{ run func() (tea.Model, error) }

func (s stubProgram) Run() (tea.Model, error) { return s.run() }

type stubMQTTClient struct{ disconnected bool }

func (s *stubMQTTClient) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	return nil
}

func (s *stubMQTTClient) Disconnect() { s.disconnected = true }

type stubHistoryStore struct{ closed bool }

func (s *stubHistoryStore) Append(history.Message) error { return nil }
func (s *stubHistoryStore) Search(bool, []string, time.Time, time.Time, string) []history.Message {
	return nil
}
func (s *stubHistoryStore) Delete(string) error  { return nil }
func (s *stubHistoryStore) Archive(string) error { return nil }
func (s *stubHistoryStore) Count(bool) int       { return 0 }
func (s *stubHistoryStore) Close() error {
	s.closed = true
	return nil
}

func TestRunTrace(t *testing.T) {
	st := &stubTraceStore{}
	called := false
	d := &appDeps{
		traceKey:    "k",
		traceTopics: "a,b",
		profileName: "p",
		traceStore:  st,
		traceRun: func(k, tp, pf, stt, end string) error {
			called = true
			if k != "k" || tp != "a,b" || pf != "p" {
				t.Fatalf("unexpected args %v %v %v", k, tp, pf)
			}
			return nil
		},
	}
	if err := runTrace(d); err != nil {
		t.Fatalf("runTrace error: %v", err)
	}
	if !called {
		t.Fatalf("traceRun not called")
	}
	if st.addCfg.Key != "k" || st.addCfg.Profile != "p" {
		t.Fatalf("unexpected cfg: %#v", st.addCfg)
	}
}

func TestRunTraceEndPast(t *testing.T) {
	past := time.Now().Add(-time.Hour).Format(time.RFC3339)
	d := &appDeps{
		traceKey:    "k",
		traceTopics: "t",
		profileName: "p",
		traceEnd:    past,
		traceStore:  &stubTraceStore{},
		traceRun:    func(string, string, string, string, string) error { return nil },
	}
	if err := runTrace(d); err == nil {
		t.Fatalf("expected error for past end time")
	}
}

func TestMainDispatchImportFlags(t *testing.T) {
	orig := initProxy
	initProxy = func() (string, *proxy.Proxy) { return "", nil }
	called := false
	d := newAppDeps()
	d.runImport = func(ad *appDeps) error {
		called = true
		if ad.importFile != "f" || ad.profileName != "pr" {
			t.Fatalf("unexpected params %v %v", ad.importFile, ad.profileName)
		}
		return nil
	}
	d.runTrace = func(*appDeps) error { t.Fatalf("runTrace called"); return nil }
	d.runUI = func(*appDeps) error { t.Fatalf("runUI called"); return nil }
	runMain(d, cfg.AppConfig{ImportFile: "f", ProfileName: "pr"})
	if !called {
		t.Fatalf("runImport not called")
	}
	initProxy = orig
}

func TestMainDispatchTrace(t *testing.T) {
	orig := initProxy
	initProxy = func() (string, *proxy.Proxy) { return "", nil }
	called := false
	d := newAppDeps()
	d.runTrace = func(ad *appDeps) error {
		called = true
		if ad.traceKey != "k" || ad.traceTopics != "t" {
			t.Fatalf("unexpected params %v %v", ad.traceKey, ad.traceTopics)
		}
		return nil
	}
	d.runImport = func(*appDeps) error { t.Fatalf("runImport called"); return nil }
	d.runUI = func(*appDeps) error { t.Fatalf("runUI called"); return nil }
	runMain(d, cfg.AppConfig{TraceKey: "k", TraceTopics: "t"})
	if !called {
		t.Fatalf("runTrace not called")
	}
	initProxy = orig
}

func TestMainDispatchUI(t *testing.T) {
	orig := initProxy
	initProxy = func() (string, *proxy.Proxy) { return "", nil }
	called := false
	d := newAppDeps()
	d.runUI = func(*appDeps) error { called = true; return nil }
	d.runImport = func(*appDeps) error { t.Fatalf("runImport called"); return nil }
	d.runTrace = func(*appDeps) error { t.Fatalf("runTrace called"); return nil }
	runMain(d, cfg.AppConfig{})
	if !called {
		t.Fatalf("runUI not called")
	}
	initProxy = orig
}

func TestRunImport(t *testing.T) {
	t.Setenv("EMQUTITI_DEFAULT_PASSWORD", "pw")

	client := &stubMQTTClient{}
	d := &appDeps{
		importFile:  "file",
		profileName: "pr",
		loadProfile: func(name, _ string) (*connections.Profile, error) {
			if name != "pr" {
				t.Fatalf("unexpected profile %s", name)
			}
			return &connections.Profile{}, nil
		},
		newMQTTClient: func(p connections.Profile, fn statusFunc) (mqttClient, error) {
			if p.Password != "pw" {
				t.Fatalf("expected password override, got %s", p.Password)
			}
			return client, nil
		},
		newImporter: func(cl importer.Publisher, path string) *importer.Model {
			if cl != client || path != "file" {
				t.Fatalf("unexpected importer args")
			}
			return nil
		},
		newProgram: func(m tea.Model, opts ...tea.ProgramOption) program {
			return stubProgram{run: func() (tea.Model, error) { return nil, nil }}
		},
	}
	if err := runImport(d); err != nil {
		t.Fatalf("runImport error: %v", err)
	}
	if !client.disconnected {
		t.Fatalf("client not disconnected")
	}
}

func TestRunUI(t *testing.T) {
	st := &stubHistoryStore{}
	d := &appDeps{
		initialModel: func(*connections.Connections) (*model, error) {
			return &model{help: &help.Component{}}, nil
		},
		newProgram: func(m tea.Model, opts ...tea.ProgramOption) program {
			return stubProgram{run: func() (tea.Model, error) {
				hc := history.NewComponent(nil, st)
				return &model{history: hc}, nil
			}}
		},
	}
	if err := runUI(d); err != nil {
		t.Fatalf("runUI error: %v", err)
	}
	if !st.closed {
		t.Fatalf("store not closed")
	}
}
