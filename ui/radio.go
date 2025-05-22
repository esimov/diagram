package ui

import (
	"fmt"
	"log"
	"slices"

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

var (
	selectedLayoutColor gocui.Attribute
	currentOption       int
)

// NewRadioButton creates a new radio button widget.
func NewRadioButton[T WidgetEmbedder](ui *UI, modal *gocui.View, viewName string, posX, posY int) (*RadioBtnWidget, error) {
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
			WithDefaultWidgetOptions[*RadioBtnWidget](viewName, posX, posY),
			WithUIHandler[*RadioBtnWidget](ui),
		}...,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create radio button widget: %w", err)
	}

	currentOption = ui.activeLayoutOption

	radioBtn := *radioBtnWidget
	radioBtn.AddHandler(gocui.KeyTab, radioBtn.nextRadio).
		AddHandler(gocui.KeyArrowRight, radioBtn.nextRadio).
		AddHandler(gocui.KeyArrowLeft, radioBtn.prevRadio).
		AddHandler(gocui.KeyEnter, radioBtn.apply).
		AddHandler(gocui.KeySpace, radioBtn.apply).
		AddHandler(gocui.KeyEsc, radioBtn.onClose).
		AddHandler(gocui.MouseRelease, radioBtn.onClick)

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
			v.BgColor = w.activeLayoutColor
			v.SelBgColor = w.activeLayoutColor
			fmt.Fprint(v, opt.unCheck)

			if w.handlers != nil {
				for key, handler := range w.handlers {
					if err := w.gui.SetKeybinding(opt.name, key, gocui.ModNone, handler); err != nil {
						log.Fatalf("error set key bindings for radio buttons: %v", err)
					}
				}
			}
			if i == w.activeLayoutOption {
				w.Focus()
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
	w.Unfocus()
	currentOption = (currentOption + 1) % len(w.options)
	w.Focus()

	return nil
}

// prevRadio jumps to the previous radio button
func (w *RadioBtnWidget) prevRadio(g *gocui.Gui, v *gocui.View) error {
	w.Unfocus()

	if currentOption-1 < 0 {
		currentOption = len(w.options) - 1
	} else {
		currentOption = (currentOption - 1) % len(w.options)
	}
	w.Focus()

	return nil
}

// Focus brings into Focus the active radio button
func (w *RadioBtnWidget) Focus() {
	if len(w.options) != 0 {
		v, _ := w.gui.SetCurrentView(w.options[currentOption].name)

		_ = w.check(v)
		switch v.Name() {
		case defaultLayout.ToString():
			selectedLayoutColor = gocui.ColorDefault
		case blackLayout.ToString():
			selectedLayoutColor = gocui.ColorBlack
		case blueLayout.ToString():
			selectedLayoutColor = gocui.ColorBlue
		case magentaLayout.ToString():
			selectedLayoutColor = gocui.ColorMagenta
		case cyanLayout.ToString():
			selectedLayoutColor = gocui.ColorCyan
		case greenLayout.ToString():
			selectedLayoutColor = gocui.ColorGreen
		default:
			selectedLayoutColor = gocui.ColorDefault
		}
		v.Highlight = true
		w.gui.Cursor = false

		w.applySelectedColor(selectedLayoutColor)
	}
}

// Unfocus radio button
func (w *RadioBtnWidget) Unfocus() {
	if len(w.options) != 0 {
		v, _ := w.gui.SetCurrentView(w.options[currentOption].name)
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

	w.options[currentOption].isChecked = true
	v.Clear()
	fmt.Fprint(v, w.options[currentOption].checked)

	return nil
}

// apply applies the selected choice and closes all the view elements of the opened modal
func (w *RadioBtnWidget) apply(g *gocui.Gui, v *gocui.View) error {
	w.activeLayoutColor = selectedLayoutColor
	w.activeLayoutOption = currentOption
	if err := w.closeModals(layoutModalViews...); err != nil {
		return err
	}
	return nil
}

// onClose closes the layout modal
func (w *RadioBtnWidget) onClose(g *gocui.Gui, v *gocui.View) error {
	if err := w.closeModals(layoutModalViews...); err != nil {
		return err
	}

	w.applySelectedColor(w.activeLayoutColor)
	return nil
}

// onClick activates the selected radio button on click
func (w *RadioBtnWidget) onClick(*gocui.Gui, *gocui.View) error {
	for idx, opt := range layoutOptions {
		v, _ := w.gui.View(opt)
		cx, _ := v.Cursor()

		if cx > 0 {
			v.SetCursor(0, 0)
			w.Unfocus()
			currentOption = idx
			w.Focus()
			continue
		}
	}

	return nil
}

// applySelectedColor applies the selected color to the layout views
func (ui *UI) applySelectedColor(layoutColor gocui.Attribute) {
	views := slices.Concat(mainViews, layoutModalViews, []string{diagramsPanel})
	for _, view := range views {
		if v, err := ui.gui.View(view); v != nil && err != gocui.ErrUnknownView {
			v.BgColor = layoutColor
			v.SelBgColor = layoutColor
			switch layoutColor {
			case gocui.ColorMagenta,
				gocui.ColorCyan:

				v.SelFgColor = gocui.ColorBlack
				if view == diagramsPanel {
					v.SelFgColor = gocui.ColorBlack
					v.SelBgColor = gocui.ColorGreen
				}
			case gocui.ColorBlue,
				gocui.ColorGreen:

				v.SelFgColor = gocui.ColorBlack

				if view == diagramsPanel {
					v.SelFgColor = gocui.ColorWhite
					v.SelBgColor = gocui.ColorBlack
				}
			default:
				v.SelFgColor = gocui.ColorGreen
				if view == diagramsPanel {
					v.SelFgColor = gocui.ColorBlack
					v.SelBgColor = gocui.ColorGreen
				}
			}
		}
	}
	ui.gui.BgColor = layoutColor
}
