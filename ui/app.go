package ui

func InitApp() {
	ui := NewUI()
	defer ui.Close()

	ui.Init()

	// Main Loop
	ui.Loop()
}
