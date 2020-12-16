// Package canvas is responsible to convert the ascii symbols to hand drawn diagrams.
// It implements the basic canvas drawing operations like moveTo, lineTo, fillText.

// The Draw method signature declared in the Drawer interface implements the method separately on Line and Text struct.
package canvas

import (
	"math"
	"math/rand"

	"github.com/fogleman/gg"
)

// Canvas defines the canvas basic elements.
type Canvas struct {
	*gg.Context
	font      string
	lineWidth float64
}

// Drawer interface defines the Canvas drawing method.
// Struct needs to implement the Draw method.
type Drawer interface {
	Draw(*Canvas)
}

// CellSize defines symbol's cell size.
const CellSize float64 = 20

// NewCanvas is a constructor method, which instantiates a new Canvas element.
func NewCanvas(ctx *gg.Context, font string, lineWidth float64) *Canvas {
	if err := ctx.LoadFontFace(font, 20); err != nil {
		panic(err)
	}
	ctx.SetLineWidth(lineWidth)
	return &Canvas{ctx, font, lineWidth}
}

var _x0, _y0 float64

// moveTo move the pointer to (x0,y0) position
func (ctx *Canvas) moveTo(x0, y0 float64) {
	_x0 = x0
	_y0 = y0
}

// lineTo move the pointer to (x1,y1) position
func (ctx *Canvas) lineTo(x1, y1 float64) {
	ctx.shakyLine(_x0, _y0, x1, y1)
	ctx.moveTo(x1, y1)
}

// shakyLine draw a shaky line between (x0, y0) and (x1, y1).
func (ctx *Canvas) shakyLine(x0, y0, x1, y1 float64) {
	var dx, dy float64
	var k1, k2, l3, l4, x3, y3, x4, y4 float64
	dx = x1 - x0
	dy = y1 - y0

	l := math.Sqrt(dx*dx + dy*dy)

	// Pick two random points that are placed on different sides of the line that passes through.
	K := math.Sqrt(l) / 1.5
	k1 = rand.Float64()
	k2 = rand.Float64()
	l3 = rand.Float64() * K
	l4 = rand.Float64() * K

	// Pick a random point on the line between P0 and P1.
	x3 = x0 + dx*k1 + dy/l*l3
	y3 = y0 + dy*k1 - dx/l*l3

	// Pick a random point on the line between P0 and P1 but in the opposite direction.
	x4 = x0 + dx*k2 - dy/l*l4
	y4 = y0 + dy*k2 + dx/l*l4

	// Draw a bezier curve through the four selected points.
	ctx.MoveTo(x0, y0)
	ctx.CubicTo(x3, y3, x4, y4, x1, y1)
}

// bulb draws a shaky bulb (used for line endings).
func (ctx *Canvas) bulb(x0, y0 float64) {
	fuzziness := random()*2 - 1

	for i := 0; i < 3; i++ {
		ctx.DrawArc(x0+fuzziness, y0+fuzziness, 5, 0, math.Pi*2)
		ctx.ClosePath()
		ctx.Fill()
	}
}

// arrowHead draws a shaky arrowhead at the (x1, y1) as an ending
// for the line from (x0, y0) to (x1, y1).
func (ctx *Canvas) arrowHead(x0, y0, x1, y1 float64) {
	dx := x0 - x1
	dy := y0 - y1

	alpha := math.Atan(dy / dx)

	if dy == 0 {
		if dx < 0 {
			alpha = -math.Pi
		} else {
			alpha = 0
		}
	}
	alpha3 := alpha + 0.5
	alpha4 := alpha - 0.5

	l3 := float64(20.0)
	x3 := x1 + l3*math.Cos(alpha3)
	y3 := y1 + l3*math.Sin(alpha3)

	ctx.moveTo(x3, y3)
	ctx.lineTo(x1, y1)
	ctx.Stroke()

	l4 := float64(20.0)
	x4 := x1 + l4*math.Cos(alpha4)
	y4 := y1 + l4*math.Sin(alpha4)

	ctx.moveTo(x4, y4)
	ctx.lineTo(x1, y1)
	ctx.Stroke()
}

// fillText fill out the text.
func (ctx *Canvas) fillText(text string, x0, y0 float64) {
	ctx.DrawString(text, x0, y0)
}

// Draw draws the text annotation at (x0, y0) with the given color.
func (text *Text) Draw(ctx *Canvas) {
	ctx.SetHexColor(text.color)
	ctx.fillText(text.text, X(float64(text.x0)), Y(float64(text.y0)+0.5))
}

// Draw draws a line from (x0, y0) to (x1, y1) with the given color.
func (line *Line) Draw(ctx *Canvas) {
	ctx.SetHexColor(line.color)
	ctx.SetLineWidth(ctx.lineWidth)
	ctx.moveTo(X(float64(line.x0)), Y(float64(line.y0)))
	ctx.lineTo(X(float64(line.x1)), Y(float64(line.y1)))
	ctx.Stroke()

	// Draw given type of ending on the (x1, y1).
	_ending := func(ctx *Canvas, typ string, x0, y0, x1, y1 float64) {
		switch typ {
		case "circle":
			ctx.bulb(x1, y1)
			return
		case "arrow":
			ctx.arrowHead(x0, y0, x1, y1)
			return
		}
	}

	_ending(ctx, line.start, X(float64(line.x1)), Y(float64(line.y1)), X(float64(line.x0)), Y(float64(line.y0)))
	_ending(ctx, line.end, X(float64(line.x0)), Y(float64(line.y0)), X(float64(line.x1)), Y(float64(line.y1)))
}

// X returns the symbols x position.
func X(x float64) float64 {
	return x*CellSize + (CellSize / 2)
}

// Y returns the symbol y position.
func Y(y float64) float64 {
	return y*CellSize + (CellSize / 2)
}
