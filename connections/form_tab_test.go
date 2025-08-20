package connections

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Test that Tab and Shift+Tab cycle focus through form fields.
func TestFormCyclesFocusWithTab(t *testing.T) {
	f := NewForm(Profile{}, -1)
	if f.Focus != 0 {
		t.Fatalf("initial focus=%d want=0", f.Focus)
	}
	f, _ = f.Update(tea.KeyMsg{Type: tea.KeyTab})
	if f.Focus != 1 {
		t.Fatalf("after Tab focus=%d want=1", f.Focus)
	}
	f, _ = f.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if f.Focus != 0 {
		t.Fatalf("after Shift+Tab focus=%d want=0", f.Focus)
	}
}
