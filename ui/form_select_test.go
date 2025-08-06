package ui

import (
	"strings"
	"testing"
)

func TestNewSelectFieldNoOptions(t *testing.T) {
	sf, err := NewSelectField("", nil)
	if err == nil || sf != nil {
		t.Fatalf("expected error for empty options")
	}
}

func TestSelectFieldEmptyOptions(t *testing.T) {
	sf := &SelectField{}
	if v := sf.Value(); v != "" {
		t.Fatalf("expected empty value, got %q", v)
	}
	view := sf.View()
	if !strings.Contains(view, "-") {
		t.Fatalf("expected placeholder in view, got %q", view)
	}
}
