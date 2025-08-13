package emqutiti

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/emqutiti/constants"
)

// handleTabKey moves focus forward.
func (m *model) handleTabKey() tea.Cmd {
	if len(m.ui.focusOrder) > 0 {
		m.focus.Next()
		m.ui.focusIndex = m.focus.Index()
		id := m.ui.focusOrder[m.ui.focusIndex]
		m.SetFocus(id)
		if id == idTopics {
			if len(m.topics.Items) > 0 {
				sel := m.topics.Selected()
				if sel < 0 || sel >= len(m.topics.Items) {
					m.topics.SetSelected(0)
				}
				m.topics.EnsureVisible(m.ui.width - 4)
			} else {
				m.topics.SetSelected(-1)
			}
		}
	}
	return nil
}

// handleShiftTabKey moves focus backward.
func (m *model) handleShiftTabKey() tea.Cmd {
	if len(m.ui.focusOrder) > 0 {
		m.focus.Prev()
		m.ui.focusIndex = m.focus.Index()
		id := m.ui.focusOrder[m.ui.focusIndex]
		m.SetFocus(id)
		if id == idTopics {
			if len(m.topics.Items) > 0 {
				sel := m.topics.Selected()
				if sel < 0 || sel >= len(m.topics.Items) {
					m.topics.SetSelected(0)
				}
				m.topics.EnsureVisible(m.ui.width - 4)
			} else {
				m.topics.SetSelected(-1)
			}
		}
	}
	return nil
}

// handleResizeUpKey reduces the height of the focused pane.
func (m *model) handleResizeUpKey() tea.Cmd {
	id := m.ui.focusOrder[m.ui.focusIndex]
	switch id {
	case idMessage:
		if m.layout.Message.Height > 1 {
			m.layout.Message.Height--
			m.message.Input().SetHeight(m.layout.Message.Height)
		}
	case idHistory:
		if m.layout.History.Height > 1 {
			m.layout.History.Height--
			m.history.List().SetSize(m.ui.width-4, m.layout.History.Height)
		}
	case idTopics:
		if m.layout.Topics.Height > 1 {
			m.layout.Topics.Height--
		}
	}
	return nil
}

// handleResizeDownKey increases the height of the focused pane.
func (m *model) handleResizeDownKey() tea.Cmd {
	id := m.ui.focusOrder[m.ui.focusIndex]
	switch id {
	case idMessage:
		m.layout.Message.Height++
		m.message.Input().SetHeight(m.layout.Message.Height)
	case idHistory:
		m.layout.History.Height++
		m.history.List().SetSize(m.ui.width-4, m.layout.History.Height)
	case idTopics:
		m.layout.Topics.Height++
	}
	return nil
}

// handleModeSwitchKey switches application modes for special key combos.
func (m *model) handleModeSwitchKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case constants.KeyCtrlB:
		if err := m.connections.Manager.LoadProfiles(""); err != nil {
			m.history.Append("", err.Error(), "log", false, err.Error())
		}
		m.connections.RefreshConnectionItems()
		m.connections.SaveCurrent(m.topics.Snapshot(), m.payloads.Snapshot())
		m.traces.SavePlannedTraces()
		return m.SetMode(constants.ModeConnections)
	case constants.KeyCtrlT:
		m.topics.SetActivePane(0)
		m.topics.RebuildActiveTopicList()
		m.topics.SetSelected(0)
		m.topics.List().SetSize(m.ui.width/2-4, m.ui.height-4)
		return m.SetMode(constants.ModeTopics)
	case constants.KeyCtrlP:
		m.payloads.List().SetSize(m.ui.width-4, m.ui.height-4)
		return m.SetMode(constants.ModePayloads)
	case constants.KeyCtrlR:
		m.traces.List().SetSize(m.ui.width-4, m.ui.height-4)
		return m.SetMode(constants.ModeTracer)
	case constants.KeyCtrlL:
		m.logs.SetSize(m.ui.width, m.ui.height)
		m.logs.Focus()
		return m.SetMode(constants.ModeLogs)
	default:
		return nil
	}
}
