package emqutiti

import (
	"flag"
	"fmt"
	connections "github.com/marang/emqutiti/connections"
	"log"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/emqutiti/constants"
	"github.com/marang/emqutiti/importer"
	"github.com/marang/emqutiti/traces"
)

type importerTeaModel struct{ *importer.Model }

func (m importerTeaModel) Init() tea.Cmd { return m.Model.Init() }

func (m importerTeaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmd := m.Model.Update(msg)
	return m, cmd
}

func (m importerTeaModel) View() string { return m.Model.View() }

var (
	importFile  string
	profileName string
	traceKey    string
	traceTopics string
	traceStart  string
	traceEnd    string

	traceStore traces.Store = traces.FileStore{}
	traceRun                = traces.Run

	runTraceFn  = runTrace
	runImportFn = runImport
	runUIFn     = runUI
)

func registerFlags(fs *flag.FlagSet) {
	fs.StringVar(&importFile, "import", "", "Launch import wizard with optional file path")
	fs.StringVar(&importFile, "i", "", "(shorthand)")
	fs.StringVar(&profileName, "profile", "", "Connection profile to use")
	fs.StringVar(&profileName, "p", "", "(shorthand)")
	fs.StringVar(&traceKey, "trace", "", "Trace key to store messages under")
	fs.StringVar(&traceTopics, "topics", "", "Comma-separated topics to trace")
	fs.StringVar(&traceStart, "start", "", "Optional RFC3339 trace start time")
	fs.StringVar(&traceEnd, "end", "", "Optional RFC3339 trace end time")
}

// init registers CLI flags for tracing and import modes.
func init() { registerFlags(flag.CommandLine) }

// Main parses flags, sets up logging, and launches the UI or other modes.

func Main() {
	flag.Parse()

	if traceKey != "" {
		if err := runTraceFn(traceKey, traceTopics, profileName, traceStart, traceEnd); err != nil {
			log.Println(err)
		}
		return
	}

	if importFile != "" {
		if err := runImportFn(importFile, profileName); err != nil {
			log.Println(err)
		}
		return
	}

	if err := runUIFn(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}

func runTrace(key, topics, profile, startStr, endStr string) error {
	tlist := strings.Split(topics, ",")
	for i := range tlist {
		tlist[i] = strings.TrimSpace(tlist[i])
	}
	var start, end time.Time
	var err error
	if startStr != "" {
		start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			return fmt.Errorf("invalid trace start time: %w", err)
		}
	}
	if endStr != "" {
		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			return fmt.Errorf("invalid trace end time: %w", err)
		}
		if end.Before(time.Now()) {
			return fmt.Errorf("trace end time already passed")
		}
	}
	exists, err := traceStore.HasData(profile, key)
	if err != nil {
		return fmt.Errorf("trace data check failed: %w", err)
	}
	if exists {
		return fmt.Errorf("trace key already exists")
	}
	cfg := traces.TracerConfig{
		Profile: profile,
		Topics:  tlist,
		Start:   start,
		End:     end,
		Key:     key,
	}
	if err := traceStore.AddTrace(cfg); err != nil {
		return err
	}
	return traceRun(key, topics, profile, startStr, endStr)
}

// runImport launches the interactive import wizard using the provided file
// path and profile name.
func runImport(path, profile string) error {
	p, err := connections.LoadProfile(profile, "")
	if err != nil {
		return fmt.Errorf("error loading profile: %w", err)
	}
	if env := os.Getenv("EMQUTITI_DEFAULT_PASSWORD"); env != "" && !p.FromEnv {
		p.Password = env
	}

	client, err := NewMQTTClient(*p, nil)
	if err != nil {
		return fmt.Errorf("connect error: %w", err)
	}
	defer client.Disconnect()

	w := importer.New(client, path)
	prog := tea.NewProgram(importerTeaModel{w}, tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		return fmt.Errorf("import error: %w", err)
	}
	return nil
}

func runUI() error {
	initial, err := initialModel(nil)
	if err != nil {
		log.Printf("Warning: %v", err)
	}
	_ = initial.SetMode(constants.ModeConnections)
	p := tea.NewProgram(initial, tea.WithMouseAllMotion(), tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return err
	}
	if m, ok := finalModel.(*model); ok {
		if st := m.history.Store(); st != nil {
			st.Close()
		}
	}
	return nil
}
