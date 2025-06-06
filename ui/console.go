package ui

type AnsiColor int

const (
	Red AnsiColor = iota
	Green
	Yellow
)

// decorate changes the color of a string
func decorate(s string, color AnsiColor) string {
	switch color {
	case Red:
		s = "\x1b[0;31m" + s
	case Green:
		s = "\x1b[0;32m" + s
	case Yellow:
		s = "\x1b[0;33m" + s
	default:
		s = s + "\x1b[0m"
	}

	return s
}

// log writes the log message
func (ui *UI) log(message string, isError bool) error {
	if isError {
		message = decorate(message, Red)
	} else {
		message = decorate(message, Green)
	}
	return ui.writeContent(logPanel, message)
}

// clearLog clears the log message.
func (ui *UI) clearLog() error {
	return ui.writeContent(logPanel, "")
}
