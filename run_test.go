package emqutiti

import (
	"flag"
	"os"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	connections "github.com/marang/emqutiti/connections"
	"github.com/marang/emqutiti/help"
	"github.com/marang/emqutiti/history"
	"github.com/marang/emqutiti/importer"
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

func (s *stubHistoryStore) Append(history.Message) {}
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
	origStore, origRun := traceStore, traceRun
	defer func() { traceStore = origStore; traceRun = origRun }()

	st := &stubTraceStore{}
	traceStore = st
	called := false
	traceRun = func(k, tp, pf, stt, end string) error {
		called = true
		if k != "k" || tp != "a,b" || pf != "p" {
			t.Fatalf("unexpected args %v %v %v", k, tp, pf)
		}
		return nil
	}

	if err := runTrace("k", "a,b", "p", "", ""); err != nil {
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
	origStore, origRun := traceStore, traceRun
	defer func() { traceStore = origStore; traceRun = origRun }()

	traceStore = &stubTraceStore{}
	past := time.Now().Add(-time.Hour).Format(time.RFC3339)
	if err := runTrace("k", "t", "p", "", past); err == nil {
		t.Fatalf("expected error for past end time")
	}
}

func resetFlags(args []string) func() {
	origArgs := os.Args
	origFS := flag.CommandLine
	os.Args = append([]string{"cmd"}, args...)
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	registerFlags(flag.CommandLine)
	return func() {
		os.Args = origArgs
		flag.CommandLine = origFS
	}
}

func TestMainDispatchImportFlags(t *testing.T) {
	cases := [][]string{{"-i", "f", "-p", "pr"}, {"--import", "f", "--profile", "pr"}}
	for _, c := range cases {
		undo := resetFlags(c)
		called := false
		origImport, origTrace, origUI := runImportFn, runTraceFn, runUIFn
		runImportFn = func(path, profile string) error {
			called = true
			if path != "f" || profile != "pr" {
				t.Fatalf("unexpected params %v %v", path, profile)
			}
			return nil
		}
		runTraceFn = func(string, string, string, string, string) error {
			t.Fatalf("runTrace called")
			return nil
		}
		runUIFn = func() error {
			t.Fatalf("runUI called")
			return nil
		}
		Main()
		if !called {
			t.Fatalf("runImport not called")
		}
		runImportFn, runTraceFn, runUIFn = origImport, origTrace, origUI
		undo()
	}
}

func TestMainDispatchTrace(t *testing.T) {
	undo := resetFlags([]string{"--trace", "k", "--topics", "t"})
	defer undo()
	called := false
	origImport, origTrace, origUI := runImportFn, runTraceFn, runUIFn
	runTraceFn = func(k, tp, pf, st, en string) error {
		called = true
		if k != "k" || tp != "t" {
			t.Fatalf("unexpected params %v %v", k, tp)
		}
		return nil
	}
	runImportFn = func(string, string) error { t.Fatalf("runImport called"); return nil }
	runUIFn = func() error { t.Fatalf("runUI called"); return nil }
	Main()
	if !called {
		t.Fatalf("runTrace not called")
	}
	runImportFn, runTraceFn, runUIFn = origImport, origTrace, origUI
}

func TestMainDispatchUI(t *testing.T) {
	undo := resetFlags(nil)
	defer undo()
	called := false
	origImport, origTrace, origUI := runImportFn, runTraceFn, runUIFn
	runUIFn = func() error { called = true; return nil }
	runImportFn = func(string, string) error { t.Fatalf("runImport called"); return nil }
	runTraceFn = func(string, string, string, string, string) error {
		t.Fatalf("runTrace called")
		return nil
	}
	Main()
	if !called {
		t.Fatalf("runUI not called")
	}
	runImportFn, runTraceFn, runUIFn = origImport, origTrace, origUI
}

func TestRunImport(t *testing.T) {
	origLoad, origClient, origImp, origProg := loadProfileFn, newMQTTClientFn, newImporterFn, newProgramFn
	defer func() {
		loadProfileFn, newMQTTClientFn, newImporterFn, newProgramFn = origLoad, origClient, origImp, origProg
	}()
	os.Setenv("EMQUTITI_DEFAULT_PASSWORD", "pw")
	defer os.Unsetenv("EMQUTITI_DEFAULT_PASSWORD")

	loadProfileFn = func(name, _ string) (*connections.Profile, error) {
		if name != "pr" {
			t.Fatalf("unexpected profile %s", name)
		}
		return &connections.Profile{}, nil
	}
	client := &stubMQTTClient{}
	newMQTTClientFn = func(p connections.Profile, fn statusFunc) (mqttClient, error) {
		if p.Password != "pw" {
			t.Fatalf("expected password override, got %s", p.Password)
		}
		return client, nil
	}
	newImporterFn = func(cl importer.Publisher, path string) *importer.Model {
		if cl != client || path != "file" {
			t.Fatalf("unexpected importer args")
		}
		return nil
	}
	newProgramFn = func(m tea.Model, opts ...tea.ProgramOption) program {
		return stubProgram{run: func() (tea.Model, error) { return nil, nil }}
	}
	if err := runImport("file", "pr"); err != nil {
		t.Fatalf("runImport error: %v", err)
	}
	if !client.disconnected {
		t.Fatalf("client not disconnected")
	}
}

func TestRunUI(t *testing.T) {
	origInit, origProg := initialModelFn, newProgramFn
	defer func() { initialModelFn, newProgramFn = origInit, origProg }()

	initialModelFn = func(*connections.Connections) (*model, error) {
		return &model{help: &help.Component{}}, nil
	}
	st := &stubHistoryStore{}
	newProgramFn = func(m tea.Model, opts ...tea.ProgramOption) program {
		return stubProgram{run: func() (tea.Model, error) {
			hc := history.NewComponent(nil, st)
			return &model{history: hc}, nil
		}}
	}
	if err := runUI(); err != nil {
		t.Fatalf("runUI error: %v", err)
	}
	if !st.closed {
		t.Fatalf("store not closed")
	}
}
