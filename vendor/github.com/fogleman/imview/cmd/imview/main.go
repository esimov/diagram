package main

import (
	"image"
	"os"

	"github.com/fogleman/imview"
)

func main() {
	images := LoadImages(os.Args[1:])
	imview.Show(images...)
}

func LoadImages(paths []string) []*image.RGBA {
	var result []*image.RGBA
	for _, path := range paths {
		im, err := imview.LoadImage(path)
		if err != nil {
			continue
		}
		result = append(result, imview.ImageToRGBA(im))
	}
	return result
}
