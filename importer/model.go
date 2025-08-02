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
	"github.com/marang/emqutiti/ui"
)

// Publisher abstracts the MQTT client for publishing.
type Publisher interface {
	Publish(topic string, qos byte, retained bool, payload interface{}) error
}

// Wizard runs an interactive import wizard.
type Model struct {
	step        int
	file        textinput.Model
	headers     []string
	form        ui.Form
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
	rnd         *rand.Rand
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

// New creates a new wizard. A non-empty path pre-fills the file field.
func New(client Publisher, path string) *Model {
	ti := textinput.New()
	ti.Placeholder = "CSV or XLS file"
	ti.Focus()
	ti.SetValue(path)
	tmpl := textinput.New()
	tmpl.Placeholder = "Topic template"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	hv := ui.NewHistoryView(50, 10)
	return &Model{file: ti, tmpl: tmpl, client: client, progress: progress.New(progress.WithDefaultGradient()), history: hv, rnd: r}
}

func (m *Model) Init() tea.Cmd { return textinput.Blink }

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if km, ok := msg.(tea.KeyMsg); ok && km.Type == tea.KeyCtrlD {
		return tea.Quit
	}
	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = v.Width
		m.height = v.Height
		m.progress.Width = v.Width - 4
		m.history.SetSize(v.Width-2, m.historyHeight())
		return nil
	case progress.FrameMsg:
		nm, cmd := m.progress.Update(msg)
		m.progress = nm.(progress.Model)
		return cmd
	}
	switch m.step {
	case stepFile:
		return m.updateFile(msg)
	case stepMap:
		return m.updateMap(msg)
	case stepTemplate:
		return m.updateTemplate(msg)
	case stepReview:
		return m.updateReview(msg)
	case stepPublish:
		return m.updatePublish(msg)
	case stepDone:
		return m.updateDone(msg)
	}
	return nil
}

// Focus satisfies Component.
func (m *Model) Focus() tea.Cmd { return textinput.Blink }

// Blur satisfies Component.
func (m *Model) Blur() {}

// View renders the wizard at the current step.
func (w *Model) View() string {
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
		box = w.viewFile(bw, wrap)
	case stepMap:
		box = w.viewMap(bw, wrap)
	case stepTemplate:
		box = w.viewTemplate(bw, wrap)
	case stepReview:
		box = w.viewReview(bw, wrap)
	case stepPublish:
		box = w.viewPublish(bw, wrap)
	case stepDone:
		box = w.viewDone(bw, wrap)
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, box)
}

// nextPublishCmd publishes the next row or records it during dry run.
func (w *Model) nextPublishCmd() tea.Cmd {
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
			if err := w.client.Publish(topic, 0, false, payload); err != nil {
				errLine := fmt.Sprintf("error publishing %s: %v", topic, err)
				w.published = append(w.published, errLine)
			} else {
				if len(w.published) < limit {
					w.published = append(w.published, line)
				} else if r := w.rnd.Intn(i + 1); r < limit {
					w.published[r] = line
				}
			}
		}
		w.history.SetLines(w.published)
		w.history.GotoBottom()
		return publishMsg{}
	}
}

// renameFields applies mapping names to row keys, leaving originals intact.
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

// mapping returns the column mapping defined by the user.
func (w *Model) mapping() map[string]string {
	m := map[string]string{}
	for i, h := range w.headers {
		m[h] = strings.TrimSpace(w.form.Fields[i].Value())
	}
	return m
}

// stepsView renders the progress header for the wizard.
func (w *Model) stepsView() string {
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

// spacedLines inserts blank lines between each provided line for readability.
func spacedLines(lines []string) []string {
	out := make([]string, 0, len(lines)*2)
	for _, l := range lines {
		out = append(out, l, "")
	}
	return out
}

// sampleSize determines how many sample lines to keep during publishing.
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

// lineLimit calculates the maximum lines of output based on window height.
func (w *Model) lineLimit() int {
	limit := w.height - 6
	if limit < 3 {
		limit = 3
	}
	return limit
}

// historyHeight returns the height of the history view section.
func (w *Model) historyHeight() int {
	h := w.lineLimit()
	if h > 20 {
		h = 20
	}
	return h
}
