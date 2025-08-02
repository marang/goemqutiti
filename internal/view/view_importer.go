package view

// viewImporter renders the importer wizard view.
func (m model) viewImporter() string {
	m.ui.elemPos = map[string]int{}
	if m.importWizard == nil {
		return ""
	}
	return m.overlayHelp(m.importWizard.View())
}
