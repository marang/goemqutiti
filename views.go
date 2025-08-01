package main

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/goemqutiti/ui"
)

func (m *model) overlayHelp(view string) string {
	help := ui.HelpStyle.Render("?")
	if m.help.focused {
		help = ui.HelpFocused.Render("?")
	}
	m.ui.elemPos[idHelp] = 0
	lines := strings.Split(view, "\n")
	if len(lines) == 0 {
		return help
	}
	first := lipgloss.NewStyle().Width(m.ui.width-lipgloss.Width(help)).Render(lines[0]) + help
	if len(lines) == 1 {
		return first
	}
	lines[0] = first
	return strings.Join(lines, "\n")
}

// layoutChips lays out chips horizontally wrapping within width.
func layoutChips(chips []string, width int) ([]string, []chipBound) {
	var lines []string
	var row []string
	var bounds []chipBound
	curX := 0
	rowTop := 0
	chipH := lipgloss.Height(ui.ChipStyle.Render("test"))
	// Each chip is exactly chipH lines tall. Newlines inserted between rows
	// simply stack the rows without extra blank lines, so the vertical offset
	// for the next row increases by chipH.
	rowSpacing := chipH
	for _, c := range chips {
		cw := lipgloss.Width(c)
		if curX+cw > width && len(row) > 0 {
			line := lipgloss.JoinHorizontal(lipgloss.Top, row...)
			line = strings.TrimRightFunc(line, unicode.IsSpace)
			lines = append(lines, line)
			row = []string{}
			curX = 0
			rowTop += rowSpacing
		}
		row = append(row, c)
		bounds = append(bounds, chipBound{xPos: curX, yPos: rowTop, width: cw, height: chipH})
		curX += cw
	}
	if len(row) > 0 {
		line := lipgloss.JoinHorizontal(lipgloss.Top, row...)
		line = strings.TrimRightFunc(line, unicode.IsSpace)
		lines = append(lines, line)
	}
	return lines, bounds
}

// viewClient renders the main client view.
func (m *model) viewClient() string {
	m.ui.elemPos = map[string]int{}
	infoLine := m.clientInfoLine()
	topicsBox, topicBox, bounds := m.clientTopicsSection()
	messageBox := m.clientMessageSection()
	messagesBox := m.clientHistorySection()

	content := lipgloss.JoinVertical(lipgloss.Left, topicsBox, topicBox, messageBox, messagesBox)

	y := 1
	m.ui.elemPos[idTopics] = y
	y += lipgloss.Height(topicsBox)
	m.ui.elemPos[idTopic] = y
	y += lipgloss.Height(topicBox)
	m.ui.elemPos[idMessage] = y
	y += lipgloss.Height(messageBox)
	m.ui.elemPos[idHistory] = y

	startX := 2
	startY := m.ui.elemPos[idTopics] + 1
	m.topics.chipBounds = make([]chipBound, len(bounds))
	for i, b := range bounds {
		m.topics.chipBounds[i] = chipBound{xPos: startX + b.xPos, yPos: startY + b.yPos, width: b.width, height: b.height}
	}

	box := lipgloss.NewStyle().Width(m.ui.width).Padding(0, 1, 1, 1).Render(content)
	m.ui.viewport.SetContent(box)
	m.ui.viewport.Width = m.ui.width
	// Deduct two lines for the info header rendered above the viewport.
	m.ui.viewport.Height = m.ui.height - 2

	view := m.ui.viewport.View()
	return m.overlayHelp(lipgloss.JoinVertical(lipgloss.Left, infoLine, view))
}

// viewConnections shows the list of saved broker profiles.
func (m model) viewConnections() string {
	m.ui.elemPos = map[string]int{}
	m.ui.elemPos[idConnList] = 1
	listView := m.connections.manager.ConnectionsList.View()
	help := ui.InfoStyle.Render("[enter] connect/open client  [x] disconnect  [a]dd [e]dit [del] delete  Ctrl+R traces")
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	view := ui.LegendBox(content, "Brokers", m.ui.width-2, 0, ui.ColBlue, true, -1)
	return m.overlayHelp(view)
}

// viewForm renders the add/edit broker form alongside the list.
func (m model) viewForm() string {
	m.ui.elemPos = map[string]int{}
	m.ui.elemPos[idConnList] = 1
	if m.connections.form == nil {
		return ""
	}
	listView := ui.LegendBox(m.connections.manager.ConnectionsList.View(), "Brokers", m.ui.width/2-2, 0, ui.ColBlue, false, -1)
	formLabel := "Add Broker"
	if m.connections.form.index >= 0 {
		formLabel = "Edit Broker"
	}
	formView := ui.LegendBox(m.connections.form.View(), formLabel, m.ui.width/2-2, 0, ui.ColBlue, true, -1)
	return m.overlayHelp(lipgloss.JoinHorizontal(lipgloss.Top, listView, formView))
}

// viewConfirmDelete displays a confirmation dialog.
func (m model) viewConfirmDelete() string {
	m.ui.elemPos = map[string]int{}
	content := m.confirmPrompt
	if m.confirmInfo != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, m.confirmPrompt, m.confirmInfo)
	}
	content = lipgloss.NewStyle().Padding(1, 2).Render(content)
	box := ui.LegendBox(content, "Confirm", m.ui.width/2, 0, ui.ColBlue, true, -1)
	return lipgloss.Place(m.ui.width, m.ui.height, lipgloss.Center, lipgloss.Center, box)
}

// viewHistoryFilter displays the history filter form.
func (m model) viewHistoryFilter() string {
	m.ui.elemPos = map[string]int{}
	if m.history.filterForm == nil {
		return ""
	}
	content := lipgloss.NewStyle().Padding(1, 2).Render(m.history.filterForm.View())
	box := ui.LegendBox(content, "Filter", m.ui.width/2, 0, ui.ColBlue, true, -1)
	return lipgloss.Place(m.ui.width, m.ui.height, lipgloss.Center, lipgloss.Center, box)
}

// viewTopics displays the topic manager list.
func (m model) viewTopics() string {
	m.ui.elemPos = map[string]int{}
	m.ui.elemPos[idTopicsEnabled] = 1
	m.ui.elemPos[idTopicsDisabled] = 1
	help := ui.InfoStyle.Render("[space] toggle  [del] delete  [esc] back")
	activeView := m.topics.list.View()
	var left, right string
	if m.topics.panes.active == 0 {
		other := list.New(m.unsubscribedItems(), list.NewDefaultDelegate(), m.topics.list.Width(), m.topics.list.Height())
		other.DisableQuitKeybindings()
		other.SetShowTitle(false)
		other.Paginator.Page = m.topics.panes.unsubscribed.page
		other.Select(m.topics.panes.unsubscribed.sel)
		left = ui.LegendBox(activeView, "Enabled", m.ui.width/2-2, 0, ui.ColBlue, m.ui.focusOrder[m.ui.focusIndex] == idTopicsEnabled, -1)
		right = ui.LegendBox(other.View(), "Disabled", m.ui.width/2-2, 0, ui.ColBlue, false, -1)
	} else {
		other := list.New(m.subscribedItems(), list.NewDefaultDelegate(), m.topics.list.Width(), m.topics.list.Height())
		other.DisableQuitKeybindings()
		other.SetShowTitle(false)
		other.Paginator.Page = m.topics.panes.subscribed.page
		other.Select(m.topics.panes.subscribed.sel)
		left = ui.LegendBox(other.View(), "Enabled", m.ui.width/2-2, 0, ui.ColBlue, false, -1)
		right = ui.LegendBox(activeView, "Disabled", m.ui.width/2-2, 0, ui.ColBlue, m.ui.focusOrder[m.ui.focusIndex] == idTopicsDisabled, -1)
	}
	panes := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	content := lipgloss.JoinVertical(lipgloss.Left, panes, help)
	return m.overlayHelp(content)
}

// viewPayloads shows stored payloads for reuse.
func (m model) viewPayloads() string {
	m.ui.elemPos = map[string]int{}
	m.ui.elemPos[idPayloadList] = 1
	listView := m.message.list.View()
	help := ui.InfoStyle.Render("[enter] load  [del] delete  [esc] back")
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	focused := m.ui.focusOrder[m.ui.focusIndex] == idPayloadList
	view := ui.LegendBox(content, "Payloads", m.ui.width-2, 0, ui.ColBlue, focused, -1)
	return m.overlayHelp(view)
}

// viewTraces lists configured traces and their state.
func (m model) viewTraces() string {
	m.ui.elemPos = map[string]int{}
	m.ui.elemPos[idTraceList] = 1
	listView := m.traces.list.View()
	help := ui.InfoStyle.Render("[a] add  [enter] start/stop  [v] view  [del] delete  [esc] back")
	content := lipgloss.JoinVertical(lipgloss.Left, listView, help)
	focused := m.ui.focusOrder[m.ui.focusIndex] == idTraceList
	view := ui.LegendBox(content, "Traces", m.ui.width-2, 0, ui.ColBlue, focused, -1)
	return m.overlayHelp(view)
}

// viewTraceForm renders the form for new traces.
func (m model) viewTraceForm() string {
	m.ui.elemPos = map[string]int{}
	content := m.traces.form.View()
	view := ui.LegendBox(content, "New Trace", m.ui.width-2, 0, ui.ColBlue, true, -1)
	return m.overlayHelp(view)
}

// viewTraceMessages shows captured messages for a trace.
func (m model) viewTraceMessages() string {
	m.ui.elemPos = map[string]int{}
	title := fmt.Sprintf("Trace %s", m.traces.viewKey)
	listLines := strings.Split(m.traces.view.View(), "\n")
	help := ui.InfoStyle.Render("[esc] back")
	listLines = append(listLines, help)
	target := len(listLines)
	minHeight := m.layout.trace.height + 1
	if target < minHeight {
		for len(listLines) < minHeight {
			listLines = append(listLines, "")
		}
		target = minHeight
	}
	content := strings.Join(listLines, "\n")
	view := ui.LegendBox(content, title, m.ui.width-2, target, ui.ColBlue, true, -1)
	return m.overlayHelp(view)
}

// viewImporter renders the importer wizard view.
func (m model) viewImporter() string {
	m.ui.elemPos = map[string]int{}
	if m.importWizard == nil {
		return ""
	}
	return m.overlayHelp(m.importWizard.View())
}

func (m model) viewHelp() string {
	m.ui.elemPos = map[string]int{}
	m.help.vp.SetContent(helpText)
	content := m.help.vp.View()
	sp := -1.0
	if m.help.vp.Height < lipgloss.Height(content) {
		sp = m.help.vp.ScrollPercent()
	}
	box := ui.LegendBox(content, "Help", m.ui.width-2, m.ui.height-2, ui.ColGreen, true, sp)
	return box
}

// View renders the application UI based on the current mode.
func (m *model) View() string {
	switch m.currentMode() {
	case modeClient:
		return m.viewClient()
	case modeConnections:
		return m.viewConnections()
	case modeEditConnection:
		return m.viewForm()
	case modeConfirmDelete:
		return m.viewConfirmDelete()
	case modeTopics:
		return m.viewTopics()
	case modePayloads:
		return m.viewPayloads()
	case modeTracer:
		return m.viewTraces()
	case modeEditTrace:
		return m.viewTraceForm()
	case modeViewTrace:
		return m.viewTraceMessages()
	case modeImporter:
		return m.viewImporter()
	case modeHistoryFilter:
		return m.viewHistoryFilter()
	case modeHelp:
		return m.viewHelp()
	default:
		return ""
	}
}
