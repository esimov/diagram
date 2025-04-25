// Diagram is a Go library to generate hand drawn diagrams from ASCII arts.
//
// It's a full featured CLI application which converts the ASCII text into hand drawn diagrams.
package main

import (
	"flag"
	"fmt"
	"image"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/esimov/diagram/canvas"
	"github.com/esimov/diagram/gui"
	"github.com/esimov/diagram/io"
	"github.com/esimov/diagram/ui"
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

var defaultFontFile = "font/gloriahallelujah.ttf"

var (
	source      = flag.String("in", "", "Source")
	destination = flag.String("out", "", "Destination")
	fontPath    = flag.String("font", defaultFontFile, "Path to the font file")
	preview     = flag.Bool("preview", true, "Show the preview window")
)

func main() {
	rand.NewSource(time.Now().UnixNano())

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(HelpBanner, version))
		flag.PrintDefaults()
	}
	flag.Parse()

	// In case the option parameters are used, the hand-drawn diagrams are generated without to enter into the CLI app.
	if (*source != "") && (*destination != "") {
		input, err := io.ReadFile("sample.txt")
		if err != nil {
			log.Fatalf("error reading the sample file: %v", err)
		}

		err = canvas.DrawDiagram(string(input), *destination, *fontPath)
		if err != nil {
			log.Fatal("Error on converting the ascii art to hand drawn diagrams!")
		} else if *preview {
			f, err := os.Open(*destination)
			if err != nil {
				log.Fatalf("Failed to open image '%s': %v\n", *destination, err)
			}

			source, _, err := image.Decode(f)
			if err != nil {
				log.Fatalf("Failed to read image '%s': %v\n", *destination, err)
			}

			gui := gui.NewGUI(source)

			if err := gui.Draw(); err != nil {
				log.Fatalf("diagram GUI draw error: %v", err)
			}
		}
	} else {
		ui.InitApp(*fontPath)
	}
}
