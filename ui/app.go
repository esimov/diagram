package ui

import (
	"os"
)

// InitApp initialize the CLI application.
func InitApp(fontPath string, defaultContent []byte) {
	ui := NewUI(fontPath)

	// This will close the Gio application, which is running on the main thread.
	defer func() {
		os.Exit(0)
	}()
	defer ui.Close()

	ui.Init(defaultContent)
	ui.Loop()
}
