package main

import (
	"github.com/esimov/diagram/io"
	"github.com/esimov/diagram/ui"
	"github.com/esimov/diagram/canvas"
	"math/rand"
	"time"
	"flag"
	"log"
	"os"
	"github.com/fogleman/imview"
)

var (
	source		= flag.String("in", "", "Source")
	destination 	= flag.String("out", "", "Destination")
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	// Generate diagram directly with command line tool.
	if len(os.Args) > 1 {
		flag.Parse()
		input := string(io.ReadFile(*source))

		err := canvas.DrawDiagram(input, *destination)
		if err != nil {
			log.Fatal("Error on converting the ascii art to hand drawn diagrams!")
		} else {
			image, _ := imview.LoadImage(*destination)
			view := imview.ImageToRGBA(image)
			imview.Show(view)
		}
	} else {
		ui.InitApp()
	}
}