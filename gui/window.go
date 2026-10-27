package gui

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"sync"
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
	maxWindowHeight = 720

	minScaleFactor  = 0.5
	maxScaleFactor  = 3.5
	zoomScaleFactor = 1.2
	outerPadding    = 2
	scaleFactor     = 0.1
	zoomFactor      = 0.4

	inf = 1e6
)

var (
	windowWidth  float32
	windowHeight float32
	zoomPanelDim float32
	zoomPanelImg paint.ImageOp
)

type GUI struct {
	Image  paint.ImageOp
	Window *app.Window
	pan    *Animation
	scroll *Animation

	initZoom sync.Once
}

type diagram struct {
	*scrollTracker
	*mouseTracker
	imgPos   f32.Point
	imgScale float32
}

type scrollTracker struct {
	isScrolling bool
	scrollX     float64
	deltaY      float32
	scroll      gesture.Scroll
}

type mouseTracker struct {
	isDragging        bool
	posX              float32
	posY              float32
	moveX             float32
	moveY             float32
	imgOffsetX        float32
	imgOffsetY        float32
	currentImgOffsetX float32
	currentImgOffsetY float32
	timeStamp         time.Time
}

func NewGUI() *GUI {
	return &GUI{
		Window: new(app.Window),
		pan:    &Animation{Duration: 400 * time.Millisecond},
		scroll: &Animation{Duration: 400 * time.Millisecond},

		initZoom: sync.Once{},
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

	// Bring this window on top of all the opened windows.
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

	// Initialize diagram.
	diagram := &diagram{
		scrollTracker: &scrollTracker{
			isScrolling: false,
			deltaY:      1.0,
			scroll:      gesture.Scroll{},
		},
		mouseTracker: &mouseTracker{},
	}

	scroll, mouse := diagram.scrollTracker, diagram.mouseTracker

	var mpx, mpy float32
	for {
		switch ev := w.Event().(type) {
		case app.FrameEvent:
			gtx := app.NewContext(&ops, ev)
			for {
				// Register for pointer move events over the entire window.
				r := image.Rectangle{Max: image.Point{X: gtx.Constraints.Max.X, Y: gtx.Constraints.Max.Y}}
				area := clip.Rect(r).Push(&ops)
				pointer.CursorPointer.Add(gtx.Ops)
				event.Op(&ops, scroll)
				area.Pop()
				rangeMin, rangeMax := int(-inf), int(inf)

				event, ok := gtx.Event(
					key.Filter{
						Name: key.NameEscape,
					},
					pointer.Filter{
						Target:  scroll,
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
					case pointer.Move:
						mouse.moveX = ev.Position.X
						mouse.moveY = ev.Position.Y
					case pointer.Press:
						mouse.posX = ev.Position.X - mouse.imgOffsetX
						mouse.posY = ev.Position.Y - mouse.imgOffsetY
					case pointer.Drag:
						mouse.imgOffsetX = ev.Position.X - mouse.posX
						mouse.imgOffsetY = ev.Position.Y - mouse.posY

						pointer.CursorGrabbing.Add(gtx.Ops)
						mouse.isDragging = true
					case pointer.Release:
						mouse.currentImgOffsetX = mouse.imgOffsetX
						mouse.currentImgOffsetY = mouse.imgOffsetY

						gui.pan.Delta = time.Since(gui.pan.StartTime)
						mouse.timeStamp = time.Now()
						mouse.isDragging = false
					case pointer.Scroll:
						scroll.isScrolling = true
						scroll.scroll.Add(&ops)

						scroll.deltaY += ev.Scroll.Y * 0.002
						dy := float32(gtx.Dp(unit.Dp(scroll.deltaY))) * 0.01

						if scroll.deltaY > dy {
							scroll.deltaY += dy
						} else {
							scroll.deltaY -= dy
						}

						scroll.scroll.Update(gtx.Metric, gtx.Source, gtx.Now, gesture.Vertical,
							pointer.ScrollRange{Min: rangeMin, Max: rangeMax},
							pointer.ScrollRange{Min: rangeMin, Max: rangeMax})

						if scroll.deltaY < minScaleFactor {
							scroll.deltaY = minScaleFactor
						} else if scroll.deltaY > maxScaleFactor {
							scroll.deltaY = maxScaleFactor
						}

						mpx = mouse.moveX * 0.3
						mpy = mouse.moveY * 0.3

						gui.scroll.Delta = time.Since(gui.scroll.StartTime)
						gui.scroll.Duration = 700 * time.Millisecond
					}
				}
			}
			var scrollEase float64
			diagram.scrollX = gui.scroll.Update(gtx)

			if !scroll.isScrolling {
				scrollEase = gui.scroll.Animate(EaseInOutBack, float64(diagram.scrollX))
			} else {
				scrollEase = 1 + (0.2 * gui.scroll.Animate(EaseInOutSine, float64(diagram.scrollX)))
			}

			var offsetX, offsetY float32
			if !mouse.isDragging {
				panEase := 1 + (0.005 * gui.pan.Animate(EaseInOut, float64(diagram.scrollX)))

				duration := time.Since(mouse.timeStamp).Seconds()
				if duration < 0.2 {
					mouse.currentImgOffsetX *= 0.995 * float32(panEase)
					mouse.currentImgOffsetY *= 0.995 * float32(panEase)
				}
				offsetX = mouse.currentImgOffsetX
				offsetY = mouse.currentImgOffsetY
			} else {
				offsetX = mouse.imgOffsetX
				offsetY = mouse.imgOffsetY
			}

			if mpx < windowWidth/2 {
				offsetX += mpx
			} else {
				offsetX -= mpx
			}

			if mpy < windowHeight/2 {
				offsetY += mpy
			} else {
				offsetY -= mpy
			}

			gui.scroll.StartTime = time.Now()
			gui.pan.StartTime = time.Now()

			diagram.imgScale = scroll.deltaY * float32(scrollEase)
			diagram.imgPos = f32.Pt(offsetX, offsetY)

			gui.drawDiagram(gtx, diagram)
			ev.Frame(gtx.Ops)
		case app.DestroyEvent:
			return ev.Err
		}
	}
}

func (gui *GUI) drawDiagram(gtx layout.Context, d *diagram) {
	gtx.Execute(op.InvalidateCmd{})
	tr := f32.Affine2D{}
	start := time.Time{}

	layout.Stack{}.Layout(gtx,
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			paint.FillShape(gtx.Ops, color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xff},
				clip.Rect{Max: gtx.Constraints.Max}.Op(),
			)

			return layout.UniformInset(unit.Dp(outerPadding)).Layout(gtx,
				func(gtx layout.Context) layout.Dimensions {
					// Offset the image origins.
					offStack := op.Affine(tr.Offset(d.imgPos).
						Scale(
							f32.Pt(windowWidth/2, windowHeight/2),
							f32.Pt(d.imgScale, d.imgScale),
						)).Push(gtx.Ops)

					widget.Image{
						Src:      gui.Image,
						Position: layout.Center,
						Fit:      widget.ScaleDown,
					}.Layout(gtx)
					offStack.Pop()

					// Start the timer when we stopped zooming.
					if d.scrollX == 1 {
						start = time.Now()
					}

					// Hide the navigator window after a short delay once we stopped zooming
					// or when the zoomed diagram is less then a certain percentage.
					if time.Since(start).Milliseconds() == 0 || d.imgScale < zoomScaleFactor {
						if !d.mouseTracker.isDragging {
							return layout.Dimensions{}
						}
					}

					// Initialize the navigator window.
					gui.initZoom.Do(func() {
						zoomPanelDim = d.imgScale * scaleFactor
						zoomPanelImg = gui.Image
					})

					// Show the navigator window.
					gui.showNavigatorWindow(gtx, d)

					return layout.Dimensions{Size: gtx.Constraints.Max}
				})
		}),
	)
}

func (gui *GUI) showNavigatorWindow(gtx layout.Context, d *diagram) {
	tr := f32.Affine2D{}

	defer op.Offset(image.Point{X: 0, Y: 0}).Push(gtx.Ops).Pop()
	layout.Stack{
		Alignment: layout.NW,
	}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(0)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return widget.Border{
					Color: color.NRGBA{R: 0x6c, G: 0x75, B: 0x7d, A: 120},
					Width: unit.Dp(0.5),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(unit.Dp(5)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return widget.Image{
							Src:      zoomPanelImg,
							Scale:    zoomPanelDim,
							Position: layout.Center,
							Fit:      widget.Unscaled,
						}.Layout(gtx)
					})
				})
			})
		}),

		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(0)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				centerX := windowWidth / 2 * zoomPanelDim
				centerY := windowHeight / 2 * zoomPanelDim

				imgSizeX := float32(gui.Image.Size().X) * zoomPanelDim
				imgSizeY := float32(gui.Image.Size().Y) * zoomPanelDim

				zoomDimX := imgSizeX - (imgSizeX * zoomPanelDim)
				zoomDimY := imgSizeY - (imgSizeY * zoomPanelDim)

				px := (zoomDimX + d.imgPos.X) / 2 * zoomFactor
				py := (zoomDimY + d.imgPos.Y) / 2 * zoomFactor

				dx, dy := centerX-px, centerY-py
				if dx < 1 {
					dx = 1
				} else if dx > zoomDimX {
					dx = zoomDimX
				}

				if dy < 1 {
					dy = 1
				} else if dy > zoomDimY {
					dy = zoomDimY
				}

				offStack := op.Affine(tr.Offset(
					f32.Point{X: dx, Y: dy}).
					Scale(
						f32.Pt(dx, dy),
						f32.Pt(1/d.imgScale*0.4, 1/d.imgScale*0.4),
					)).Push(gtx.Ops)

				paint.FillShape(gtx.Ops, color.NRGBA{R: 0xff, A: 90},
					clip.UniformRRect(image.Rectangle{
						Max: image.Point{X: 100, Y: 100},
					}, gtx.Dp(0)).Op(gtx.Ops))

				paint.FillShape(gtx.Ops, color.NRGBA{R: 0xff, A: 0xff},
					clip.Stroke{
						Path: clip.Rect{
							Max: image.Point{X: 100, Y: 100},
						}.Path(),
						Width: 2.0,
					}.Op(),
				)
				offStack.Pop()

				return layout.Dimensions{Size: gtx.Constraints.Max}
			})
		}),
	)
}
