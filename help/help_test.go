package help

import (
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type testNav struct{ mode int }

func (t *testNav) SetMode(mode int) tea.Cmd { t.mode = mode; return nil }
func (t *testNav) PreviousMode() int        { return 7 }
func (t *testNav) Width() int               { return 0 }
func (t *testNav) Height() int              { return 0 }

func TestEscReturnsToPreviousMode(t *testing.T) {
	w, h := 0, 0
	elemPos := map[string]int{}
	nav := &testNav{}
	c := New(nav, &w, &h, &elemPos)
	c.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if nav.mode != nav.PreviousMode() {
		t.Fatalf("expected mode %d, got %d", nav.PreviousMode(), nav.mode)
	}
}

func TestFocusablesExposeID(t *testing.T) {
	w, h := 0, 0
	elemPos := map[string]int{}
	c := New(&testNav{}, &w, &h, &elemPos)
	if _, ok := c.Focusables()[ID]; !ok {
		t.Fatalf("focusables missing %s", ID)
	}
}

func TestRenderHelpGroupsSections(t *testing.T) {
	txt := renderHelp()
	ansi := regexp.MustCompile("\x1b\\[[0-9;]*m")
	plain := ansi.ReplaceAllString(txt, "")
	expected := []string{"Global", "Navigation", "Broker Manager", "History View", "Tips"}
	for _, e := range expected {
		if !strings.Contains(plain, e) {
			t.Fatalf("missing section %q in help text", e)
		}
	}
}
