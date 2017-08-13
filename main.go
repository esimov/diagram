package main

import (
	"github.com/esimov/diagram/ui"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	ui.InitApp()
}