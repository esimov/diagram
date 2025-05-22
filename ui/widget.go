package ui

import (
	"fmt"
	"strings"

	"github.com/jroimartin/gocui"
)

type Widget struct {
	name       string
	groupName  string
	posX, posY int
	width      int
	handlerFn  func(*gocui.Gui, *gocui.View) error
	*UI
}

type WidgetEmbedder interface {
	GetWidget() *Widget
}

type HandlerFn func(g *gocui.Gui, v *gocui.View) error
type WidgetOption[T WidgetEmbedder] func(T) error

var _ ComponentHandler = (*Widget)(nil)

// New creates a new widget.
func NewWidget[T WidgetEmbedder](w T, options ...WidgetOption[T]) (*T, error) {
	for _, opt := range options {
		if err := opt(w); err != nil {
			return nil, fmt.Errorf("widget option error: %w", err)
		}
	}

	return &w, nil
}

// WithDefaultWidgetOptions sets the default widget options for an already created widget element.
func WithDefaultWidgetOptions[T WidgetEmbedder](name string, posX, posY int) WidgetOption[T] {
	return func(w T) error {
		w.GetWidget().setDefaultWidgetOptions(name, posX, posY)
		return nil
	}
}

// WithWidgetWidth sets the widget element width.
func WithWidgetWidth[T WidgetEmbedder](width int) WidgetOption[T] {
	return func(w T) error {
		w.GetWidget().setWidth(width)
		return nil
	}
}

// WithHandlerFn assigns a handler to the widget.
func WithHandlerFn[T WidgetEmbedder](handlerFn HandlerFn) WidgetOption[T] {
	return func(w T) error {
		w.GetWidget().setHandlerFn(handlerFn)
		return nil
	}
}

// WithUIHandler assigns an already initialized UI struct to the widget.
func WithUIHandler[T WidgetEmbedder](ui *UI) WidgetOption[T] {
	return func(w T) error {
		if ui == nil {
			return fmt.Errorf("UI not initialized")
		}
		w.GetWidget().setUI(ui)
		return nil
	}
}

// GetWidget implements the interface method definition.
func (w *Widget) GetWidget() *Widget {
	return w
}

func (w *Widget) setDefaultWidgetOptions(name string, posX, posY int) {
	w.name = name
	w.posX = posX
	w.posY = posY
}
func (w *Widget) setWidth(width int) { w.width = width }
func (w *Widget) setUI(ui *UI)       { w.UI = ui }
func (w *Widget) setHandlerFn(handlerFn HandlerFn) {
	w.handlerFn = handlerFn
}

// Draw draws the widget element.
func (w *Widget) Draw() (*gocui.View, error) {
	v, err := w.gui.SetView(w.name, w.posX, w.posY, w.posX+w.width, w.posY+2)
	if err != gocui.ErrUnknownView {
		return nil, err
	}
	if err := w.writeContent(w.name, strings.ToUpper(w.name)); err != nil {
		return nil, err
	}
	if err := w.gui.SetKeybinding(w.name, gocui.KeyEnter, gocui.ModNone, w.handlerFn); err != nil {
		return nil, err
	}

	w.widgetItems[w.GetWidget().groupName] = append(w.widgetItems[w.GetWidget().groupName], v.Name())

	return v, nil
}

// NextElement activate the next element inside the modal view.
func (w *Widget) NextElement(g *gocui.Gui, v *gocui.View) error {
	w.unfocus()
	w.activeModalView = (w.activeModalView + 1) % len(w.widgetItems[w.GetWidget().groupName])
	w.focus()

	return nil
}

// PrevElement activate the previous element inside the modal view.
func (w *Widget) PrevElement(g *gocui.Gui, v *gocui.View) error {
	w.unfocus()

	if w.activeModalView-1 < 0 {
		w.activeModalView = len(w.widgetItems[w.GetWidget().groupName]) - 1
	} else {
		w.activeModalView = (w.activeModalView - 1) % len(w.widgetItems[w.GetWidget().groupName])
	}
	w.focus()

	return nil
}

func (w *Widget) focus() {
	if len(w.widgetItems[w.GetWidget().groupName]) != 0 {
		v, _ := w.gui.SetCurrentView(w.widgetItems[w.GetWidget().groupName][w.activeModalView])
		v.Highlight = true
		v.SelFgColor = gocui.ColorWhite
		v.SelBgColor = gocui.ColorBlack
		w.gui.Cursor = false
	}
}

func (w *Widget) unfocus() {
	if len(w.widgetItems[w.GetWidget().groupName]) != 0 {
		v, _ := w.gui.SetCurrentView(w.widgetItems[w.GetWidget().groupName][w.activeModalView])
		v.Highlight = false
		v.SelBgColor = gocui.Attribute(w.activeLayoutColor)
	}
}
