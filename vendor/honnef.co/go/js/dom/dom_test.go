package dom

import (
	"image"
	"image/draw"
)

var _ Node = &BasicNode{}
var _ HTMLElement = &BasicHTMLElement{}
var _ Element = &BasicElement{}
var _ Document = &document{}
var _ Window = &window{}
var _ HTMLDocument = &htmlDocument{}
var _ image.Image = &ImageData{}
var _ draw.Image = &ImageData{}
