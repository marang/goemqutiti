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

	// Load configuration
	config, err := LoadFromConfig("")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Find the default profile
	var profile Profile
	for _, p := range config.Profiles {
		if p.Name == config.DefaultProfileName {
			profile = p
			break
		}
	}

	if profile.Name == "" {
		log.Fatalf("Default profile '%s' not found", config.DefaultProfileName)
	}

	// Override password with environment variable if set
	envPassword := os.Getenv("MQTT_PASSWORD")
	var password string
	if envPassword != "" {
		password = envPassword
	} else {
		password = profile.Password // Already []byte, no conversion needed
	}

	// Initialize MQTT client
	mqttClient, err := NewMQTTClient(profile.Broker, profile.ClientID, profile.Username, password)
	if err != nil {
		log.Fatalf("Failed to connect to MQTT broker: %v", err)
	}
	defer mqttClient.Client.Disconnect(250)

	// Start Bubble Tea UI
	initial := initialModel()
	initial.connection = "Connected to " + profile.Broker
	p := tea.NewProgram(initial)
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}
