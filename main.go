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
)

var (
	source		= flag.String("source", "", "Source")
	destination = flag.String("destination", "", "Destination")
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	// Generate diagram directly with command line tool.
	if len(os.Args) > 1 {
		flag.Parse()
		input := string(io.ReadFile(*source))

		if err := canvas.DrawDiagram(input, *destination); err != nil {
			log.Fatal("Error on converting the ascii art to hand drawn diagrams!")
		}
	} else {
		ui.InitApp()
	}
}