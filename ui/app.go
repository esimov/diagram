package ui

// InitApp initialize the CLI application.
func InitApp(fontPath string) {
	ui := NewUI(fontPath)
	defer ui.Close()

	ui.Init()
	ui.Loop()
}
