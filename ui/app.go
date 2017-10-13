package ui

func InitApp(fontpath string) {
	ui := NewUI(fontpath)
	defer ui.Close()

	ui.Init()

	// Main Loop
	ui.Loop()
}
