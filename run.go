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

// init registers CLI flags for tracing and import modes.
func init() {
	flag.StringVar(&importFile, "import", "", "Launch import wizard with optional file path")
	flag.StringVar(&importFile, "i", "", "(shorthand)")
	flag.StringVar(&profileName, "profile", "", "Connection profile to use")
	flag.StringVar(&profileName, "p", "", "(shorthand)")
	flag.StringVar(&traceKey, "trace", "", "Trace key to store messages under")
	flag.StringVar(&traceTopics, "topics", "", "Comma-separated topics to trace")
	flag.StringVar(&traceStart, "start", "", "Optional RFC3339 trace start time")
	flag.StringVar(&traceEnd, "end", "", "Optional RFC3339 trace end time")
}

// Main parses flags, sets up logging, and launches the UI or other modes.

func Main() {
	flag.Parse()

	if traceKey != "" {
		tlist := strings.Split(traceTopics, ",")
		for i := range tlist {
			tlist[i] = strings.TrimSpace(tlist[i])
		}
		var start, end time.Time
		if traceStart != "" {
			var err error
			start, err = time.Parse(time.RFC3339, traceStart)
			if err != nil {
				log.Println("invalid trace start time:", err)
				return
			}
		}
		if traceEnd != "" {
			var err error
			end, err = time.Parse(time.RFC3339, traceEnd)
			if err != nil {
				log.Println("invalid trace end time:", err)
				return
			}
			if end.Before(time.Now()) {
				log.Println("trace end time already passed")
				return
			}
		}
		exists, err := traces.FileStore{}.HasData(profileName, traceKey)
		if err != nil {
			log.Println("trace data check failed:", err)
			return
		}
		if exists {
			log.Println("trace key already exists")
			return
		}
		if err := (traces.FileStore{}).AddTrace(traces.TracerConfig{
			Profile: profileName,
			Topics:  tlist,
			Start:   start,
			End:     end,
			Key:     traceKey,
		}); err != nil {
			log.Println(err)
			return
		}
		if err := traces.Run(traceKey, traceTopics, profileName, traceStart, traceEnd); err != nil {
			log.Println(err)
		}
		return
	}

	if importFile != "" {
		if err := runImport(importFile, profileName); err != nil {
			log.Println(err)
		}
		return
	}

	// Start Bubble Tea UI without connecting. The user can choose a profile
	// from the connection manager once the program starts.
	initial, err := initialModel(nil)
	if err != nil {
		log.Printf("Warning: %v", err)
	}
	_ = initial.setMode(modeConnections)
	p := tea.NewProgram(initial, tea.WithMouseAllMotion(), tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		log.Fatalf("Error running program: %v", err)
	}
	if m, ok := finalModel.(*model); ok {
		if st := m.history.Store(); st != nil {
			st.Close()
		}
	}
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
