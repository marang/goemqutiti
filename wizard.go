package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"goemqutiti/importer"
	"goemqutiti/ui"
)

// Publisher abstracts the MQTT client for publishing.
type Publisher interface {
	Publish(topic string, qos byte, retained bool, payload interface{}) error
}

// Wizard runs an interactive import wizard.
type Wizard struct {
	step     int
	file     textinput.Model
	headers  []string
	fields   []textinput.Model
	tmpl     textinput.Model
	rows     []map[string]string
	index    int
	progress progress.Model
	client   Publisher
	dryRun   bool
}

const (
	stepFile = iota
	stepMap
	stepTemplate
	stepReview
	stepPublish
	stepDone
)

// NewWizard creates a new wizard. A non-empty path pre-fills the file field.
func NewWizard(client Publisher, path string) *Wizard {
	ti := textinput.New()
	ti.Placeholder = "CSV or XLS file"
	ti.Focus()
	ti.SetValue(path)
	tmpl := textinput.New()
	tmpl.Placeholder = "Topic template"
	return &Wizard{file: ti, tmpl: tmpl, client: client, progress: progress.New(progress.WithDefaultGradient())}
}

func (w *Wizard) Init() tea.Cmd { return textinput.Blink }

func (w *Wizard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch w.step {
	case stepFile:
		var cmd tea.Cmd
		w.file, cmd = w.file.Update(msg)
		if km, ok := msg.(tea.KeyMsg); ok && km.Type == tea.KeyEnter {
			path := strings.TrimSpace(w.file.Value())
			if path == "" {
				return w, nil
			}
			rows, err := importer.ReadFile(path)
			if err != nil {
				w.file.SetValue(path + " (" + err.Error() + ")")
				return w, nil
			}
			if len(rows) == 0 {
				w.file.SetValue(path + " (no data)")
				return w, nil
			}
			w.rows = rows
			for k := range rows[0] {
				w.headers = append(w.headers, k)
				fi := textinput.New()
				// Default mapping uses the original column name.
				fi.SetValue(k)
				w.fields = append(w.fields, fi)
			}
			w.step = stepMap
		}
		return w, cmd
	case stepMap:
		if len(w.fields) == 0 {
			return w, nil
		}
		for i := range w.fields {
			w.fields[i], _ = w.fields[i].Update(msg)
		}
		if km, ok := msg.(tea.KeyMsg); ok && km.Type == tea.KeyEnter {
			w.step = stepTemplate
		}
		return w, nil
	case stepTemplate:
		var cmd tea.Cmd
		w.tmpl, cmd = w.tmpl.Update(msg)
		if km, ok := msg.(tea.KeyMsg); ok && km.Type == tea.KeyEnter {
			if strings.TrimSpace(w.tmpl.Value()) != "" {
				w.step = stepReview
			}
		}
		return w, cmd
	case stepReview:
		if km, ok := msg.(tea.KeyMsg); ok {
			switch km.String() {
			case "p":
				w.dryRun = false
				w.step = stepPublish
				return w, w.nextPublishCmd()
			case "d":
				w.dryRun = true
				w.step = stepPublish
				return w, w.nextPublishCmd()
			case "e":
				w.step = stepMap
			case "q":
				w.step = stepDone
			}
		}
		return w, nil
	case stepPublish:
		if _, ok := msg.(publishMsg); ok {
			w.index++
			if w.index >= len(w.rows) {
				w.step = stepDone
			} else {
				return w, w.nextPublishCmd()
			}
		}
		return w, nil
	}
	return w, nil
}

func (w *Wizard) View() string {
	switch w.step {
	case stepFile:
		return ui.LegendBox(w.file.View()+"\n[enter] load file", "Import", 50, true)
	case stepMap:
		var s string
		for i, h := range w.headers {
			s += fmt.Sprintf("%s -> %s\n", h, w.fields[i].View())
		}
		s += "\n[enter] continue"
		return ui.LegendBox(s, "Map Columns", 50, true)
	case stepTemplate:
		return ui.LegendBox(w.tmpl.View()+"\n[enter] continue", "Topic Template", 50, true)
	case stepReview:
		topic := w.tmpl.Value()
		mapping := w.mapping()
		previews := ""
		max := 3
		if len(w.rows) < max {
			max = len(w.rows)
		}
		for i := 0; i < max; i++ {
			t := importer.BuildTopic(topic, renameFields(w.rows[i], mapping))
			p, _ := importer.RowToJSON(w.rows[i], mapping)
			previews += fmt.Sprintf("%s -> %s\n", t, string(p))
		}
		s := fmt.Sprintf("Rows: %d\n%s\n[p] publish  [d] dry run  [e] edit  [q] quit", len(w.rows), previews)
		return ui.LegendBox(s, "Review", 50, true)
	case stepPublish:
		p := float64(w.index) / float64(len(w.rows))
		w.progress.SetPercent(p)
		bar := w.progress.View()
		return ui.LegendBox(fmt.Sprintf("Publishing %d/%d\n%s", w.index, len(w.rows), bar), "Progress", 50, true)
	case stepDone:
		return ui.LegendBox("Done", "Import", 50, true)
	}
	return ""
}

func (w *Wizard) nextPublishCmd() tea.Cmd {
	if w.index >= len(w.rows) {
		return nil
	}
	row := w.rows[w.index]
	mapping := w.mapping()
	topic := importer.BuildTopic(w.tmpl.Value(), renameFields(row, mapping))
	payload, _ := importer.RowToJSON(row, mapping)
	return func() tea.Msg {
		if !w.dryRun {
			w.client.Publish(topic, 0, false, payload)
		}
		return publishMsg{}
	}
}

func renameFields(row map[string]string, mapping map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range row {
		if name := mapping[k]; name != "" {
			out[name] = v
		}
	}
	return out
}

func (w *Wizard) mapping() map[string]string {
	m := map[string]string{}
	for i, h := range w.headers {
		m[h] = strings.TrimSpace(w.fields[i].Value())
	}
	return m
}

type publishMsg struct{}
