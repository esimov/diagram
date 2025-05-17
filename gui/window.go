package gui

import (
	"fmt"
	"image"
	"image/color"
	"os"

	"gioui.org/app"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
)

const title = "Diagram preview..."

const (
	maxWindowWidth  = 1024
	maxWindowHeight = 768
)

type GUI struct {
	image  paint.ImageOp
	window *app.Window
}

func NewGUI() *GUI {
	return &GUI{
		window: new(app.Window),
	}
}

func (gui *GUI) Draw(img image.Image) error {
	gui.image = paint.NewImageOp(img)

	var windowWidth, windowHeight float32
	imgWidth, imgHeight := img.Bounds().Dx(), img.Bounds().Dy()

	aspectRatio := float32(imgWidth) / float32(imgHeight)

	if aspectRatio > 1 {
		windowWidth = min(maxWindowWidth, float32(imgWidth))
		windowHeight = windowWidth / aspectRatio
	} else {
		windowHeight = min(maxWindowHeight, float32(imgHeight))
		windowWidth = windowHeight * aspectRatio
	}

	windowWidth = max(windowWidth, maxWindowWidth)
	windowHeight = max(windowHeight, maxWindowHeight)

	// Swap the GUI window width & height in case the image height is greater than its width.
	if imgHeight > imgWidth {
		tmpWindowWidth := windowWidth
		windowWidth = windowHeight
		windowHeight = tmpWindowWidth
	}

	gui.window.Option(
		app.Size(
			unit.Dp(windowWidth),
			unit.Dp(windowHeight),
		),
		app.MaxSize(unit.Dp(windowWidth), unit.Dp(windowHeight)),
		app.Title(title),
	)

	// Center the window on the screen.
	gui.window.Perform(system.ActionCenter)
	// Bring this window to the top of all open windows.
	gui.window.Perform(system.ActionRaise)

	if err := gui.run(gui.window); err != nil {
		defer func() {
			os.Exit(0)
		}()

		return fmt.Errorf("GUI rendering error: %w", err)
	}

	return nil
}

func (gui *GUI) run(w *app.Window) error {
	var ops op.Ops

	for {
		switch e := w.Event().(type) {
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			for {
				event, ok := gtx.Event(key.Filter{
					Name: key.NameEscape,
				})
				if !ok {
					break
				}
				switch event := event.(type) {
				case key.Event:
					switch event.Name {
					case key.NameEscape:
						w.Perform(system.ActionClose)
					}
				}
			}
			gui.drawDiagram(gtx)
			e.Frame(gtx.Ops)
		case app.DestroyEvent:
			return e.Err
		}
	}
}

func (gui *GUI) drawDiagram(gtx layout.Context) {
	layout.Stack{}.Layout(gtx,
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			paint.FillShape(gtx.Ops, color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xff},
				clip.Rect{Max: gtx.Constraints.Max}.Op(),
			)

			return layout.UniformInset(unit.Dp(0)).Layout(gtx,
				func(gtx layout.Context) layout.Dimensions {
					widget.Image{
						Src:   gui.image,
						Scale: 1 / float32(gtx.Dp(unit.Dp(2))),
						Fit:   widget.Unscaled,
					}.Layout(gtx)

					return layout.Dimensions{Size: gtx.Constraints.Max}
				})
		}),
	)
}
