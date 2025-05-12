package ui

import (
	"fmt"
	"strings"

	"github.com/jroimartin/gocui"
)

type buttonWidget struct {
	name      string
	x, y      int
	width     int
	handlerFn func(*gocui.Gui, *gocui.View) error
	*UI
}

type handlerFn func(g *gocui.Gui, v *gocui.View) error
type WidgetOption func(*buttonWidget) error

var _ WidgetHandler = (*buttonWidget)(nil)

// NewButton creates a new button widget.
func NewButton(name string, x, y, w int, options ...WidgetOption) (*buttonWidget, error) {
	widget := &buttonWidget{
		name:  name,
		x:     x,
		y:     y,
		width: w,
	}

	for _, opt := range options {
		if err := opt(widget); err != nil {
			return nil, fmt.Errorf("button widget option error: %w", err)
		}
	}

	return widget, nil
}

func WithHandlerFn(handlerFn *handlerFn) WidgetOption {
	return func(w *buttonWidget) error {
		w.handlerFn = *handlerFn

		return nil
	}
}

func WithGUI(ui *UI) WidgetOption {
	return func(w *buttonWidget) error {
		if ui == nil {
			return fmt.Errorf("UI not initialized")
		}
		w.UI = ui

		return nil
	}
}

func (w *buttonWidget) Draw() (*gocui.View, error) {
	v, err := w.gui.SetView(w.name, w.x, w.y, w.x+w.width, w.y+2)
	if err != gocui.ErrUnknownView {
		return nil, err
	}
	if err := w.writeContent(w.name, strings.ToUpper(w.name)); err != nil {
		return nil, err
	}
	if err := w.gui.SetKeybinding(w.name, gocui.KeyEnter, gocui.ModNone, w.handlerFn); err != nil {
		return nil, err
	}

	return v, nil
}

// NextElement activate the next element inside the modal view.
func (w *buttonWidget) NextElement(views []string) error {
	var index int
	index = w.nextItem + 1
	if index > len(views)-1 {
		index = 0
	}
	w.nextItem = index % len(views)
	v, err := w.gui.SetCurrentView(views[w.nextItem])
	if err != nil {
		if err == gocui.ErrUnknownView {
			return nil
		}
		return err
	}
	if w.nextItem != 0 {
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		w.gui.Cursor = false
	}

	w.nextItem = index
	return nil
}

// PrevElement activate the previous element inside the modal view.
func (w *buttonWidget) PrevElement(views []string) error {
	var index int
	index = w.nextItem - 1
	if index < 0 {
		index = len(views) - 1

	}
	w.nextItem = index % len(views)
	v, err := w.gui.SetCurrentView(views[w.nextItem])
	if err != nil {
		if err == gocui.ErrUnknownView {
			return nil
		}
		return err
	}
	if w.nextItem != 0 {
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		w.gui.Cursor = false
	}

	w.nextItem = index
	return nil
}
