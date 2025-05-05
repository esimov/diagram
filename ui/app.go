package ui

import "os"

// InitApp initialize the CLI application.
func InitApp(fontPath string) {
	ui := NewUI(fontPath)

	defer func() {
		os.Exit(0)
	}()
	defer ui.Close()

	ui.Init()
	ui.Loop()
}
