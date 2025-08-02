package emqutiti

import (
	"github.com/charmbracelet/bubbles/textarea"
)

type payloadItem struct{ topic, payload string }

func (p payloadItem) FilterValue() string { return p.topic }
func (p payloadItem) Title() string       { return p.topic }
func (p payloadItem) Description() string { return p.payload }

type messageState struct {
	input textarea.Model
}

func (m *messageState) setPayload(payload string) { m.input.SetValue(payload) }
