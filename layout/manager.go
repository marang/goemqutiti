package layout

// Box represents a UI box with a height.
type Box struct {
	Height int
}

// Manager computes layout sizes and stores user-configurable heights.
type Manager struct {
	Message Box
	History Box
	Topics  Box
	Trace   Box
}

// MessageWidth returns the width for message inputs and lists.
func (m Manager) MessageWidth(width int) int {
	return width - 4
}

// TopicsInputWidth returns the width for the topics input. It subtracts space
// for the prompt and cursor so the surrounding box stays on a single line.
func (m Manager) TopicsInputWidth(width int) int {
	w := m.MessageWidth(width) - 3
	if w < 0 {
		return 0
	}
	return w
}

// ConnectionsSize returns the width and height for the connections list.
func (m Manager) ConnectionsSize(width, height int) (int, int) {
	return m.MessageWidth(width), height - 6
}

// HistorySize returns the width and height for the history list. It defaults
// the height when the current value is zero.
func (m *Manager) HistorySize(width, height int) (int, int) {
	if m.History.Height == 0 {
		m.History.Height = (height-1)/3 + 10
	}
	return m.MessageWidth(width), m.History.Height
}

// TraceHeight returns the height for trace views, defaulting if zero.
func (m *Manager) TraceHeight(height int) int {
	if m.Trace.Height == 0 {
		m.Trace.Height = height - 6
	}
	return m.Trace.Height
}

// TraceListSize returns size for trace lists.
func (m Manager) TraceListSize(width, height int) (int, int) {
	return m.MessageWidth(width), height - 4
}

// TopicsListSize returns size for the topics list.
func (m Manager) TopicsListSize(width, height int) (int, int) {
	return width/2 - 4, height - 4
}

// DetailSize returns size for the history detail view.
func (m Manager) DetailSize(width, height int) (int, int) {
	return m.MessageWidth(width), height - 4
}

// ViewportHeight returns the viewport height, reserving two lines for headers.
func (m Manager) ViewportHeight(height int) int {
	return height - 2
}
