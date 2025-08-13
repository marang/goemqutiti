package steps

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/marang/emqutiti/constants"
	"github.com/marang/emqutiti/ui"
)

// ReviewStep previews messages before publishing.
type ReviewStep struct{ *Base }

func NewReviewStep(b *Base) *ReviewStep {
	b.Current = Review
	return &ReviewStep{Base: b}
}

func (s *ReviewStep) Update(msg tea.Msg) (Step, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case constants.KeyP:
			s.Prefs.Mapping = s.mapping()
			s.Prefs.Template = s.Tmpl.Value()
			SavePrefs(s.Prefs)
			s.DryRun = false
			s.Index = 0
			s.Published = nil
			s.Finished = false
			s.History.GotoTop()
			return NewPublishStep(s.Base), tea.Batch(s.Progress.SetPercent(0), s.nextPublishCmd())
		case constants.KeyD:
			s.Prefs.Mapping = s.mapping()
			s.Prefs.Template = s.Tmpl.Value()
			SavePrefs(s.Prefs)
			s.DryRun = true
			s.Index = 0
			s.Published = nil
			s.Finished = false
			s.History.GotoTop()
			return NewPublishStep(s.Base), tea.Batch(s.Progress.SetPercent(0), s.nextPublishCmd())
		case constants.KeyE:
			return NewMapStep(s.Base), nil
		case constants.KeyQ:
			return NewDoneStep(s.Base), nil
		case constants.KeyCtrlP:
			s.Tmpl.Focus()
			return NewTemplateStep(s.Base), nil
		}
	}
	return s, nil
}

func (s *ReviewStep) View(bw, wrap int) string {
	topic := s.Tmpl.Value()
	mapping := s.mapping()
	previews := ""
	max := 3
	if len(s.Rows) < max {
		max = len(s.Rows)
	}
	for i := 0; i < max; i++ {
		t := BuildTopic(topic, renameFields(s.Rows[i], mapping))
		p, _ := RowToJSON(s.Rows[i], mapping)
		line := fmt.Sprintf("%s -> %s", t, string(p))
		previews += ansi.Wrap(line, wrap, " ") + "\n"
	}
	out := fmt.Sprintf("Rows: %d\n%s\n[p] publish  [d] dry run  [e] edit  [ctrl+p] back  [q] quit", len(s.Rows), previews)
	return ui.LegendBox(out, "Review", bw, 0, ui.ColBlue, true, -1)
}
