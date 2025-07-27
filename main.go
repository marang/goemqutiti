package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// type Profile struct {
// 	Name     string `toml:"name"`
// 	Broker   string `toml:"broker"`
// 	ClientId string `toml:"client_id"`
// 	Username string `toml:"username"`
// 	Password string `toml:"password"`
// }

// type Config struct {
// 	DefaultProfile string    `toml:"default_profile"`
// 	Profiles       []Profile `toml:"profiles"`
// }

// func loadConfig() (*Config, error) {
// 	var config Config

// 	homeDir, err := os.UserHomeDir()
// 	if err != nil {
// 		fmt.Println("Error getting home directory:", err)
// 		return nil, err
// 	}
// 	filePath := filepath.Join(homeDir, ".emqutiti", "config.toml")
// 	file, err := os.Open(filePath)
// 	if err != nil {
// 		fmt.Println("Error opening file:", err)
// 		return nil, err
// 	}
// 	defer file.Close()

// 	if _, err = toml.NewDecoder(file).Decode(&config); err != nil {
// 		return nil, err
// 	}

// 	return &config, nil
// }

var (
	importFile  = flag.String("import", "", "Launch import wizard with optional file path")
	profileName = flag.String("profile", "", "Connection profile to use")
)

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

	if *importFile != "" {
		runImport(*importFile, *profileName)
		return
	}

	// Start Bubble Tea UI without connecting. The user can choose a profile
	// from the connection manager once the program starts.
	initial := initialModel(nil)
	initial.mode = modeConnections
	p := tea.NewProgram(initial, tea.WithMouseAllMotion(), tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		log.Fatalf("Error running program: %v", err)
	}
	if m, ok := finalModel.(*model); ok {
		if m.tracer != nil {
			m.tracer.Close()
		}
	}
}

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
		applyEnvVars(p)
	} else if env := os.Getenv("MQTT_PASSWORD"); env != "" {
		p.Password = env
	}

	client, err := NewMQTTClient(*p, nil)
	if err != nil {
		fmt.Println("connect error:", err)
		return
	}
	defer client.Disconnect()

	w := NewWizard(client, path)
	prog := tea.NewProgram(w, tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		fmt.Println("import error:", err)
	}
}
