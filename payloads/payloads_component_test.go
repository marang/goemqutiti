package payloads

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type testModel struct{ clientMode bool }

func (t *testModel) SetClientMode() tea.Cmd        { t.clientMode = true; return nil }
func (t *testModel) FocusedID() string             { return "" }
func (t *testModel) ResetElemPos()                 {}
func (t *testModel) SetElemPos(id string, pos int) {}
func (t *testModel) OverlayHelp(s string) string   { return s }
func (t *testModel) Width() int                    { return 0 }

type testStatus struct{}

func (testStatus) ListenStatus() tea.Cmd { return nil }

func TestEscReturnsClientMode(t *testing.T) {
	m := &testModel{}
	p := New(m, testStatus{})
	p.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !m.clientMode {
		t.Fatalf("expected client mode")
	}
}
