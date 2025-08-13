package steps

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/marang/emqutiti/ui"
)

// TemplateStep collects the topic template.
type TemplateStep struct{ *Base }

func NewTemplateStep(b *Base) *TemplateStep {
	b.Current = Template
	b.Tmpl.Focus()
	return &TemplateStep{Base: b}
}

func (s *TemplateStep) Update(msg tea.Msg) (Step, tea.Cmd) {
	var cmd tea.Cmd
	s.Tmpl, cmd = s.Tmpl.Update(msg)
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.Type {
		case tea.KeyCtrlP:
			return NewMapStep(s.Base), cmd
		case tea.KeyCtrlN, tea.KeyEnter:
			if strings.TrimSpace(s.Tmpl.Value()) != "" {
				return NewReviewStep(s.Base), cmd
			}
		}
	}
	return s, cmd
}

func (s *TemplateStep) View(bw, wrap int) string {
	names := make([]string, len(s.Headers))
	for i, h := range s.Headers {
		names[i] = "{" + h + "}"
	}
	help := "Available fields: " + strings.Join(names, " ")
	help = ansi.Wrap(help, wrap, " ")
	content := s.Tmpl.View() + "\n" + help + "\n[enter] continue  [ctrl+n] next  [ctrl+p] back"
	return ui.LegendBox(content, "Topic Template", bw, 0, ui.ColBlue, true, -1)
}
