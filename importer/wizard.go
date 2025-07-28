package importer

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"goemqutiti/ui"
)

// Publisher abstracts the MQTT client for publishing.
type Publisher interface {
	Publish(topic string, qos byte, retained bool, payload interface{}) error
}

// Wizard runs an interactive import wizard.
type Wizard struct {
	step        int
	file        textinput.Model
	headers     []string
	fields      []textinput.Model
	focus       int
	tmpl        textinput.Model
	rows        []map[string]string
	index       int
	progress    progress.Model
	client      Publisher
	dryRun      bool
	published   []string
	finished    bool
	sampleLimit int
	width       int
	height      int
	history     ui.HistoryView
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
	rand.Seed(time.Now().UnixNano())
	hv := ui.NewHistoryView(50, 10)
	return &Wizard{file: ti, tmpl: tmpl, client: client, progress: progress.New(progress.WithDefaultGradient()), history: hv}
}

func (w *Wizard) Init() tea.Cmd { return textinput.Blink }

func (w *Wizard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok && km.Type == tea.KeyCtrlD {
		return w, tea.Quit
	}
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		w.width = m.Width
		w.height = m.Height
		w.progress.Width = m.Width - 4
		w.history.SetSize(m.Width-2, w.historyHeight())
		return w, nil
	case progress.FrameMsg:
		nm, cmd := w.progress.Update(msg)
		w.progress = nm.(progress.Model)
		return w, cmd
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
			rows, err := ReadFile(path)
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
				w.history.GotoTop()
				w.step = stepPublish
				return w, tea.Batch(w.progress.SetPercent(0), w.nextPublishCmd())
			case "d":
				w.dryRun = true
				w.index = 0
				w.published = nil
				w.finished = false
				w.history.GotoTop()
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
			switch m.Type {
			case tea.KeyCtrlN:
				if w.finished {
					w.step = stepDone
				}
			case tea.KeyCtrlP:
				if w.finished {
					w.step = stepReview
					w.finished = false
				}
			}
			cmd := w.history.Update(m)
			return w, cmd
		}
		return w, nil
	case stepDone:
		if km, ok := msg.(tea.KeyMsg); ok {
			switch km.Type {
			case tea.KeyCtrlP:
				w.step = stepReview
				w.finished = false
			}
			if cmd := w.history.Update(km); cmd != nil {
				return w, cmd
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
	bw := w.width - 2
	if bw <= 0 {
		bw = 50
	}
	wrap := bw - 2
	if wrap <= 0 {
		wrap = 1
	}
	w.progress.Width = wrap
	var box string
	switch w.step {
	case stepFile:
		content := w.file.View() + "\n[enter] load file  [ctrl+n] next"
		box = ui.LegendBox(content, "Import", bw, true)
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
				label = ui.FocusedStyle.Render(h)
			}
			fmt.Fprintf(&b, "%*s : %s\n", colw, label, w.fields[i].View())
		}
		b.WriteString("\nUse a.b to nest fields\n[enter] continue  [ctrl+n] next  [ctrl+p] back")
		box = ui.LegendBox(b.String(), "Map Columns", bw, true)
	case stepTemplate:
		names := make([]string, len(w.headers))
		for i, h := range w.headers {
			names[i] = "{" + h + "}"
		}
		help := "Available fields: " + strings.Join(names, " ")
		help = ansi.Wrap(help, wrap, " ")
		content := w.tmpl.View() + "\n" + help + "\n[enter] continue  [ctrl+n] next  [ctrl+p] back"
		box = ui.LegendBox(content, "Topic Template", bw, true)
	case stepReview:
		topic := w.tmpl.Value()
		mapping := w.mapping()
		previews := ""
		max := 3
		if len(w.rows) < max {
			max = len(w.rows)
		}
		for i := 0; i < max; i++ {
			t := BuildTopic(topic, renameFields(w.rows[i], mapping))
			p, _ := RowToJSON(w.rows[i], mapping)
			line := fmt.Sprintf("%s -> %s", t, string(p))
			previews += ansi.Wrap(line, wrap, " ") + "\n"
		}
		s := fmt.Sprintf("Rows: %d\n%s\n[p] publish  [d] dry run  [e] edit  [ctrl+p] back  [q] quit", len(w.rows), previews)
		box = ui.LegendBox(s, "Review", bw, true)
	case stepPublish:
		bar := w.progress.View()
		lines := w.published
		limit := w.sampleLimit
		if limit == 0 {
			limit = sampleSize(len(w.rows))
			w.sampleLimit = limit
		}
		if len(lines) > limit {
			lines = lines[len(lines)-limit:]
		}
		w.history.SetSize(bw, w.historyHeight())
		w.history.SetLines(spacedLines(lines))
		recent := w.history.View()
		if recent != "" {
			recent += "\n"
		}
		headerLine := ""
		if w.finished {
			headerLine = fmt.Sprintf("Published %d messages", len(w.rows))
		} else {
			headerLine = fmt.Sprintf("Publishing %d/%d", w.index, len(w.rows))
		}
		msg := fmt.Sprintf("%s\n%s\n%s", headerLine, bar, recent)
		msg = ansi.Wrap(msg, wrap, " ")
		box = ui.LegendGreenBox(msg, "Progress", bw, true)
	case stepDone:
		if w.dryRun {
			w.history.SetSize(bw, w.historyHeight())
			w.history.SetLines(spacedLines(w.published))
			out := w.history.View()
			out = ansi.Wrap(out, wrap, " ") + "\n[ctrl+p] back  [q] quit"
			box = ui.LegendGreenBox(out, "Dry Run", bw, true)
		} else if w.finished {
			msg := fmt.Sprintf("Published %d messages\n[ctrl+p] back  [q] quit", len(w.rows))
			msg = ansi.Wrap(msg, wrap, " ")
			box = ui.LegendBox(msg, "Import", bw, true)
		} else {
			box = ui.LegendBox("Done", "Import", bw, true)
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
	topic := BuildTopic(w.tmpl.Value(), renameFields(row, mapping))
	payload, _ := RowToJSON(row, mapping)
	limit := w.sampleLimit
	if limit == 0 {
		limit = sampleSize(len(w.rows))
		w.sampleLimit = limit
	}
	i := w.index
	return func() tea.Msg {
		line := fmt.Sprintf("%s -> %s", topic, string(payload))
		if w.dryRun {
			w.published = append(w.published, line)
		} else {
			if len(w.published) < limit {
				w.published = append(w.published, line)
			} else if r := rand.Intn(i + 1); r < limit {
				w.published[r] = line
			}
			w.client.Publish(topic, 0, false, payload)
		}
		w.history.SetLines(w.published)
		w.history.GotoBottom()
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
		st := ui.BlurredStyle
		if i == w.step {
			st = ui.FocusedStyle
		}
		parts = append(parts, st.Render(name))
	}
	return strings.Join(parts, " > ")
}

type publishMsg struct{}

func spacedLines(lines []string) []string {
	out := make([]string, 0, len(lines)*2)
	for _, l := range lines {
		out = append(out, l, "")
	}
	return out
}

func sampleSize(total int) int {
	if total <= 5 {
		return total
	}
	size := int(math.Sqrt(float64(total)))
	if size < 5 {
		size = 5
	}
	if size > 20 {
		size = 20
	}
	return size
}

func (w *Wizard) lineLimit() int {
	limit := w.height - 6
	if limit < 3 {
		limit = 3
	}
	return limit
}

func (w *Wizard) historyHeight() int {
	h := w.lineLimit()
	if h > 20 {
		h = 20
	}
	return h
}
