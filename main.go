// Package Diagram is a Go package to generate hand drawn diagrams from ASCII arts.
//
// It's a full featured CLI application which converts the ASCII text into hand drawn diagrams.
package main

import (
	"flag"
	"fmt"
	"go/build"
	"image"
	"image/draw"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/esimov/diagram/canvas"
	"github.com/esimov/diagram/io"
	"github.com/esimov/diagram/ui"
	"github.com/google/gxui"
	"github.com/google/gxui/drivers/gl"
	"github.com/google/gxui/samples/flags"
)

const HelpBanner = `
┌┬┐┬┌─┐┌─┐┬─┐┌─┐┌┬┐
 │││├─┤│ ┬├┬┘├─┤│││
─┴┘┴┴ ┴└─┘┴└─┴ ┴┴ ┴
    Version: %s

CLI app to convert ASCII arts into hand drawn diagrams.

`
// Version indicates the current build version.
var version string

var defaultFontFile = build.Default.GOPATH + "/src/github.com/esimov/diagram" + "/font/gloriahallelujah.ttf"

var (
	source      = flag.String("in", "", "Source")
	destination = flag.String("out", "", "Destination")
	fontPath    = flag.String("font", defaultFontFile, "Path to the font file")
	preview     = flag.Bool("preview", true, "Show the preview window")
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, fmt.Sprintf(HelpBanner, version))
		flag.PrintDefaults()
	}
	flag.Parse()

	// In case the option parameters are used, the hand-drawn diagrams are generated without to enter into the CLI app.
	if (*source != "") && (*destination != "") {
		input := string(io.ReadFile(*source))

		err := canvas.DrawDiagram(input, *destination, *fontPath)
		if err != nil {
			log.Fatal("Error on converting the ascii art to hand drawn diagrams!")
		} else if *preview {
			gl.StartDriver(func(driver gxui.Driver) {
				f, err := os.Open(*destination)
				if err != nil {
					log.Fatalf("Failed to open image '%s': %v\n", *destination, err)
				}
				source, _, err := image.Decode(f)
				if err != nil {
					log.Fatalf("Failed to read image '%s': %v\n", *destination, err)
				}
				theme := flags.CreateTheme(driver)
				img := theme.CreateImage()

				window := theme.CreateWindow(source.Bounds().Max.X, source.Bounds().Max.Y, "Diagram preview")
				window.SetScale(flags.DefaultScaleFactor)
				window.AddChild(img)

				// Copy the image to a RGBA format before handing to a gxui.Texture
				rgba := image.NewRGBA(source.Bounds())
				draw.Draw(rgba, source.Bounds(), source, image.ZP, draw.Src)
				texture := driver.CreateTexture(rgba, 1)
				img.SetTexture(texture)

				window.OnClose(driver.Terminate)
			})
		}
	} else {
		ui.InitApp(*fontPath)
	}
}
