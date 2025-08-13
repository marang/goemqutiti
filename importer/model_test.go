package importer

import (
	"errors"
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/marang/emqutiti/importer/steps"
	"github.com/marang/emqutiti/ui"
)

type mockPublisher struct{}

func (m *mockPublisher) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	return nil
}

type errPublisher struct{}

func (m *errPublisher) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	return errors.New("fail")
}

// Test wizard progresses through file, map, template, and publish steps.
func TestModelStepProgression(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	f, err := os.CreateTemp("", "wiz-*.csv")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString("a,b\n1,2\n"); err != nil {
		t.Fatalf("write: %v", err)
	}
	f.Close()

	w := New(&mockPublisher{}, f.Name())

	if _, ok := w.current.(*steps.FileStep); !ok {
		t.Fatalf("expected FileStep")
	}
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if _, ok := w.current.(*steps.MapStep); !ok {
		t.Fatalf("expected MapStep")
	}
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if _, ok := w.current.(*steps.TemplateStep); !ok {
		t.Fatalf("expected TemplateStep")
	}
	w.Base.Tmpl.SetValue("topic/{a}")
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if _, ok := w.current.(*steps.ReviewStep); !ok {
		t.Fatalf("expected ReviewStep")
	}
	cmd := w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	if _, ok := w.current.(*steps.PublishStep); !ok {
		t.Fatalf("expected PublishStep")
	}
	if cmd != nil {
		if msg := cmd(); msg != nil {
			switch m := msg.(type) {
			case tea.BatchMsg:
				for _, c := range m {
					if c != nil {
						if mm := c(); mm != nil {
							w.Update(mm)
						}
					}
				}
			default:
				w.Update(msg)
			}
		}
	}
	if w.Base.Index != 1 {
		t.Fatalf("expected index 1 after first publish, got %d", w.Base.Index)
	}
}

func TestModelFileStep(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	f, err := os.CreateTemp("", "wiz-*.csv")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString("a,b\n1,2\n"); err != nil {
		t.Fatalf("write: %v", err)
	}
	f.Close()

	w := New(&mockPublisher{}, f.Name())
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if _, ok := w.current.(*steps.MapStep); !ok {
		t.Fatalf("expected MapStep")
	}
	if len(w.Base.Headers) != 2 {
		t.Fatalf("expected 2 headers, got %d", len(w.Base.Headers))
	}
}

func TestModelMapNavigation(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	f, err := os.CreateTemp("", "wiz-*.csv")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString("a,b\n1,2\n"); err != nil {
		t.Fatalf("write: %v", err)
	}
	f.Close()

	w := New(&mockPublisher{}, f.Name())
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	w.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	if _, ok := w.current.(*steps.FileStep); !ok {
		t.Fatalf("expected FileStep")
	}
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	w.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	if _, ok := w.current.(*steps.TemplateStep); !ok {
		t.Fatalf("expected TemplateStep")
	}
}

func TestModelReviewPublish(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	f, err := os.CreateTemp("", "wiz-*.csv")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString("a,b\n1,2\n"); err != nil {
		t.Fatalf("write: %v", err)
	}
	f.Close()

	w := New(&mockPublisher{}, f.Name())
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	w.Base.Tmpl.SetValue("topic/{a}")
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	cmd := w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if _, ok := w.current.(*steps.PublishStep); !ok || !w.Base.DryRun {
		t.Fatalf("expected dry run publish, got %T dry %v", w.current, w.Base.DryRun)
	}
	if cmd != nil {
		if msg := cmd(); msg != nil {
			switch m := msg.(type) {
			case tea.BatchMsg:
				for _, c := range m {
					if c != nil {
						if mm := c(); mm != nil {
							w.Update(mm)
						}
					}
				}
			default:
				w.Update(msg)
			}
		}
	}
	if !w.Base.Finished {
		t.Fatalf("expected finished after publishing rows")
	}
	w.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	if _, ok := w.current.(*steps.DoneStep); !ok {
		t.Fatalf("expected DoneStep")
	}
}

func TestModelPublishError(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	f, err := os.CreateTemp("", "wiz-*.csv")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString("a,b\n1,2\n"); err != nil {
		t.Fatalf("write: %v", err)
	}
	f.Close()

	w := New(&errPublisher{}, f.Name())
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	w.Base.Tmpl.SetValue("topic/{a}")
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	cmd := w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	if cmd != nil {
		if msg := cmd(); msg != nil {
			switch m := msg.(type) {
			case tea.BatchMsg:
				for _, c := range m {
					if c != nil {
						if mm := c(); mm != nil {
							w.Update(mm)
						}
					}
				}
			default:
				w.Update(msg)
			}
		}
	}
	if len(w.Base.Published) == 0 || !strings.Contains(w.Base.Published[0], "error") {
		t.Fatalf("expected error message, got %v", w.Base.Published)
	}
	if !w.Base.Finished {
		t.Fatalf("expected finished after processing rows")
	}
}

func TestSettingsRoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	f, err := os.CreateTemp("", "wiz-*.csv")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString("a,b\n1,2\n"); err != nil {
		t.Fatalf("write: %v", err)
	}
	f.Close()

	w := New(&mockPublisher{}, f.Name())
	w.Update(tea.KeyMsg{Type: tea.KeyEnter}) // load file -> stepMap
	if tf, ok := w.Base.Form.Fields[0].(*ui.TextField); ok {
		tf.SetValue("aa")
	}
	w.Update(tea.KeyMsg{Type: tea.KeyEnter}) // to stepTemplate
	w.Base.Tmpl.SetValue("topic/{aa}")
	w.Update(tea.KeyMsg{Type: tea.KeyEnter}) // to stepReview
	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})

	w2 := New(&mockPublisher{}, f.Name())
	w2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if v := w2.Base.Form.Fields[0].Value(); v != "aa" {
		t.Fatalf("expected mapping aa, got %q", v)
	}
	if v := w2.Base.Tmpl.Value(); v != "topic/{aa}" {
		t.Fatalf("expected template topic/{aa}, got %q", v)
	}
}
