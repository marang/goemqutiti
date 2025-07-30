package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/goemqutiti/config"
)

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

// main parses flags, sets up logging, and launches the UI or other modes.

func main() {
	flag.Parse()

	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Failed to open log file:", err)
		return
	}
	defer logFile.Close()
	// Set log output to file
	log.SetOutput(logFile)

	if traceKey != "" {
		tlist := strings.Split(traceTopics, ",")
		for i := range tlist {
			tlist[i] = strings.TrimSpace(tlist[i])
		}
		var start, end time.Time
		if traceStart != "" {
			start, _ = time.Parse(time.RFC3339, traceStart)
		}
		if traceEnd != "" {
			end, _ = time.Parse(time.RFC3339, traceEnd)
			if end.Before(time.Now()) {
				fmt.Println("trace end time already passed")
				return
			}
		}
		exists, _ := tracerHasData(profileName, traceKey)
		if exists {
			fmt.Println("trace key already exists")
			return
		}
		addTrace(TracerConfig{Profile: profileName, Topics: tlist, Start: start, End: end, Key: traceKey})
		if err := tracerRun(traceKey, traceTopics, profileName, traceStart, traceEnd); err != nil {
			fmt.Println(err)
		}
		return
	}

	if importFile != "" {
		runImport(importFile, profileName)
		return
	}

	// Start Bubble Tea UI without connecting. The user can choose a profile
	// from the connection manager once the program starts.
	initial := initialModel(nil)
	initial.setMode(modeConnections)
	p := tea.NewProgram(initial, tea.WithMouseAllMotion(), tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		log.Fatalf("Error running program: %v", err)
	}
	if m, ok := finalModel.(*model); ok {
		if m.history.store != nil {
			m.history.store.Close()
		}
	}
}

// runImport launches the interactive import wizard using the provided file
// path and profile name.
func runImport(path, profile string) {
	conns := NewConnectionsModel()
	if err := conns.LoadProfiles(""); err != nil {
		fmt.Println("Error loading profiles:", err)
		return
	}
	var p *Profile
	if profile != "" {
		for i := range conns.Profiles {
			if conns.Profiles[i].Name == profile {
				p = &conns.Profiles[i]
				break
			}
		}
	} else if conns.DefaultProfileName != "" {
		for i := range conns.Profiles {
			if conns.Profiles[i].Name == conns.DefaultProfileName {
				p = &conns.Profiles[i]
				break
			}
		}
	}
	if p == nil && len(conns.Profiles) > 0 {
		p = &conns.Profiles[0]
	}
	if p == nil {
		fmt.Println("no connection profile available")
		return
	}
	if p.FromEnv {
		config.ApplyEnvVars(p)
	} else if env := os.Getenv("MQTT_PASSWORD"); env != "" {
		p.Password = env
	}

	client, err := NewMQTTClient(*p, nil)
	if err != nil {
		fmt.Println("connect error:", err)
		return
	}
	defer client.Disconnect()

	w := NewImportWizard(client, path)
	prog := tea.NewProgram(w, tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		fmt.Println("import error:", err)
	}
}
