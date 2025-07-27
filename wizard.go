package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"goemqutiti/importer"
	"goemqutiti/ui"
)

// Publisher abstracts the MQTT client for publishing.
type Publisher interface {
	Publish(topic string, qos byte, retained bool, payload interface{}) error
}

// Wizard runs an interactive import wizard.
type Wizard struct {
	step      int
	file      textinput.Model
	headers   []string
	fields    []textinput.Model
	focus     int
	tmpl      textinput.Model
	rows      []map[string]string
	index     int
	progress  progress.Model
	client    Publisher
	dryRun    bool
	published []string
	finished  bool
}

const (
	stepFile = iota
	stepMap
	stepTemplate
	stepReview
	stepPublish
	stepDone
)

var wizardSteps = []string{"File", "Map", "Template", "Review", "Publish", "Done"}

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
	if km, ok := msg.(tea.KeyMsg); ok && km.Type == tea.KeyCtrlD {
		return w, tea.Quit
	}
	switch w.step {
	case stepFile:
		var cmd tea.Cmd
		w.file, cmd = w.file.Update(msg)
		if km, ok := msg.(tea.KeyMsg); ok && (km.Type == tea.KeyEnter || km.Type == tea.KeyCtrlN) {
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
				fi.SetValue(k)
				w.fields = append(w.fields, fi)
			}
			if len(w.fields) > 0 {
				w.focus = 0
				w.fields[0].Focus()
			}
			w.step = stepMap
		}
		return w, cmd
	case stepMap:
		if len(w.fields) == 0 {
			return w, nil
		}
		switch m := msg.(type) {
		case tea.KeyMsg:
			switch m.String() {
			case "tab", "shift+tab", "up", "down", "k", "j":
				step := 1
				if m.String() == "shift+tab" || m.String() == "up" || m.String() == "k" {
					step = -1
				}
				w.focus += step
				if w.focus < 0 {
					w.focus = len(w.fields) - 1
				}
				if w.focus >= len(w.fields) {
					w.focus = 0
				}
			case "ctrl+n":
				w.step = stepTemplate
				w.tmpl.Focus()
				return w, nil
			case "ctrl+p":
				w.step = stepFile
				return w, nil
			}
		case tea.MouseMsg:
			if m.Action == tea.MouseActionPress && m.Button == tea.MouseButtonLeft {
				if m.Y >= 1 && m.Y-1 < len(w.fields) {
					w.focus = m.Y - 1
				}
			}
		}
		for i := range w.fields {
			if i == w.focus {
				w.fields[i].Focus()
			} else {
				w.fields[i].Blur()
			}
		}
		var cmd tea.Cmd
		w.fields[w.focus], cmd = w.fields[w.focus].Update(msg)
		if km, ok := msg.(tea.KeyMsg); ok && km.Type == tea.KeyEnter {
			w.step = stepTemplate
			w.tmpl.Focus()
		}
		return w, cmd
	case stepTemplate:
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
	case stepReview:
		if km, ok := msg.(tea.KeyMsg); ok {
			switch km.String() {
			case "p":
				w.dryRun = false
				w.index = 0
				w.published = nil
				w.finished = false
				w.step = stepPublish
				return w, tea.Batch(w.progress.SetPercent(0), w.nextPublishCmd())
			case "d":
				w.dryRun = true
				w.index = 0
				w.published = nil
				w.finished = false
				w.step = stepPublish
				return w, tea.Batch(w.progress.SetPercent(0), w.nextPublishCmd())
			case "e":
				w.step = stepMap
			case "q":
				w.step = stepDone
			case "ctrl+p":
				w.step = stepTemplate
				w.tmpl.Focus()
			}
		}
		return w, nil
	case stepPublish:
		switch m := msg.(type) {
		case publishMsg:
			w.index++
			p := float64(w.index) / float64(len(w.rows))
			if p > 1 {
				p = 1
			}
			cmd := w.progress.SetPercent(p)
			if w.index >= len(w.rows) {
				w.finished = true
				return w, cmd
			}
			return w, tea.Batch(cmd, w.nextPublishCmd())
		case tea.KeyMsg:
			if w.finished {
				switch m.Type {
				case tea.KeyCtrlN:
					w.step = stepDone
				case tea.KeyCtrlP:
					w.step = stepReview
					w.finished = false
				}
			}
		}
		return w, nil
	case stepDone:
		if km, ok := msg.(tea.KeyMsg); ok {
			switch km.Type {
			case tea.KeyCtrlP:
				w.step = stepReview
				w.finished = false
			}
			if km.String() == "q" {
				return w, tea.Quit
			}
		}
		return w, nil
	}
	return w, nil
}

func (w *Wizard) View() string {
	header := w.stepsView()
	var box string
	switch w.step {
	case stepFile:
		box = ui.LegendBox(w.file.View()+"\n[enter] load file  [ctrl+n] next", "Import", 50, true)
	case stepMap:
		colw := 0
		for _, h := range w.headers {
			if w := lipgloss.Width(h); w > colw {
				colw = w
			}
		}
		var b strings.Builder
		for i, h := range w.headers {
			label := h
			if i == w.focus {
				label = focusedStyle.Render(h)
			}
			fmt.Fprintf(&b, "%*s : %s\n", colw, label, w.fields[i].View())
		}
		b.WriteString("\nUse a.b to nest fields\n[enter] continue  [ctrl+n] next  [ctrl+p] back")
		box = ui.LegendBox(b.String(), "Map Columns", 50, true)
	case stepTemplate:
		names := make([]string, len(w.headers))
		for i, h := range w.headers {
			names[i] = "{" + h + "}"
		}
		help := "Available fields: " + strings.Join(names, " ")
		help = ansi.Wrap(help, 48, " ")
		box = ui.LegendBox(w.tmpl.View()+"\n"+help+"\n[enter] continue  [ctrl+n] next  [ctrl+p] back", "Topic Template", 50, true)
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
			line := fmt.Sprintf("%s -> %s", t, string(p))
			previews += ansi.Wrap(line, 48, " ") + "\n"
		}
		s := fmt.Sprintf("Rows: %d\n%s\n[p] publish  [d] dry run  [e] edit  [ctrl+p] back  [q] quit", len(w.rows), previews)
		box = ui.LegendBox(s, "Review", 50, true)
	case stepPublish:
		bar := w.progress.View()
		lines := w.published
		if len(lines) > 5 {
			lines = lines[len(lines)-5:]
		}
		recent := strings.Join(lines, "\n")
		recent = ansi.Wrap(recent, 48, " ")
		if recent != "" {
			recent += "\n"
		}
		if w.finished {
			msg := fmt.Sprintf("Published %d messages\n%s%s[ctrl+n] next  [ctrl+p] back", len(w.rows), recent, bar)
			box = ui.LegendBox(msg, "Progress", 50, true)
		} else {
			msg := fmt.Sprintf("Publishing %d/%d\n%s%s", w.index, len(w.rows), recent, bar)
			box = ui.LegendBox(msg, "Progress", 50, true)
		}
	case stepDone:
		if w.dryRun {
			out := strings.Join(w.published, "\n")
			out = ansi.Wrap(out, 48, " ")
			out += "\n[ctrl+p] back  [q] quit"
			box = ui.LegendBox(out, "Dry Run", 50, true)
		} else if w.finished {
			msg := fmt.Sprintf("Published %d messages\n[ctrl+p] back  [q] quit", len(w.rows))
			box = ui.LegendBox(msg, "Import", 50, true)
		} else {
			box = ui.LegendBox("Done", "Import", 50, true)
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, box)
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
		if w.dryRun || len(w.published) < 5 {
			w.published = append(w.published, fmt.Sprintf("%s -> %s", topic, string(payload)))
		}
		if !w.dryRun {
			w.client.Publish(topic, 0, false, payload)
		}
		return publishMsg{}
	}
}

func renameFields(row map[string]string, mapping map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range row {
		out[k] = v
		if mapped, ok := mapping[k]; ok {
			name := strings.TrimSpace(mapped)
			if name == "" {
				name = k
			}
			if name != k {
				out[name] = v
			}
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

func (w *Wizard) stepsView() string {
	var parts []string
	for i, name := range wizardSteps {
		st := blurredStyle
		if i == w.step {
			st = focusedStyle
		}
		parts = append(parts, st.Render(name))
	}
	return strings.Join(parts, " > ")
}

type publishMsg struct{}
