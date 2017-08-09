package ui

// decorate changes the color of a string
func decorate(s string, color string) string {
	switch color {
	case "green":
		s = "\x1b[0;32m" + s
	case "red":
		s = "\x1b[0;31m" + s
	default:
		return s
	}
	return s + "\x1b[0m"
}

// Insert log message
func (ui *UI) log(message string, isError bool) error {
	if isError {
		message = decorate(message, "red")
	} else {
		message = decorate(message, "green")
	}
	if err := ui.writeContent(LOG_PANEL, message); err != nil {
		return err
	}
	return nil
}

// Clear log message
func (ui *UI) clearLog() error {
	if err := ui.writeContent(LOG_PANEL, ""); err != nil {
		return err
	}
	return nil
}
