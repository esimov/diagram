package ui

import (
	"github.com/jroimartin/gocui"
)

type buttonWidget struct {
	name 	string
	x, y 	int
	w	int
	label	string
	handler	func(*gocui.Gui, *gocui.View) error
}

func NewButtonWidget(name string, x, y int, label string, handler func(g *gocui.Gui, v *gocui.View) error) *buttonWidget {
	return &buttonWidget{name, x, y, len(label)+1, label, handler}
}

func (ui *UI) createButtonWidget(name string, x, y int, label string, handler func(g *gocui.Gui, v *gocui.View) error) (*gocui.View, error) {
	button := NewButtonWidget(name, x, y, label, handler)
	v, err := ui.gui.SetView(button.name, button.x, button.y, button.x + button.w, button.y+2)

	if err != gocui.ErrUnknownView {
		return nil, err
	}
	if err := ui.writeContent(name, button.label); err != nil {
		return nil, err
	}
	if err := ui.gui.SetKeybinding(button.name, gocui.KeyEnter, gocui.ModNone, button.handler); err != nil {
		return nil, err
	}
	return v, nil
}

// nextElement activate the next element inside the modal view.
func (ui *UI) nextElement(views []string) error {
	var index int
	index = ui.nextItem + 1
	if index > len(views) - 1 {
		index = 0
	}
	ui.nextItem = index % len(views)
	v, err := ui.gui.SetCurrentView(views[ui.nextItem])
	if err != nil {
		if err == gocui.ErrUnknownView {
			return nil
		}
		return err
	}
	if ui.nextItem != 0 {
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		ui.gui.Cursor = false
	}

	ui.nextItem = index
	return nil
}

// prevElement activate the previous element inside the modal view.
func (ui *UI) prevElement(views []string) error {
	var index int
	index = ui.nextItem - 1
	if index < 0 {
		index = len(views) - 1

	}
	ui.nextItem = index % len(views)
	v, err := ui.gui.SetCurrentView(views[ui.nextItem])
	if err != nil {
		if err == gocui.ErrUnknownView {
			return nil
		}
		return err
	}
	if ui.nextItem != 0 {
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		ui.gui.Cursor = false
	}

	ui.nextItem = index
	return nil
}