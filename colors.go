package main

import "github.com/charmbracelet/lipgloss"

// Color palette used throughout the TUI. Keeping these constants in one place
// ensures a consistent look and makes it easy to tweak the theme.
const (
	colPink     = lipgloss.Color("205") // focused elements
	colBlue     = lipgloss.Color("63")  // accents and borders
	colGreen    = lipgloss.Color("34")  // success borders
	colGray     = lipgloss.Color("240") // blurred or secondary text
	colCyan     = lipgloss.Color("51")  // info box borders
	colPurple   = lipgloss.Color("212") // selection highlight borders
	colPub      = lipgloss.Color("219") // published payload text
	colSub      = lipgloss.Color("81")  // subscribed payload text
	colDarkGray = lipgloss.Color("237") // background of selected history lines
)
