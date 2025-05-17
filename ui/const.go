package ui

type LayoutOption int

const (
	// Main panels
	logoPanel     = "logo_panel"
	diagramsPanel = "diagrams_panel"
	editorPanel   = "editor_panel"
	logPanel      = "log_panel"

	// Modal names
	helpModal     = "help_modal"
	saveModal     = "save_modal"
	layoutModal   = "layout_modal"
	progressModal = "progress_modal"

	// Log messages
	errorEmpty     = "The editor should not be empty!"
	invalidContent = "Cannot display the file content!"

	// Save modal elements
	saveButton   = "save"
	cancelButton = "cancel"

	mainDir = "/diagrams"
)

const (
	defaultLayout LayoutOption = iota
	blackLayout
	blueLayout
	greenLayout
	magentaLayout
	cyanLayout
)

func (o LayoutOption) ToString() string {
	switch o {
	case defaultLayout:
		return "Default"
	case blackLayout:
		return "Black"
	case blueLayout:
		return "Blue"
	case greenLayout:
		return "Green"
	case magentaLayout:
		return "Magenta"
	case cyanLayout:
		return "Cyan"
	default:
		return ""
	}
}
