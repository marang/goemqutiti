package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/goemqutiti/ui"
)

type historyItem struct {
	timestamp           time.Time
	topic               string
	payload             string
	kind                string // pub, sub, log
	archived            bool
	isSelected          *bool
	isMarkedForDeletion *bool
}

func (h historyItem) FilterValue() string { return h.payload }

func (h historyItem) Title() string {
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
func (h historyItem) Description() string { return "" }

type historyState struct {
	list            list.Model
	items           []historyItem
	store           *HistoryStore
	selectionAnchor int
	showArchived    bool
	filterForm      *historyFilterForm
	filterQuery     string
	detail          viewport.Model
	detailItem      historyItem
}
