package ui

import (
	"os"
)

// InitApp initialize the CLI application.
func InitApp(fontPath, content string) {
	ui := NewUI(fontPath)

	// This will close the Gio application, which is running on the main thread.
	defer func() {
		os.Exit(0)
	}()
	defer ui.Close()

	ui.Init(content)
	ui.Loop()
}
