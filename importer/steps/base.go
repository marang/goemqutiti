package steps

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/marang/emqutiti/ui"
)

type Base struct {
	File    textinput.Model
	Headers []string
	Form    ui.Form
	Tmpl    textinput.Model
	Rows    []map[string]string

	Index       int
	Progress    progress.Model
	Client      Publisher
	DryRun      bool
	Published   []string
	Finished    bool
	SampleLimit int
	Width       int
	Height      int
	History     ui.HistoryView
	Rnd         *rand.Rand
	Prefs       WizardPrefs

	Current int
}

func NewBase(client Publisher, path string) *Base {
	ti := textinput.New()
	ti.Placeholder = "CSV or XLS file"
	ti.Focus()
	ti.SetValue(path)
	tmpl := textinput.New()
	tmpl.Placeholder = "Topic template"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	hv := ui.NewHistoryView(50, 10)
	prefs := LoadPrefs()
	if prefs.Template != "" {
		tmpl.SetValue(prefs.Template)
	}
	return &Base{
		File:     ti,
		Tmpl:     tmpl,
		Client:   client,
		Progress: progress.New(progress.WithDefaultGradient()),
		History:  hv,
		Rnd:      r,
		Prefs:    prefs,
	}
}

type WizardPrefs struct {
	Mapping  map[string]string `toml:"mapping"`
	Template string            `toml:"template"`
}

func configFile() (string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			return "", err
		}
	}
	return filepath.Join(home, ".config", "emqutiti", "importer.toml"), nil
}

func LoadPrefs() WizardPrefs {
	fp, err := configFile()
	if err != nil {
		return WizardPrefs{Mapping: map[string]string{}}
	}
	var cfg WizardPrefs
	if _, err := toml.DecodeFile(fp, &cfg); err != nil {
		if cfg.Mapping == nil {
			cfg.Mapping = map[string]string{}
		}
		return cfg
	}
	if cfg.Mapping == nil {
		cfg.Mapping = map[string]string{}
	}
	return cfg
}

func SavePrefs(p WizardPrefs) error {
	fp, err := configFile()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(fp), os.ModePerm); err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(p); err != nil {
		return err
	}
	return os.WriteFile(fp, buf.Bytes(), 0o644)
}

func (b *Base) lineLimit() int {
	limit := b.Height - 6
	if limit < 3 {
		limit = 3
	}
	return limit
}

func (b *Base) HistoryHeight() int {
	h := b.lineLimit()
	if h > 20 {
		h = 20
	}
	return h
}

func (b *Base) mapping() map[string]string {
	m := map[string]string{}
	for i, h := range b.Headers {
		m[h] = strings.TrimSpace(b.Form.Fields[i].Value())
	}
	return m
}

func renameFields(row map[string]string, mapping map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range row {
		if name, ok := mapping[k]; ok && strings.TrimSpace(name) != "" {
			out[name] = v
			if name != k {
				out[k] = v
			}
		} else {
			out[k] = v
		}
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

func (b *Base) nextPublishCmd() tea.Cmd {
	if b.Index >= len(b.Rows) {
		return nil
	}
	row := b.Rows[b.Index]
	mapping := b.mapping()
	topic := BuildTopic(b.Tmpl.Value(), renameFields(row, mapping))
	payload, _ := RowToJSON(row, mapping)
	limit := b.SampleLimit
	if limit == 0 {
		limit = sampleSize(len(b.Rows))
		b.SampleLimit = limit
	}
	i := b.Index
	return func() tea.Msg {
		line := fmt.Sprintf("%s -> %s", topic, string(payload))
		if b.DryRun {
			b.Published = append(b.Published, line)
		} else {
			if err := b.Client.Publish(topic, 0, false, payload); err != nil {
				errLine := fmt.Sprintf("error publishing %s: %v", topic, err)
				b.Published = append(b.Published, errLine)
			} else {
				if len(b.Published) < limit {
					b.Published = append(b.Published, line)
				} else if r := b.Rnd.Intn(i + 1); r < limit {
					b.Published[r] = line
				}
			}
		}
		b.History.SetLines(b.Published)
		b.History.GotoBottom()
		return PublishMsg{}
	}
}

// spacedLines inserts blank lines between each provided line for readability.
func spacedLines(lines []string) []string {
	out := make([]string, 0, len(lines)*2)
	for _, l := range lines {
		out = append(out, l, "")
	}
	return out
}
