package emqutiti

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	connections "github.com/marang/emqutiti/connections"
	history "github.com/marang/emqutiti/history"

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
)

type program interface{ Run() (tea.Model, error) }

type mqttClient interface {
	importer.Publisher
	Disconnect()
}

type appDeps struct {
	importFile  string
	profileName string
	traceKey    string
	traceTopics string
	traceStart  string
	traceEnd    string

	traceStore traces.Store
	traceRun   func(string, string, string, string, string) error

	loadProfile   func(string, string) (*connections.Profile, error)
	newMQTTClient func(connections.Profile, statusFunc) (mqttClient, error)
	newImporter   func(importer.Publisher, string) *importer.Model
	initialModel  func(*connections.Connections) (*model, error)
	newProgram    func(tea.Model, ...tea.ProgramOption) program

	runTrace  func(*appDeps) error
	runImport func(*appDeps) error
	runUI     func(*appDeps) error

	proxyAddr string
}

func newAppDeps() *appDeps {
	return &appDeps{
		traceStore:    traces.FileStore{},
		traceRun:      traces.Run,
		loadProfile:   connections.LoadProfile,
		newMQTTClient: func(p connections.Profile, fn statusFunc) (mqttClient, error) { return NewMQTTClient(p, fn) },
		newImporter:   importer.New,
		initialModel:  initialModel,
		newProgram: func(m tea.Model, opts ...tea.ProgramOption) program {
			return tea.NewProgram(m, opts...)
		},
		runTrace:  runTrace,
		runImport: runImport,
		runUI:     runUI,
	}
}

func registerFlags(fs *flag.FlagSet) {
	fs.StringVar(&importFile, "import", "", "Launch import wizard with optional file path")
	fs.StringVar(&importFile, "i", "", "(shorthand)")
	fs.StringVar(&profileName, "profile", "", "Connection profile name to use")
	fs.StringVar(&profileName, "p", "", "(shorthand)")
	fs.StringVar(&traceKey, "trace", "", "Trace key name to store messages")
	fs.StringVar(&traceTopics, "topics", "", "Comma-separated topics to trace")
	fs.StringVar(&traceStart, "start", "", "Optional RFC3339 trace start time")
	fs.StringVar(&traceEnd, "end", "", "Optional RFC3339 trace end time")

	fs.Usage = func() {
		w := fs.Output()
		fmt.Fprintf(w, "Usage: %s [flags]\n\n", os.Args[0])
		fmt.Fprintln(w, "General:")
		fmt.Fprintln(w, "  -i, --import FILE     Launch import wizard with optional file path (e.g., -i data.csv)")
		fmt.Fprintln(w, "  -p, --profile NAME    Connection profile name to use (e.g., -p local)")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "Trace:")
		fmt.Fprintln(w, "      --trace KEY       Trace key name to store messages (e.g., --trace run1)")
		fmt.Fprintln(w, "      --topics LIST     Comma-separated topics to trace (e.g., --topics \"sensors/#\")")
		fmt.Fprintln(w, "      --start TIME      Optional RFC3339 trace start time (e.g., --start \"2025-08-05T11:47:00Z\")")
		fmt.Fprintln(w, "      --end TIME        Optional RFC3339 trace end time (e.g., --end \"2025-08-05T11:49:00Z\")")
	}
}

// init registers CLI flags for tracing and import modes.
func init() { registerFlags(flag.CommandLine) }

// Main parses flags, sets up logging, and launches the UI or other modes.

func Main() {
	flag.Parse()
	d := newAppDeps()
	d.importFile = importFile
	d.profileName = profileName
	d.traceKey = traceKey
	d.traceTopics = traceTopics
	d.traceStart = traceStart
	d.traceEnd = traceEnd
	runMain(d)
}

func runMain(d *appDeps) {
	addr, _ := initProxy()
	history.SetProxyAddr(addr)
	traces.SetProxyAddr(addr)
	d.proxyAddr = addr
	if d.traceKey != "" {
		if err := d.runTrace(d); err != nil {
			log.Println(err)
		}
		return
	}

	if d.importFile != "" {
		if err := d.runImport(d); err != nil {
			log.Println(err)
		}
		return
	}

	if err := d.runUI(d); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}

func runTrace(d *appDeps) error {
	tlist := strings.Split(d.traceTopics, ",")
	for i := range tlist {
		tlist[i] = strings.TrimSpace(tlist[i])
	}
	var start, end time.Time
	var err error
	if d.traceStart != "" {
		start, err = time.Parse(time.RFC3339, d.traceStart)
		if err != nil {
			return fmt.Errorf("invalid trace start time: %w", err)
		}
	}
	if d.traceEnd != "" {
		end, err = time.Parse(time.RFC3339, d.traceEnd)
		if err != nil {
			return fmt.Errorf("invalid trace end time: %w", err)
		}
		if end.Before(time.Now()) {
			return fmt.Errorf("trace end time already passed")
		}
	}
	exists, err := d.traceStore.HasData(d.profileName, d.traceKey)
	if err != nil {
		return fmt.Errorf("trace data check failed: %w", err)
	}
	if exists {
		return fmt.Errorf("trace key already exists")
	}
	cfg := traces.TracerConfig{
		Profile: d.profileName,
		Topics:  tlist,
		Start:   start,
		End:     end,
		Key:     d.traceKey,
	}
	if err := d.traceStore.AddTrace(cfg); err != nil {
		return err
	}
	return d.traceRun(d.traceKey, d.traceTopics, d.profileName, d.traceStart, d.traceEnd)
}

// runImport launches the interactive import wizard using the provided file
// path and profile name.
func runImport(d *appDeps) error {
	p, err := d.loadProfile(d.profileName, "")
	if err != nil {
		return fmt.Errorf("error loading profile: %w", err)
	}
	if env := os.Getenv("EMQUTITI_DEFAULT_PASSWORD"); env != "" && !p.FromEnv {
		p.Password = env
	}

	client, err := d.newMQTTClient(*p, nil)
	if err != nil {
		return fmt.Errorf("connect error: %w", err)
	}
	defer client.Disconnect()

	w := d.newImporter(client, d.importFile)
	prog := d.newProgram(importerTeaModel{w}, tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		return fmt.Errorf("import error: %w", err)
	}
	return nil
}

func runUI(d *appDeps) error {
	initial, err := d.initialModel(nil)
	if err != nil {
		log.Printf("Warning: %v", err)
	}
	log.SetFlags(0)
	log.SetOutput(initial.logs)
	stop := startProxyStatusLogger(d.proxyAddr)
	defer stop()
	_ = initial.SetMode(constants.ModeConnections)
	p := d.newProgram(initial, tea.WithMouseAllMotion(), tea.WithAltScreen())
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
