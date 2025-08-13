package steps

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/marang/emqutiti/ui"
)

// PublishStep processes publishing progress.
type PublishStep struct{ *Base }

func NewPublishStep(b *Base) *PublishStep {
	b.Current = Publish
	return &PublishStep{Base: b}
}

func (s *PublishStep) Update(msg tea.Msg) (Step, tea.Cmd) {
	switch ev := msg.(type) {
	case PublishMsg:
		s.Index++
		p := float64(s.Index) / float64(len(s.Rows))
		if p > 1 {
			p = 1
		}
		cmd := s.Progress.SetPercent(p)
		if s.Index >= len(s.Rows) {
			s.Finished = true
			return s, cmd
		}
		return s, tea.Batch(cmd, s.nextPublishCmd())
	case tea.KeyMsg:
		switch ev.Type {
		case tea.KeyCtrlN:
			if s.Finished {
				return NewDoneStep(s.Base), nil
			}
		case tea.KeyCtrlP:
			if s.Finished {
				s.Finished = false
				return NewReviewStep(s.Base), nil
			}
		}
		cmd := s.History.Update(ev)
		return s, cmd
	}
	return s, nil
}

func (s *PublishStep) View(bw, wrap int) string {
	bar := s.Progress.View()
	lines := s.Published
	limit := s.SampleLimit
	if limit == 0 {
		limit = sampleSize(len(s.Rows))
		s.SampleLimit = limit
	}
	if len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}
	s.History.SetSize(bw, s.HistoryHeight())
	s.History.SetLines(spacedLines(lines))
	recent := s.History.View()
	if recent != "" {
		recent += "\n"
	}
	headerLine := ""
	if s.Finished {
		headerLine = fmt.Sprintf("Published %d messages", len(s.Rows))
	} else {
		headerLine = fmt.Sprintf("Publishing %d/%d", s.Index, len(s.Rows))
	}
	msg := fmt.Sprintf("%s\n%s\n%s", headerLine, bar, recent)
	msg = ansi.Wrap(msg, wrap, " ")
	return ui.LegendBox(msg, "Progress", bw, 0, ui.ColGreen, true, s.History.ScrollPercent())
}
