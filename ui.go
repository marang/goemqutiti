package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle  = focusedStyle
	noCursor     = lipgloss.NewStyle()
	borderStyle  = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).Padding(0, 1)
)

type model struct {
	mqttClient   *MQTTClient
	connection   string
	messages     []string
	topicInput   textinput.Model
	messageInput textinput.Model
	payloads     map[string]string
	subscribed   bool
	focusIndex   int
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter Topic"
	ti.Focus()
	ti.CharLimit = 32
	ti.Prompt = "> "
	ti.Cursor.Style = cursorStyle
	ti.TextStyle = focusedStyle
	ti.Width = 40

	mi := textinput.New()
	mi.Placeholder = "Enter Message"
	mi.CharLimit = 128
	mi.Prompt = "> "
	mi.Blur()
	mi.Cursor.Style = noCursor
	mi.TextStyle = blurredStyle
	mi.Width = 80

	return model{
		messages:     make([]string, 0),
		payloads:     make(map[string]string),
		topicInput:   ti,
		messageInput: mi,
		focusIndex:   0,
	}
}

func (m model) Init() tea.Cmd {
	// return nil
	return tea.EnableMouseCellMotion
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			log.Printf("tab event %v", msg)

			if m.focusIndex == 0 {
				m.topicInput.Blur()
				m.messageInput.Focus()
				m.focusIndex = 1
			} else {
				m.messageInput.Blur()
				m.topicInput.Focus()
				m.focusIndex = 0
			}
		case "enter":
			if !m.subscribed {
				m.messages = append(m.messages, fmt.Sprintf("Subscribed to topic: %s", m.topicInput.Value()))
				m.subscribed = true
			} else {
				topic := m.topicInput.Value()
				payload := m.messageInput.Value()
				m.payloads[topic] = payload
				m.messages = append(m.messages, fmt.Sprintf("Published to %s: %s", topic, payload))
			}
		}

		// case tea.MouseMsg:
		// 	switch msg.Action {
		// 	case tea.MouseActionPress:
		// 		log.Printf("Mouse Press: button=%v at (%d, %d)", msg.Button, msg.X, msg.Y)
		// 	case tea.MouseActionRelease:
		// 		log.Printf("Mouse Release: button=%v at (%d, %d)", msg.Button, msg.X, msg.Y)
		// 	case tea.MouseActionMotion:
		// 		log.Printf("Mouse Move to (%d, %d)", msg.X, msg.Y)
		// 	}
	}

	m.topicInput, cmd = m.topicInput.Update(msg)
	m.messageInput, _ = m.messageInput.Update(msg)

	return m, cmd
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString("GoEmqutiti - MQTT Client\n")
	b.WriteString("Connection: " + m.connection + "\n")
	b.WriteString("\nMessages:\n")
	for _, msg := range m.messages {
		b.WriteString("- " + msg + "\n")
	}

	b.WriteString("\nTopic:\n" + borderStyle.Render(m.topicInput.View()))
	b.WriteString("\nMessage:\n" + borderStyle.Render(m.messageInput.View()))

	b.WriteString("\nStored Payloads:\n")
	for topic, payload := range m.payloads {
		b.WriteString(fmt.Sprintf("- %s: %s\n", topic, payload))
	}

	return b.String()
}
