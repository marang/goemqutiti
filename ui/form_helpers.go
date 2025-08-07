package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// FormField represents a labeled form field with optional help text.
type FormField struct {
	Label    string
	Field    Field
	HelpText string
	Error    string
}

// FormLayout renders a form with proper spacing and labeling.
type FormLayout struct {
	Fields     []FormField
	focusIndex int
}

// View renders the form layout with labels, fields, and help text.
func (fl *FormLayout) View() string {
	var sections []string

	for i, field := range fl.Fields {
		var lines []string

		// Add label
		labelStyle := FormLabel
		if i == fl.focusIndex && !field.Field.ReadOnly() {
			labelStyle = FormLabelFocused
		}
		lines = append(lines, labelStyle.Render(field.Label+":"))

		// Add field
		fieldView := field.Field.View()
		if field.Field.ReadOnly() {
			fieldView = ReadOnlyIndicator.Render("(read-only) ") + fieldView
		}
		lines = append(lines, "  "+fieldView)

		// Add help text if focused
		if i == fl.focusIndex && field.HelpText != "" {
			lines = append(lines, "  "+FormHelp.Render(field.HelpText))
		}

		// Add error if present
		if field.Error != "" {
			lines = append(lines, "  "+FormError.Render(field.Error))
		}

		sections = append(sections, strings.Join(lines, "\n"))
	}

	return strings.Join(sections, "\n\n")
}

// Update updates the focused field.
func (fl *FormLayout) Update(msg tea.Msg) tea.Cmd {
	if fl.focusIndex >= 0 && fl.focusIndex < len(fl.Fields) {
		return fl.Fields[fl.focusIndex].Field.Update(msg)
	}
	return nil
}

// ApplyFocus applies focus to the current field.
func (fl *FormLayout) ApplyFocus() {
	for i := range fl.Fields {
		if i == fl.focusIndex {
			fl.Fields[i].Field.Focus()
		} else {
			fl.Fields[i].Field.Blur()
		}
	}
}

// CycleFocus moves focus between fields.
func (fl *FormLayout) CycleFocus(forward bool) {
	if forward {
		fl.focusIndex++
		if fl.focusIndex >= len(fl.Fields) {
			fl.focusIndex = 0
		}
	} else {
		fl.focusIndex--
		if fl.focusIndex < 0 {
			fl.focusIndex = len(fl.Fields) - 1
		}
	}
}

// Value returns the value of the focused field.
func (fl *FormLayout) Value() string {
	if fl.focusIndex >= 0 && fl.focusIndex < len(fl.Fields) {
		return fl.Fields[fl.focusIndex].Field.Value()
	}
	return ""
}
