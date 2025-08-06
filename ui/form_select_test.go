package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestSelectFieldCyclesOptions(t *testing.T) {
	sf := NewSelectField("one", []string{"one", "two", "three"})
	sf.Focus()

	// cycle backward from first wraps to last
	sf.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if got, want := sf.Index, 2; got != want {
		t.Fatalf("index=%d want=%d", got, want)
	}

	// cycle forward with right wraps around
	sf.Update(tea.KeyMsg{Type: tea.KeyRight})
	if got, want := sf.Index, 0; got != want {
		t.Fatalf("index=%d want=%d", got, want)
	}

	// space advances
	sf.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if got, want := sf.Index, 1; got != want {
		t.Fatalf("index=%d want=%d", got, want)
	}

	// forward again
	sf.Update(tea.KeyMsg{Type: tea.KeyRight})
	if got, want := sf.Index, 2; got != want {
		t.Fatalf("index=%d want=%d", got, want)
	}
}

func TestSelectFieldOptionsView(t *testing.T) {
	sf := NewSelectField("two", []string{"one", "two", "three"})
	if opts := sf.OptionsView(); opts != "" {
		t.Fatalf("expected empty options when unfocused, got %q", opts)
	}

	sf.Focus()
	expected := strings.Join([]string{
		lipgloss.NewStyle().Foreground(ColBlue).Render("one"),
		lipgloss.NewStyle().Foreground(ColPink).Render("two"),
		lipgloss.NewStyle().Foreground(ColBlue).Render("three"),
	}, "\n")
	if opts := sf.OptionsView(); opts != expected {
		t.Fatalf("options view=%q want=%q", opts, expected)
	}
}

func TestSelectFieldReadOnly(t *testing.T) {
	sf := NewSelectField("one", []string{"one", "two"})
	sf.SetReadOnly(true)
	sf.Focus()

	if opts := sf.OptionsView(); opts != "" {
		t.Fatalf("read-only field should not focus")
	}

	sf.Update(tea.KeyMsg{Type: tea.KeyRight})
	if got, want := sf.Value(), "one"; got != want {
		t.Fatalf("value=%s want=%s", got, want)
	}
}
