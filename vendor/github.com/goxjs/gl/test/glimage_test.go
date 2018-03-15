// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin linux,!android

// TODO(crawshaw): Run tests on other OSs when more contexts are supported.

package test

import (
	"encoding/binary"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/goxjs/gl"
	"github.com/goxjs/gl/glutil"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/geom"
)

func TestImage(t *testing.T) {
	// GL testing strategy:
	// 	1. Create an offscreen framebuffer object.
	// 	2. Configure framebuffer to render to a GL texture.
	//	3. Run test code: use glimage to draw testdata.
	//	4. Copy GL texture back into system memory.
	//	5. Compare to a pre-computed image.

	f, err := os.Open(filepath.Join("testdata", "testpattern.png"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	src, _, err := image.Decode(f)
	if err != nil {
		t.Fatal(err)
	}

	ctxGL := createContext()
	defer ctxGL.destroy()

	const (
		pixW = 100
		pixH = 100
		ptW  = geom.Pt(50)
		ptH  = geom.Pt(50)
	)
	cfg := size.Event{
		WidthPx:     pixW,
		HeightPx:    pixH,
		WidthPt:     ptW,
		HeightPt:    ptH,
		PixelsPerPt: float32(pixW) / float32(ptW),
	}

	fBuf := gl.CreateFramebuffer()
	gl.BindFramebuffer(gl.FRAMEBUFFER, fBuf)
	colorBuf := gl.CreateRenderbuffer()
	gl.BindRenderbuffer(gl.RENDERBUFFER, colorBuf)
	// https://www.khronos.org/opengles/sdk/docs/man/xhtml/glRenderbufferStorage.xml
	// says that the internalFormat "must be one of the following symbolic constants:
	// GL_RGBA4, GL_RGB565, GL_RGB5_A1, GL_DEPTH_COMPONENT16, or GL_STENCIL_INDEX8".
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.RGB565, pixW, pixH)
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.RENDERBUFFER, colorBuf)

	if status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER); status != gl.FRAMEBUFFER_COMPLETE {
		t.Fatalf("framebuffer create failed: %v", status)
	}

	gl.ClearColor(0, 0, 1, 1) // blue
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.Viewport(0, 0, pixW, pixH)

	m := NewImage(src.Bounds().Dx(), src.Bounds().Dy())
	b := m.RGBA.Bounds()
	draw.Draw(m.RGBA, b, src, src.Bounds().Min, draw.Src)
	m.Upload()
	b.Min.X += 10
	b.Max.Y /= 2

	// All-integer right-angled triangles offsetting the
	// box: 24-32-40, 12-16-20.
	ptTopLeft := geom.Point{0, 24}
	ptTopRight := geom.Point{32, 0}
	ptBottomLeft := geom.Point{12, 24 + 16}
	ptBottomRight := geom.Point{12 + 32, 16}
	m.Draw(cfg, ptTopLeft, ptTopRight, ptBottomLeft, b)

	// For unknown reasons, a windowless OpenGL context renders upside-
	// down. That is, a quad covering the initial viewport spans:
	//
	//	(-1, -1) ( 1, -1)
	//	(-1,  1) ( 1,  1)
	//
	// To avoid modifying live code for tests, we flip the rows
	// recovered from the renderbuffer. We are not the first:
	//
	// http://lists.apple.com/archives/mac-opengl/2010/Jun/msg00080.html
	got := image.NewRGBA(image.Rect(0, 0, pixW, pixH))
	upsideDownPix := make([]byte, len(got.Pix))
	gl.ReadPixels(upsideDownPix, 0, 0, pixW, pixH, gl.RGBA, gl.UNSIGNED_BYTE)
	for y := 0; y < pixH; y++ {
		i0 := (pixH - 1 - y) * got.Stride
		i1 := i0 + pixW*4
		copy(got.Pix[y*got.Stride:], upsideDownPix[i0:i1])
	}

	drawCross(got, 0, 0)
	drawCross(got, int(ptTopLeft.X.Px(cfg.PixelsPerPt)), int(ptTopLeft.Y.Px(cfg.PixelsPerPt)))
	drawCross(got, int(ptBottomRight.X.Px(cfg.PixelsPerPt)), int(ptBottomRight.Y.Px(cfg.PixelsPerPt)))
	drawCross(got, pixW-1, pixH-1)

	var wantPath = filepath.Join("testdata", "testpattern-window.png")
	f, err = os.Open(wantPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	wantSrc, _, err := image.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	want, ok := wantSrc.(*image.RGBA)
	if !ok {
		b := wantSrc.Bounds()
		want = image.NewRGBA(b)
		draw.Draw(want, b, wantSrc, b.Min, draw.Src)
	}

	if !imageEq(got, want) {
		// Write out the image we got.
		f, err = ioutil.TempFile("", "testpattern-window-got")
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
		gotPath := f.Name() + ".png"
		f, err = os.Create(gotPath)
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, got); err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
		t.Errorf("got\n%s\nwant\n%s", gotPath, wantPath)
	}
}

func drawCross(m *image.RGBA, x, y int) {
	c := color.RGBA{0xff, 0, 0, 0xff} // red
	m.SetRGBA(x+0, y-2, c)
	m.SetRGBA(x+0, y-1, c)
	m.SetRGBA(x-2, y+0, c)
	m.SetRGBA(x-1, y+0, c)
	m.SetRGBA(x+0, y+0, c)
	m.SetRGBA(x+1, y+0, c)
	m.SetRGBA(x+2, y+0, c)
	m.SetRGBA(x+0, y+1, c)
	m.SetRGBA(x+0, y+2, c)
}

func eqEpsilon(x, y uint8) bool {
	const epsilon = 9
	return x-y < epsilon || y-x < epsilon
}

func colorEq(c0, c1 color.RGBA) bool {
	return eqEpsilon(c0.R, c1.R) && eqEpsilon(c0.G, c1.G) && eqEpsilon(c0.B, c1.B) && eqEpsilon(c0.A, c1.A)
}

func imageEq(m0, m1 *image.RGBA) bool {
	b0 := m0.Bounds()
	b1 := m1.Bounds()
	if b0 != b1 {
		return false
	}
	badPx := 0
	for y := b0.Min.Y; y < b0.Max.Y; y++ {
		for x := b0.Min.X; x < b0.Max.X; x++ {
			c0, c1 := m0.At(x, y).(color.RGBA), m1.At(x, y).(color.RGBA)
			if !colorEq(c0, c1) {
				badPx++
			}
		}
	}
	badFrac := float64(badPx) / float64(b0.Dx()*b0.Dy())
	return badFrac < 0.01
}

var glimage struct {
	sync.Once
	quadXY        gl.Buffer
	quadUV        gl.Buffer
	program       gl.Program
	pos           gl.Attrib
	mvp           gl.Uniform
	uvp           gl.Uniform
	inUV          gl.Attrib
	textureSample gl.Uniform
}

func glInit() {
	var err error
	glimage.program, err = glutil.CreateProgram(vertexShader, fragmentShader)
	if err != nil {
		panic(err)
	}

	glimage.quadXY = gl.CreateBuffer()
	glimage.quadUV = gl.CreateBuffer()

	gl.BindBuffer(gl.ARRAY_BUFFER, glimage.quadXY)
	gl.BufferData(gl.ARRAY_BUFFER, quadXYCoords, gl.STATIC_DRAW)
	gl.BindBuffer(gl.ARRAY_BUFFER, glimage.quadUV)
	gl.BufferData(gl.ARRAY_BUFFER, quadUVCoords, gl.STATIC_DRAW)

	glimage.pos = gl.GetAttribLocation(glimage.program, "pos")
	glimage.mvp = gl.GetUniformLocation(glimage.program, "mvp")
	glimage.uvp = gl.GetUniformLocation(glimage.program, "uvp")
	glimage.inUV = gl.GetAttribLocation(glimage.program, "inUV")
	glimage.textureSample = gl.GetUniformLocation(glimage.program, "textureSample")
}

// Image bridges between an *image.RGBA and an OpenGL texture.
//
// The contents of the embedded *image.RGBA can be uploaded as a
// texture and drawn as a 2D quad.
//
// The number of active Images must fit in the system's OpenGL texture
// limit. The typical use of an Image is as a texture atlas.
type Image struct {
	*image.RGBA

	Texture   gl.Texture
	texWidth  int
	texHeight int
}

// NewImage creates an Image of the given size.
//
// Both a host-memory *image.RGBA and a GL texture are created.
func NewImage(w, h int) *Image {
	dx := roundToPower2(w)
	dy := roundToPower2(h)

	// TODO(crawshaw): Using VertexAttribPointer we can pass texture
	// data with a stride, which would let us use the exact number of
	// pixels on the host instead of the rounded up power 2 size.
	m := image.NewRGBA(image.Rect(0, 0, dx, dy))

	glimage.Do(glInit)

	img := &Image{
		RGBA:      m.SubImage(image.Rect(0, 0, w, h)).(*image.RGBA),
		Texture:   gl.CreateTexture(),
		texWidth:  dx,
		texHeight: dy,
	}
	// TODO(crawshaw): We don't have the context on a finalizer. Find a way.
	// runtime.SetFinalizer(img, func(img *Image) { gl.DeleteTexture(img.Texture) })
	gl.BindTexture(gl.TEXTURE_2D, img.Texture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, dx, dy, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	return img
}

func roundToPower2(x int) int {
	x2 := 1
	for x2 < x {
		x2 *= 2
	}
	return x2
}

// Upload copies the host image data to the GL device.
func (img *Image) Upload() {
	gl.BindTexture(gl.TEXTURE_2D, img.Texture)
	gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, img.texWidth, img.texHeight, gl.RGBA, gl.UNSIGNED_BYTE, img.Pix)
}

// Draw draws the srcBounds part of the image onto a parallelogram, defined by
// three of its corners, in the current GL framebuffer.
func (img *Image) Draw(c size.Event, topLeft, topRight, bottomLeft geom.Point, srcBounds image.Rectangle) {
	// TODO(crawshaw): Adjust viewport for the top bar on android?
	gl.UseProgram(glimage.program)

	{
		// We are drawing a parallelogram PQRS, defined by three of its
		// corners, onto the entire GL framebuffer ABCD. The two quads may
		// actually be equal, but in the general case, PQRS can be smaller,
		// and PQRS is not necessarily axis-aligned.
		//
		//	A +---------------+ B
		//	  |  P +-----+ Q  |
		//	  |    |     |    |
		//	  |  S +-----+ R  |
		//	D +---------------+ C
		//
		// There are two co-ordinate spaces: geom space and framebuffer space.
		// In geom space, the ABCD rectangle is:
		//
		//	(0, 0)           (geom.Width, 0)
		//	(0, geom.Height) (geom.Width, geom.Height)
		//
		// and the PQRS quad is:
		//
		//	(topLeft.X,    topLeft.Y)    (topRight.X, topRight.Y)
		//	(bottomLeft.X, bottomLeft.Y) (implicit,   implicit)
		//
		// In framebuffer space, the ABCD rectangle is:
		//
		//	(-1, +1) (+1, +1)
		//	(-1, -1) (+1, -1)
		//
		// First of all, convert from geom space to framebuffer space. For
		// later convenience, we divide everything by 2 here: px2 is half of
		// the P.X co-ordinate (in framebuffer space).
		px2 := -0.5 + float32(topLeft.X/c.WidthPt)
		py2 := +0.5 - float32(topLeft.Y/c.HeightPt)
		qx2 := -0.5 + float32(topRight.X/c.WidthPt)
		qy2 := +0.5 - float32(topRight.Y/c.HeightPt)
		sx2 := -0.5 + float32(bottomLeft.X/c.WidthPt)
		sy2 := +0.5 - float32(bottomLeft.Y/c.HeightPt)
		// Next, solve for the affine transformation matrix
		//	    [ a00 a01 a02 ]
		//	a = [ a10 a11 a12 ]
		//	    [   0   0   1 ]
		// that maps A to P:
		//	a × [ -1 +1 1 ]' = [ 2*px2 2*py2 1 ]'
		// and likewise maps B to Q and D to S. Solving those three constraints
		// implies that C maps to R, since affine transformations keep parallel
		// lines parallel. This gives 6 equations in 6 unknowns:
		//	-a00 + a01 + a02 = 2*px2
		//	-a10 + a11 + a12 = 2*py2
		//	+a00 + a01 + a02 = 2*qx2
		//	+a10 + a11 + a12 = 2*qy2
		//	-a00 - a01 + a02 = 2*sx2
		//	-a10 - a11 + a12 = 2*sy2
		// which gives:
		//	a00 = (2*qx2 - 2*px2) / 2 = qx2 - px2
		// and similarly for the other elements of a.
		writeAffine(glimage.mvp, &f32.Affine{{
			qx2 - px2,
			px2 - sx2,
			qx2 + sx2,
		}, {
			qy2 - py2,
			py2 - sy2,
			qy2 + sy2,
		}})
	}

	{
		// Mapping texture co-ordinates is similar, except that in texture
		// space, the ABCD rectangle is:
		//
		//	(0,0) (1,0)
		//	(0,1) (1,1)
		//
		// and the PQRS quad is always axis-aligned. First of all, convert
		// from pixel space to texture space.
		w := float32(img.texWidth)
		h := float32(img.texHeight)
		px := float32(srcBounds.Min.X-img.Rect.Min.X) / w
		py := float32(srcBounds.Min.Y-img.Rect.Min.Y) / h
		qx := float32(srcBounds.Max.X-img.Rect.Min.X) / w
		sy := float32(srcBounds.Max.Y-img.Rect.Min.Y) / h
		// Due to axis alignment, qy = py and sx = px.
		//
		// The simultaneous equations are:
		//	  0 +   0 + a02 = px
		//	  0 +   0 + a12 = py
		//	a00 +   0 + a02 = qx
		//	a10 +   0 + a12 = qy = py
		//	  0 + a01 + a02 = sx = px
		//	  0 + a11 + a12 = sy
		writeAffine(glimage.uvp, &f32.Affine{{
			qx - px,
			0,
			px,
		}, {
			0,
			sy - py,
			py,
		}})
	}

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, img.Texture)
	gl.Uniform1i(glimage.textureSample, 0)

	gl.BindBuffer(gl.ARRAY_BUFFER, glimage.quadXY)
	gl.EnableVertexAttribArray(glimage.pos)
	gl.VertexAttribPointer(glimage.pos, 2, gl.FLOAT, false, 0, 0)

	gl.BindBuffer(gl.ARRAY_BUFFER, glimage.quadUV)
	gl.EnableVertexAttribArray(glimage.inUV)
	gl.VertexAttribPointer(glimage.inUV, 2, gl.FLOAT, false, 0, 0)

	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)

	gl.DisableVertexAttribArray(glimage.pos)
	gl.DisableVertexAttribArray(glimage.inUV)
}

// writeAffine writes the contents of an Affine to a 3x3 matrix GL uniform.
func writeAffine(u gl.Uniform, a *f32.Affine) {
	var m [9]float32
	m[0*3+0] = a[0][0]
	m[0*3+1] = a[1][0]
	m[0*3+2] = 0
	m[1*3+0] = a[0][1]
	m[1*3+1] = a[1][1]
	m[1*3+2] = 0
	m[2*3+0] = a[0][2]
	m[2*3+1] = a[1][2]
	m[2*3+2] = 1
	gl.UniformMatrix3fv(u, m[:])
}

var quadXYCoords = f32.Bytes(binary.LittleEndian,
	-1, +1, // top left
	+1, +1, // top right
	-1, -1, // bottom left
	+1, -1, // bottom right
)

var quadUVCoords = f32.Bytes(binary.LittleEndian,
	0, 0, // top left
	1, 0, // top right
	0, 1, // bottom left
	1, 1, // bottom right
)

const vertexShader = `#version 100
uniform mat3 mvp;
uniform mat3 uvp;
attribute vec3 pos;
attribute vec2 inUV;
varying vec2 UV;
void main() {
	vec3 p = pos;
	p.z = 1.0;
	gl_Position = vec4(mvp * p, 1);
	UV = (uvp * vec3(inUV, 1)).xy;
}
`

const fragmentShader = `#version 100
precision mediump float;
varying vec2 UV;
uniform sampler2D textureSample;
void main(){
	gl_FragColor = texture2D(textureSample, UV);
}
`
