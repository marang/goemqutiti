package main

import (
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

func main() {
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Failed to open log file:", err)
		return
	}
	defer logFile.Close()
	// Set log output to file
	log.SetOutput(logFile)

	// Start Bubble Tea UI without connecting. The user can choose a profile
	// from the connection manager once the program starts.
	initial := initialModel(nil)
	initial.mode = modeConnections
	p := tea.NewProgram(initial)
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}
