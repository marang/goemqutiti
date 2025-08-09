package logs

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/marang/emqutiti/constants"
)

type testNav struct {
	mode constants.AppMode
	prev constants.AppMode
}

func (t *testNav) SetMode(m constants.AppMode) tea.Cmd { t.mode = m; return nil }
func (t *testNav) PreviousMode() constants.AppMode     { return t.prev }
func (t *testNav) Width() int                          { return 80 }
func (t *testNav) Height() int                         { return 24 }

// Ensure Write appends log lines.
func TestWriteAppendsLines(t *testing.T) {
	nav := &testNav{}
	elemPos := map[string]int{}
	c := New(nav, new(int), new(int), &elemPos)
	c.Write([]byte("hello"))
	if len(c.Lines()) != 1 || c.Lines()[0] != "hello" {
		t.Fatalf("unexpected lines: %v", c.Lines())
	}
}

// Ensure Esc returns to previous mode.
func TestEscReturnsPreviousMode(t *testing.T) {
	nav := &testNav{prev: constants.ModeClient}
	elemPos := map[string]int{}
	c := New(nav, new(int), new(int), &elemPos)
	c.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if nav.mode != constants.ModeClient {
		t.Fatalf("expected mode %v, got %v", constants.ModeClient, nav.mode)
	}
}
