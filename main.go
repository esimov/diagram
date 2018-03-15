// Package Diagram is a Go package to generate hand drawn diagrams from ASCII arts.
//
// It's a full featured CLI application which converts the ASCII text into hand drawn diagrams.
package main

import (
	"flag"
	"go/build"
	"image"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/esimov/diagram/canvas"
	"github.com/esimov/diagram/io"
	"github.com/esimov/diagram/ui"
	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtk"
)

var defaultFontFile = build.Default.GOPATH + "/src/github.com/esimov/diagram" + "/font/gloriahallelujah.ttf"

var (
	source      = flag.String("in", "", "Source")
	destination = flag.String("out", "", "Destination")
	fontPath    = flag.String("font", defaultFontFile, "Path to the font file")
	preview     = flag.Bool("preview", true, "Show the preview window")
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	flag.Parse()
	// In case the option parameters are used, the hand-drawn diagrams are generated without to enter into the CLI app.
	if (*source != "") && (*destination != "") {
		input := string(io.ReadFile(*source))

		err := canvas.DrawDiagram(input, *destination, *fontPath)
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

			gtk.Init(nil)
			window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
			window.SetPosition(gtk.WIN_POS_CENTER)
			window.SetTitle("Diagram Preview")
			window.Connect("destroy", func(ctx *glib.CallbackContext) {
				gtk.MainQuit()
			}, "")
			frame := gtk.NewVBox(false, 1)
			image := gtk.NewImageFromFile(*destination)
			frame.Add(image)
			window.SetSizeRequest(source.Bounds().Max.X, source.Bounds().Max.Y)
			window.ShowAll()
			gtk.Main()
		}
	} else {
		ui.InitApp(*fontPath)
	}
}
