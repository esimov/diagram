package main

import (
	"flag"
	"go/build"
	"log"
	"math/rand"
	"time"

	"github.com/esimov/diagram/canvas"
	"github.com/esimov/diagram/io"
	"github.com/esimov/diagram/ui"
	"github.com/fogleman/imview"
)

var defaultFontFile = build.Default.GOPATH + "/src/github.com/esimov/diagram" + "/font/gloriahallelujah.ttf"

var (
	source      = flag.String("in", "", "Source")
	destination = flag.String("out", "", "Destination")
	fontpath    = flag.String("font", defaultFontFile, "path to font file")
	preview     = flag.Bool("preview", true, "Show the preview window")
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	flag.Parse()
	// If filenames specified on the commandline generate diagram directly with command line tool.
	if (*source != "") && (*destination != "") {
		input := string(io.ReadFile(*source))

		err := canvas.DrawDiagram(input, *destination, *fontpath)
		if err != nil {
			log.Fatal("Error on converting the ascii art to hand drawn diagrams!")
		} else if *preview {
			image, _ := imview.LoadImage(*destination)
			view := imview.ImageToRGBA(image)
			imview.Show(view)
		}
	} else {
		ui.InitApp(*fontpath)
	}
}
