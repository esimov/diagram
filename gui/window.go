package gui

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/f32"
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

	minScaleFactor = 0.5
	maxScaleFactor = 3.5
	inf            = 1e6
)

var (
	windowWidth  float32
	windowHeight float32
)

type GUI struct {
	Image  paint.ImageOp
	Window *app.Window
}

type scrollTracker struct {
	isScrolling bool
	deltaY      float32
	scroll      gesture.Scroll
}

type mouseTracker struct {
	isDragging        bool
	mousePosX         float32
	mousePosY         float32
	imgOffsetX        float32
	imgOffsetY        float32
	currentImgOffsetX float32
	currentImgOffsetY float32
	timeStamp         time.Time
}

func NewGUI() *GUI {
	return &GUI{
		Window: new(app.Window),
	}
}

func (gui *GUI) Draw(img image.Image) error {
	gui.Image = paint.NewImageOp(img)

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

	// Initialize the scroll tracker.
	t := &scrollTracker{
		isScrolling: false,
		deltaY:      deltaY,
		scroll:      gesture.Scroll{},
	}

	// Initialize the mouse tracker.
	m := &mouseTracker{}

	scrollAnimation := &Animation{Duration: 800 * time.Millisecond}
	panAnimation := &Animation{Duration: 400 * time.Millisecond}

	tr := f32.Affine2D{}

	for {
		switch ev := w.Event().(type) {
		case app.FrameEvent:
			gtx := app.NewContext(&ops, ev)
			for {
				// Register for pointer move events over the entire window.
				r := image.Rectangle{Max: image.Point{X: gtx.Constraints.Max.X, Y: gtx.Constraints.Max.Y}}
				area := clip.Rect(r).Push(&ops)
				pointer.CursorPointer.Add(gtx.Ops)
				event.Op(&ops, t)
				area.Pop()
				rangeMin, rangeMax := int(-inf), int(inf)

				event, ok := gtx.Event(
					key.Filter{
						Name: key.NameEscape,
					},
					pointer.Filter{
						Target:  t,
						ScrollY: pointer.ScrollRange{Min: rangeMin, Max: rangeMax},
						Kinds:   pointer.Scroll | pointer.Press | pointer.Release | pointer.Move | pointer.Drag,
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
						m.mousePosX = ev.Position.X - m.imgOffsetX
						m.mousePosY = ev.Position.Y - m.imgOffsetY
					case pointer.Drag:
						m.imgOffsetX = ev.Position.X - m.mousePosX
						m.imgOffsetY = ev.Position.Y - m.mousePosY

						pointer.CursorGrabbing.Add(gtx.Ops)
						m.isDragging = true
					case pointer.Release:
						m.currentImgOffsetX = m.imgOffsetX
						m.currentImgOffsetY = m.imgOffsetY

						panAnimation.Delta = time.Since(panAnimation.StartTime)
						m.timeStamp = time.Now()
						m.isDragging = false
					case pointer.Scroll:
						t.isScrolling = true
						t.scroll.Add(&ops)

						t.deltaY += ev.Scroll.Y * 0.001
						dy := float32(gtx.Dp(unit.Dp(t.deltaY))) * 0.01

						if t.deltaY > dy {
							t.deltaY += dy
						} else {
							t.deltaY -= dy
						}
						t.scroll.Update(gtx.Metric, gtx.Source, gtx.Now, gesture.Vertical,
							pointer.ScrollRange{Min: rangeMin, Max: rangeMax},
							pointer.ScrollRange{Min: rangeMin, Max: rangeMax})

						if t.deltaY < minScaleFactor {
							t.deltaY = minScaleFactor
						} else if t.deltaY > maxScaleFactor {
							t.deltaY = maxScaleFactor
						}

						scrollAnimation.Delta = time.Since(scrollAnimation.StartTime)
						scrollAnimation.Duration = 600 * time.Millisecond
					}
				}
			}

			var scrollEase float64
			sx := scrollAnimation.Update(gtx)
			if !t.isScrolling {
				scrollEase = scrollAnimation.Animate(EaseInOutBack, float64(sx))
			} else {
				scrollEase = 1 + (0.2 * scrollAnimation.Animate(EaseInOutSine, float64(sx)))
			}

			var offsetX, offsetY float32
			if !m.isDragging {
				sx = panAnimation.Update(gtx)
				panEase := 1 + (0.005 * panAnimation.Animate(EaseInOut, float64(sx)))

				duration := time.Since(m.timeStamp).Seconds()
				if duration < 0.3 {
					m.currentImgOffsetX *= 0.99 * float32(panEase)
					m.currentImgOffsetY *= 0.99 * float32(panEase)
				}
				offsetX = m.currentImgOffsetX
				offsetY = m.currentImgOffsetY
			} else {
				offsetX = m.imgOffsetX
				offsetY = m.imgOffsetY
			}

			scrollAnimation.StartTime = time.Now()
			panAnimation.StartTime = time.Now()

			imgScale := t.deltaY * float32(scrollEase)
			imgPos := f32.Pt(offsetX, offsetY)

			// Offset the image origins.
			op.Affine(tr.Offset(imgPos).Scale(imgPos, f32.Pt(imgScale, imgScale))).Add(gtx.Ops)

			gui.drawDiagram(gtx, imgScale)
			ev.Frame(gtx.Ops)
		case app.DestroyEvent:
			return ev.Err
		}
	}
}

func (gui *GUI) drawDiagram(gtx layout.Context, imgScale float32) {
	gtx.Execute(op.InvalidateCmd{})

	layout.Stack{}.Layout(gtx,
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			paint.FillShape(gtx.Ops, color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xff},
				clip.Rect{Max: gtx.Constraints.Max}.Op(),
			)

			return layout.UniformInset(unit.Dp(0)).Layout(gtx,
				func(gtx layout.Context) layout.Dimensions {
					widget.Image{
						Src:      gui.Image,
						Scale:    imgScale,
						Position: layout.NW,
						Fit:      widget.ScaleDown,
					}.Layout(gtx)

					return layout.Dimensions{Size: gtx.Constraints.Max}
				})
		}),
	)
}
