package steps

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/marang/emqutiti/ui"
)

// DoneStep shows final output.
type DoneStep struct{ *Base }

func NewDoneStep(b *Base) *DoneStep {
	b.Current = Done
	return &DoneStep{Base: b}
}

func (s *DoneStep) Update(msg tea.Msg) (Step, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.Type {
		case tea.KeyCtrlP:
			s.Finished = false
			return NewReviewStep(s.Base), nil
		}
		if cmd := s.History.Update(km); cmd != nil {
			return s, cmd
		}
		if km.String() == "q" {
			return s, tea.Quit
		}
	}
	return s, nil
}

func (s *DoneStep) View(bw, wrap int) string {
	if s.DryRun {
		s.History.SetSize(bw, s.HistoryHeight())
		s.History.SetLines(spacedLines(s.Published))
		out := s.History.View()
		out = ansi.Wrap(out, wrap, " ") + "\n[ctrl+p] back  [q] quit"
		return ui.LegendBox(out, "Dry Run", bw, 0, ui.ColGreen, true, s.History.ScrollPercent())
	} else if s.Finished {
		msg := fmt.Sprintf("Published %d messages\n[ctrl+p] back  [q] quit", len(s.Rows))
		msg = ansi.Wrap(msg, wrap, " ")
		return ui.LegendBox(msg, "Import", bw, 0, ui.ColBlue, true, -1)
	}
	return ui.LegendBox("Done", "Import", bw, 0, ui.ColBlue, true, -1)
}
