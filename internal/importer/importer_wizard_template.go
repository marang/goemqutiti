package importer

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/marang/emqutiti/ui"
)

// updateTemplate handles the template entry step.
func (w *ImportWizard) updateTemplate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	w.tmpl, cmd = w.tmpl.Update(msg)
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.Type {
		case tea.KeyCtrlP:
			w.step = stepMap
		case tea.KeyCtrlN, tea.KeyEnter:
			if strings.TrimSpace(w.tmpl.Value()) != "" {
				w.step = stepReview
			}
		}
	}
	return w, cmd
}

// viewTemplate renders the topic template step.
func (w *ImportWizard) viewTemplate(bw, wrap int) string {
	names := make([]string, len(w.headers))
	for i, h := range w.headers {
		names[i] = "{" + h + "}"
	}
	help := "Available fields: " + strings.Join(names, " ")
	help = ansi.Wrap(help, wrap, " ")
	content := w.tmpl.View() + "\n" + help + "\n[enter] continue  [ctrl+n] next  [ctrl+p] back"
	return ui.LegendBox(content, "Topic Template", bw, 0, ui.ColBlue, true, -1)
}
