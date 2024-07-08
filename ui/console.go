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

// log writes the log message
func (ui *UI) log(message string, isError bool) error {
	if isError {
		message = decorate(message, "red")
	} else {
		message = decorate(message, "green")
	}
	return ui.writeContent(logPanel, message)
}

// clearLog clears the log message.
func (ui *UI) clearLog() error {
	return ui.writeContent(logPanel, "")
}
