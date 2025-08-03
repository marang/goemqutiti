package topics

import tea "github.com/charmbracelet/bubbletea"

type mockModel struct{}

func (mockModel) ShowClient() tea.Cmd        { return nil }
func (mockModel) SetFocus(id string) tea.Cmd { return nil }
func (mockModel) FocusedID() string          { return "" }
func (mockModel) StartConfirm(prompt, info string, returnFocus func() tea.Cmd, action func() tea.Cmd, cancel func()) {
}
func (mockModel) ResetElemPos()                  {}
func (mockModel) SetElemPos(id string, pos int)  {}
func (mockModel) OverlayHelp(view string) string { return view }
func (mockModel) ListenStatus() tea.Cmd          { return nil }
func (mockModel) Width() int                     { return 80 }

func newTestComponent() *Component {
	return New(mockModel{})
}
