package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestTooltipView(t *testing.T) {
	tt := Tooltip{Text: "hint", Width: 10}
	out := tt.View(2, 0)
	t.Logf("\n%s", out)
	if !strings.Contains(out, "hint") {
		t.Fatalf("tooltip missing content")
	}
	lines := strings.Split(out, "\n")
	if len(lines) == 0 || !strings.HasPrefix(lines[0], "  ") {
		t.Fatalf("tooltip not positioned correctly: %q", lines)
	}
}

func TestRenderTooltipFocused(t *testing.T) {
	wantFocus := (Tooltip{Text: "hint", Width: lipgloss.Width("hint"), Style: &TooltipFocused}).View(0, 0)
	gotFocus := RenderTooltip("hint", 0, 0, true)
	if gotFocus != wantFocus {
		t.Fatalf("focused tooltip mismatch\nwant:\n%s\n got:\n%s", wantFocus, gotFocus)
	}
	wantBlur := (Tooltip{Text: "hint", Width: lipgloss.Width("hint"), Style: &TooltipStyle}).View(0, 0)
	gotBlur := RenderTooltip("hint", 0, 0, false)
	if gotBlur != wantBlur {
		t.Fatalf("blurred tooltip mismatch\nwant:\n%s\n got:\n%s", wantBlur, gotBlur)
	}
}
