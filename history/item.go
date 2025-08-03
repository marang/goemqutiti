package history

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/marang/emqutiti/ui"
)

// Item represents a single entry in the history list.
type Item struct {
	timestamp           time.Time
	topic               string
	payload             string
	kind                string // pub, sub, log
	archived            bool
	isSelected          *bool
	isMarkedForDeletion *bool
}

// FilterValue implements list.Item and returns the payload text.
func (h Item) FilterValue() string { return h.payload }

// Title renders a colored label used by the list delegate.
func (h Item) Title() string {
	var label string
	color := ui.ColBlue
	switch h.kind {
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
		fmt.Sprintf("%s %s: %s", label, h.topic, h.payload),
	)
}

// Description implements list.Item and returns an empty string.
func (h Item) Description() string { return "" }
