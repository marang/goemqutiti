package history

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/marang/emqutiti/ui"
)

// Item represents a single entry in the history list.
type Item struct {
	Timestamp           time.Time
	Topic               string
	Payload             string
	Kind                string // pub, sub, log
	Archived            bool
	IsSelected          *bool
	IsMarkedForDeletion *bool
}

// FilterValue implements list.Item and returns the payload text.
func (h Item) FilterValue() string { return h.Payload }

// Title renders a colored label used by the list delegate.
func (h Item) Title() string {
	var label string
	color := ui.ColBlue
	switch h.Kind {
	case "sub":
		label = "SUB"
		color = ui.ColPink
	case "pub":
		label = "PUB"
		color = ui.ColBlue
	default:
		label = "LOG"
		color = ui.ColGray
	}
	return lipgloss.NewStyle().Foreground(color).Render(
		fmt.Sprintf("%s %s: %s", label, h.Topic, h.Payload),
	)
}

// Description implements list.Item and returns an empty string.
func (h Item) Description() string { return "" }
