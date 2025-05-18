package ui

import (
	"fmt"
	"log"

	"github.com/jroimartin/gocui"
)

type RadioBtnWidget struct {
	Widget
	options  []*option
	handlers Handlers
	*position
}

type position struct {
	x, y int
	w, h int
}

type option struct {
	name      string
	unCheck   string
	checked   string
	isChecked bool
	*position
}

const (
	uncheckRadioButton = "\u25ef"
	checkedRadioButton = "\u25c9"
)

var currentColor gocui.Attribute

// NewRadioButton creates a new radio button widget.
func NewRadioButton[T WidgetEmbedder](ui *UI, name string, posX, posY int) (*RadioBtnWidget, error) {
	radioBtnWidget, err := NewWidget(
		&RadioBtnWidget{
			handlers: make(Handlers),
			position: &position{
				x: posX,
				y: posY,
				w: posX,
				h: posY,
			},
		},
		[]WidgetOption[*RadioBtnWidget]{
			WithDefaultWidgetOptions[*RadioBtnWidget](name, posX, posY),
			WithUIHandler[*RadioBtnWidget](ui),
		}...,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create radio button widget: %w", err)
	}

	radioBtn := *radioBtnWidget
	radioBtn.AddHandler(gocui.KeyTab, radioBtn.nextRadio).
		AddHandler(gocui.KeyArrowRight, radioBtn.nextRadio).
		AddHandler(gocui.KeyArrowLeft, radioBtn.prevRadio).
		AddHandler(gocui.KeyEnter, radioBtn.closeModal).
		AddHandler(gocui.KeySpace, radioBtn.closeModal)

	return radioBtn, nil
}

// Draw draws the radio button
func (w *RadioBtnWidget) Draw() *RadioBtnWidget {
	for i, opt := range w.options {
		if v, err := w.gui.SetView(opt.name, opt.x, opt.y, opt.w, opt.h); err != nil {
			if err != gocui.ErrUnknownView {
				log.Fatalf("error creating a view: %v", err)
			}
			v.Frame = false
			v.BgColor = gocui.ColorBlack
			v.SelBgColor = gocui.ColorBlack
			fmt.Fprint(v, opt.unCheck)

			if w.handlers != nil {
				for key, handler := range w.handlers {
					if err := w.gui.SetKeybinding(opt.name, key, gocui.ModNone, handler); err != nil {
						log.Fatalf("error set key bindings for radio buttons: %v", err)
					}
				}
			}
			if i == w.activeLayout {
				w.focus()
				w.check(v)
			}
		}
	}

	return w
}

// AddOptions add options
func (w *RadioBtnWidget) AddOptions(names ...string) *RadioBtnWidget {
	for _, name := range names {
		w.AddOption(name)
	}

	return w
}

// AddOption add radio button option.
func (w *RadioBtnWidget) AddOption(name string) *RadioBtnWidget {
	p := w.position
	optLen := len(w.options)

	if optLen != 0 {
		p = w.options[optLen-1].position
	}
	opt := newOption(name, p.w, p.y)
	w.options = append(w.options, opt)

	return w
}

// newOption creates a new radio button option.
func newOption(name string, x, y int) *option {
	return &option{
		name:      name,
		unCheck:   fmt.Sprintf("%s  %s", uncheckRadioButton, name),
		checked:   fmt.Sprintf("%s  %s", checkedRadioButton, name),
		isChecked: false,
		position: &position{
			x: x,
			y: y,
			w: x + len(name) + 4,
			h: y + 2,
		},
	}
}

// AddHandler assigns a handler function to the key.
func (w *RadioBtnWidget) AddHandler(key Key, handler HandlerFn) *RadioBtnWidget {
	w.handlers[key] = handler
	return w
}

// nextRadio jumps to the next radio button
func (w *RadioBtnWidget) nextRadio(g *gocui.Gui, v *gocui.View) error {
	w.unFocus()
	w.activeLayout = (w.activeLayout + 1) % len(w.options)
	w.focus()

	return nil
}

// prevRadio jumps to the previous radio button
func (w *RadioBtnWidget) prevRadio(g *gocui.Gui, v *gocui.View) error {
	w.unFocus()

	if w.activeLayout-1 < 0 {
		w.activeLayout = len(w.options) - 1
	} else {
		w.activeLayout = (w.activeLayout - 1) % len(w.options)
	}
	w.focus()

	return nil
}

// focus brings into focus the active radio button
func (w *RadioBtnWidget) focus() {
	if len(w.options) != 0 {
		v, _ := w.gui.SetCurrentView(w.options[w.activeLayout].name)

		switch v.Name() {
		case defaultLayout.ToString():
			currentColor = gocui.ColorDefault
		case blackLayout.ToString():
			currentColor = gocui.ColorBlack
		case blueLayout.ToString():
			currentColor = gocui.ColorBlue
		case magentaLayout.ToString():
			currentColor = gocui.ColorMagenta
		case cyanLayout.ToString():
			currentColor = gocui.ColorCyan
		case greenLayout.ToString():
			currentColor = gocui.ColorGreen
		default:
			currentColor = gocui.ColorDefault
		}
		_ = w.check(v)
		w.ApplyLayoutColor(currentColor)

		v.Highlight = true
		w.gui.Cursor = false
	}
}

// unFocus radio button
func (w *RadioBtnWidget) unFocus() {
	if len(w.options) != 0 {
		v, _ := w.gui.SetCurrentView(w.options[w.activeLayout].name)
		v.Highlight = false
	}
}

// check activates the selected radio button
func (w *RadioBtnWidget) check(v *gocui.View) error {
	for _, opt := range w.options {
		if v, err := w.gui.View(opt.name); err == nil {
			v.Clear()
			fmt.Fprint(v, opt.unCheck)
		}
		opt.isChecked = false
	}

	w.options[w.activeLayout].isChecked = true
	v.Clear()
	fmt.Fprint(v, w.options[w.activeLayout].checked)

	return nil
}

// closeModal closes all the view elements of the opened modal
func (w *RadioBtnWidget) closeModal(g *gocui.Gui, v *gocui.View) error {
	w.selectedColor = currentColor
	if err := w.closeModals(layoutModalViews...); err != nil {
		return err
	}
	return nil
}
