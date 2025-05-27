package gui

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
)

const title = "Preview diagram..."

const (
	maxWindowWidth  = 1024
	maxWindowHeight = 920

	minScaleFactor = 0.3
	maxScaleFactor = 3.5
	inf            = 1e6
)

type GUI struct {
	Image  paint.ImageOp
	Window *app.Window
}

type scrollTracker struct {
	*Animation
	isScrolling bool
	deltaY      float32
	scroll      gesture.Scroll
}

func NewGUI() *GUI {
	return &GUI{
		Window: new(app.Window),
	}
}

func (gui *GUI) Draw(img image.Image) error {
	gui.Image = paint.NewImageOp(img)

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

	gui.Window.Option(
		app.Size(
			unit.Dp(windowWidth),
			unit.Dp(windowHeight),
		),
		app.MaxSize(unit.Dp(windowWidth), unit.Dp(windowHeight)),
		app.Title(title),
	)

	// Center the window on the screen.
	gui.Window.Perform(system.ActionCenter)
	// Bring this window to the top of all open windows.
	gui.Window.Perform(system.ActionRaise)

	if err := gui.run(gui.Window); err != nil {
		defer func() {
			os.Exit(0)
		}()

		return fmt.Errorf("GUI rendering error: %w", err)
	}

	return nil
}

func (gui *GUI) run(w *app.Window) error {
	var ops op.Ops

	var deltaY float32 = 1.0

	// Initialize the mouse position.
	scrollTracker := &scrollTracker{
		&Animation{Duration: 800 * time.Millisecond},
		false,
		deltaY,
		gesture.Scroll{},
	}

	for {
		switch ev := w.Event().(type) {
		case app.FrameEvent:
			gtx := app.NewContext(&ops, ev)
			for {
				// Register for pointer move events over the entire window.
				r := image.Rectangle{Max: image.Point{X: gtx.Constraints.Max.X, Y: gtx.Constraints.Max.Y}}
				area := clip.Rect(r).Push(&ops)
				event.Op(&ops, scrollTracker)
				area.Pop()
				rangeMin, rangeMax := int(-inf), int(inf)

				event, ok := gtx.Event(
					key.Filter{
						Name: key.NameEscape,
					},
					pointer.Filter{
						Target:  scrollTracker,
						ScrollY: pointer.ScrollRange{Min: rangeMin, Max: rangeMax},
						Kinds:   pointer.Scroll | pointer.Press | pointer.Release | pointer.Move,
					})
				if !ok {
					break
				}

				switch ev := event.(type) {
				case key.Event:
					switch ev.Name {
					case key.NameEscape:
						w.Perform(system.ActionClose)
					}
				case pointer.Event:
					switch ev.Kind {
					case pointer.Press:
						switch ev.Modifiers {
						case key.ModShortcut:
							log.Println("Mouse pressed")
						case key.ModShift:
							// Gio swaps X and Y when Shift is pressed.
							fallthrough
						default:

						}
					case pointer.Release:
						// log.Println("Mouse released")
					case pointer.Scroll:
						scrollTracker.isScrolling = true
						scrollTracker.scroll.Add(&ops)

						scrollTracker.deltaY += ev.Scroll.Y * 0.002
						dy := float32(gtx.Dp(unit.Dp(scrollTracker.deltaY))) * 0.02

						if scrollTracker.deltaY < dy {
							scrollTracker.deltaY += dy
						} else {
							scrollTracker.deltaY -= dy
						}
						scrollTracker.scroll.Update(gtx.Metric, gtx.Source, gtx.Now, gesture.Vertical,
							pointer.ScrollRange{Min: rangeMin, Max: rangeMax},
							pointer.ScrollRange{Min: rangeMin, Max: rangeMax})

						if scrollTracker.deltaY < minScaleFactor {
							scrollTracker.deltaY = minScaleFactor
						} else if scrollTracker.deltaY > maxScaleFactor {
							scrollTracker.deltaY = maxScaleFactor
						}
						scrollTracker.Delta = time.Since(scrollTracker.StartTime)
						scrollTracker.Duration = 500 * time.Millisecond
					}
				}
			}
			var t float64
			d := scrollTracker.Update(gtx)
			if !scrollTracker.isScrolling {
				t = scrollTracker.Animate(EaseInOutBack, float64(d))
			} else {
				t = 1 + (0.05 * scrollTracker.Animate(EaseInOutSine, float64(d)))
			}

			scrollTracker.StartTime = time.Now()
			gui.drawDiagram(gtx, scrollTracker.deltaY*float32(t))
			ev.Frame(gtx.Ops)
		case app.DestroyEvent:
			return ev.Err
		}
	}
}

func (gui *GUI) drawDiagram(gtx layout.Context, scale float32) {
	gtx.Execute(op.InvalidateCmd{})

	layout.Stack{}.Layout(gtx,
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			paint.FillShape(gtx.Ops, color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xff},
				clip.Rect{Max: gtx.Constraints.Max}.Op(),
			)

			return layout.Inset(layout.Inset{
				Top:  0,
				Left: 0,
			}).Layout(gtx,
				func(gtx layout.Context) layout.Dimensions {
					widget.Image{
						Src:      gui.Image,
						Scale:    scale,
						Position: layout.NW,
						Fit:      widget.Unscaled,
					}.Layout(gtx)

					return layout.Dimensions{Size: gtx.Constraints.Max}
				})
		}),
	)
}
