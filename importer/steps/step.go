package steps

import tea "github.com/charmbracelet/bubbletea"

// Step represents a wizard step.
type Step interface {
	Update(tea.Msg) (Step, tea.Cmd)
	View(int, int) string
}

// Publisher abstracts the MQTT client for publishing.
type Publisher interface {
	Publish(topic string, qos byte, retained bool, payload interface{}) error
}

const (
	File = iota
	Map
	Template
	Review
	Publish
	Done
)

var Names = []string{"File", "Map", "Template", "Review", "Publish", "Done"}

// PublishMsg signals that the next row should be processed during publishing.
type PublishMsg struct{}
