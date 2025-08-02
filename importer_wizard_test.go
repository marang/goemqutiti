package emqutiti

import (
	"errors"
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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
func TestImportWizardStepProgression(t *testing.T) {
	f, err := os.CreateTemp("", "wiz-*.csv")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString("a,b\n1,2\n"); err != nil {
		t.Fatalf("write: %v", err)
	}
	f.Close()

	w := NewImportWizard(&mockPublisher{}, f.Name())

	if w.step != stepFile {
		t.Fatalf("expected stepFile, got %d", w.step)
	}
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if w.step != stepMap {
		t.Fatalf("expected stepMap, got %d", w.step)
	}
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if w.step != stepTemplate {
		t.Fatalf("expected stepTemplate, got %d", w.step)
	}
	w.tmpl.SetValue("topic/{a}")
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if w.step != stepReview {
		t.Fatalf("expected stepReview, got %d", w.step)
	}
	_, cmd := w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	if w.step != stepPublish {
		t.Fatalf("expected stepPublish, got %d", w.step)
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
	if w.index != 1 {
		t.Fatalf("expected index 1 after first publish, got %d", w.index)
	}
}

func TestImportWizardFileStep(t *testing.T) {
	f, err := os.CreateTemp("", "wiz-*.csv")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString("a,b\n1,2\n"); err != nil {
		t.Fatalf("write: %v", err)
	}
	f.Close()

	w := NewImportWizard(&mockPublisher{}, f.Name())
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if w.step != stepMap {
		t.Fatalf("expected stepMap, got %d", w.step)
	}
	if len(w.headers) != 2 {
		t.Fatalf("expected 2 headers, got %d", len(w.headers))
	}
}

func TestImportWizardMapNavigation(t *testing.T) {
	f, err := os.CreateTemp("", "wiz-*.csv")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString("a,b\n1,2\n"); err != nil {
		t.Fatalf("write: %v", err)
	}
	f.Close()

	w := NewImportWizard(&mockPublisher{}, f.Name())
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	w.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	if w.step != stepFile {
		t.Fatalf("expected stepFile, got %d", w.step)
	}
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	w.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	if w.step != stepTemplate {
		t.Fatalf("expected stepTemplate, got %d", w.step)
	}
}

func TestImportWizardReviewPublish(t *testing.T) {
	f, err := os.CreateTemp("", "wiz-*.csv")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString("a,b\n1,2\n"); err != nil {
		t.Fatalf("write: %v", err)
	}
	f.Close()

	w := NewImportWizard(&mockPublisher{}, f.Name())
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	w.tmpl.SetValue("topic/{a}")
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_, cmd := w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if w.step != stepPublish || !w.dryRun {
		t.Fatalf("expected dry run publish, got step %d dry %v", w.step, w.dryRun)
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
	if !w.finished {
		t.Fatalf("expected finished after publishing rows")
	}
	w.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	if w.step != stepDone {
		t.Fatalf("expected stepDone, got %d", w.step)
	}
}

func TestImportWizardPublishError(t *testing.T) {
	f, err := os.CreateTemp("", "wiz-*.csv")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString("a,b\n1,2\n"); err != nil {
		t.Fatalf("write: %v", err)
	}
	f.Close()

	w := NewImportWizard(&errPublisher{}, f.Name())
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	w.tmpl.SetValue("topic/{a}")
	w.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_, cmd := w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
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
	if len(w.published) == 0 || !strings.Contains(w.published[0], "error") {
		t.Fatalf("expected error message, got %v", w.published)
	}
	if !w.finished {
		t.Fatalf("expected finished after processing rows")
	}
}
