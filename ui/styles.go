package ui

import "github.com/charmbracelet/lipgloss"

var (
	FocusedStyle = lipgloss.NewStyle().Foreground(ColPink)
	BlurredStyle = lipgloss.NewStyle().Foreground(ColGray)
	CursorStyle  = FocusedStyle
	NoCursor     = lipgloss.NewStyle()
	BorderStyle  = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(ColBlue).Padding(0, 1)
	GreenBorder  = BorderStyle.BorderForeground(ColGreen)

	Chip                = lipgloss.NewStyle().Padding(0, 1).MarginRight(1).Border(lipgloss.NormalBorder()).BorderForeground(ColBlue)
	ChipFocused         = Chip.BorderTopForeground(ColPink).BorderLeftForeground(ColPink).Foreground(ColPink)
	ChipInactive        = Chip.BorderForeground(ColGray).Foreground(ColGray)
	ChipInactiveFocused = ChipInactive.BorderTopForeground(ColPink).BorderLeftForeground(ColPink).Foreground(ColPink)
	ChipPublish         = Chip.BorderForeground(ColBlue).Background(ColBlue).Foreground(ColWhite).BorderStyle(lipgloss.InnerHalfBlockBorder())
	ChipPublishFocused  = ChipPublish.BorderTopForeground(ColPink).BorderLeftForeground(ColPink) //.Background(ColPink)

	InfoStyle   = lipgloss.NewStyle().Foreground(ColBlue).PaddingLeft(1)
	ErrorStyle  = lipgloss.NewStyle().Foreground(ColWarn).PaddingLeft(1)
	ConnStyle   = lipgloss.NewStyle().Foreground(ColGray).PaddingLeft(1)
	HelpStyle   = lipgloss.NewStyle().Foreground(ColGreen)
	HelpFocused = HelpStyle.Foreground(ColPink)
	HelpHeader  = lipgloss.NewStyle().Foreground(ColCyan).Bold(true).Underline(true)
	HelpKey     = lipgloss.NewStyle().Foreground(ColGreen).Width(20)
)
