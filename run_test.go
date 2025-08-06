package emqutiti

import (
	"flag"
	"os"
	"testing"
	"time"

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
