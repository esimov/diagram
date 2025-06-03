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
	maxWindowHeight = 920

	minScaleFactor  = 0.5
	maxScaleFactor  = 3.5
	zoomScaleFactor = 1.1
	outerPadding    = 2
	scaleFactor     = 0.1
	zoomFactor      = 0.5

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
		pan:    &Animation{Duration: 400 * time.Millisecond},
		scroll: &Animation{Duration: 800 * time.Millisecond},

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

						gui.pan.Delta = time.Since(gui.pan.StartTime)
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

						gui.scroll.Delta = time.Since(gui.scroll.StartTime)
						gui.scroll.Duration = 600 * time.Millisecond
					}
				}
			}
			var scrollEase float64
			sx := gui.scroll.Update(gtx)
			if !t.isScrolling {
				scrollEase = gui.scroll.Animate(EaseInOutBack, float64(sx))
			} else {
				scrollEase = 1 + (0.2 * gui.scroll.Animate(EaseInOutSine, float64(sx)))
			}

			var offsetX, offsetY float32
			if !m.isDragging {
				sx = gui.pan.Update(gtx)
				panEase := 1 + (0.005 * gui.pan.Animate(EaseInOut, float64(sx)))

				duration := time.Since(m.timeStamp).Seconds()
				if duration < 0.2 {
					m.currentImgOffsetX *= 0.995 * float32(panEase)
					m.currentImgOffsetY *= 0.995 * float32(panEase)
				}
				offsetX = m.currentImgOffsetX
				offsetY = m.currentImgOffsetY
			} else {
				offsetX = m.imgOffsetX
				offsetY = m.imgOffsetY
			}

			gui.scroll.StartTime = time.Now()
			gui.pan.StartTime = time.Now()

			imgScale := t.deltaY * float32(scrollEase)
			imgPos := f32.Pt(offsetX, offsetY)

			gui.drawDiagram(gtx, imgScale, imgPos)
			ev.Frame(gtx.Ops)
		case app.DestroyEvent:
			return ev.Err
		}
	}
}

func (gui *GUI) drawDiagram(gtx layout.Context, imgScale float32, imgPos f32.Point) {
	gtx.Execute(op.InvalidateCmd{})
	tr := f32.Affine2D{}

	layout.Stack{}.Layout(gtx,
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			paint.FillShape(gtx.Ops, color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xff},
				clip.Rect{Max: gtx.Constraints.Max}.Op(),
			)

			return layout.UniformInset(unit.Dp(outerPadding)).Layout(gtx,
				func(gtx layout.Context) layout.Dimensions {
					// Offset the image origins.
					offStack := op.Affine(tr.Offset(imgPos).Scale(imgPos, f32.Pt(imgScale, imgScale))).Push(gtx.Ops)
					widget.Image{
						Src:      gui.Image,
						Position: layout.NW,
						Fit:      widget.ScaleDown,
					}.Layout(gtx)
					offStack.Pop()

					if imgScale < zoomScaleFactor {
						return layout.Dimensions{}
					}

					gui.initZoom.Do(func() {
						zoomPanelDim = imgScale * scaleFactor
						zoomPanelImg = gui.Image
					})

					// Zoom navigator area.
					zoomPanelWidth := windowWidth * scaleFactor
					zoomPanelHeight := windowHeight * scaleFactor

					defer op.Offset(image.Point{X: 0, Y: 0}).Push(gtx.Ops).Pop()
					layout.Stack{
						Alignment: layout.NW,
					}.Layout(gtx,
						layout.Expanded(func(gtx layout.Context) layout.Dimensions {
							return layout.UniformInset(unit.Dp(2)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return widget.Border{
									Color: color.NRGBA{R: 0x6c, G: 0x75, B: 0x7d, A: 0xff},
									Width: unit.Dp(2),
								}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									return layout.UniformInset(unit.Dp(0)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										return widget.Image{
											Src:      zoomPanelImg,
											Scale:    zoomPanelDim / gtx.Metric.PxPerDp,
											Position: layout.NW,
											Fit:      widget.Unscaled,
										}.Layout(gtx)
									})

								})
							})
						}),
					)

					layout.Stack{}.Layout(gtx,
						layout.Expanded(func(gtx layout.Context) layout.Dimensions {
							defer op.Offset(image.Point{X: 0, Y: 10}).Push(gtx.Ops).Pop()

							return layout.UniformInset(unit.Dp(0)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								offset := f32.Point{
									X: zoomPanelWidth - (imgPos.X * scaleFactor / 2),
									Y: -(imgPos.Y * scaleFactor / 2),
								}

								zoomStack := op.Affine(
									tr.Offset(offset.Mul(imgScale)).Scale(
										f32.Pt(0, 0),
										f32.Pt(1/imgScale*zoomFactor, 1/imgScale*zoomFactor),
									)).Push(gtx.Ops)

								paint.FillShape(gtx.Ops, color.NRGBA{R: 0xff, A: 40},
									clip.UniformRRect(image.Rectangle{
										Max: image.Point{
											X: int(zoomPanelWidth),
											Y: int(zoomPanelHeight),
										},
									}, 0).Op(gtx.Ops),
								)

								paint.FillShape(gtx.Ops, color.NRGBA{R: 0xff, A: 0xff},
									clip.Stroke{
										Path: clip.Rect{Max: image.Point{
											X: int(zoomPanelWidth),
											Y: int(zoomPanelHeight),
										}}.Path(),
										Width: 2.0,
									}.Op(),
								)
								zoomStack.Pop()

								return layout.Dimensions{}
							})
						}),
					)

					return layout.Dimensions{Size: gtx.Constraints.Max}
				})
		}),
	)
}
