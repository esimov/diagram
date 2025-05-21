package ui

import "github.com/jroimartin/gocui"

type ComponentHandler interface {
	Draw() (*gocui.View, error)
	NextElement(g *gocui.Gui, v *gocui.View) error
	PrevElement(g *gocui.Gui, v *gocui.View) error
}

type Key interface{}

type Handlers map[Key]HandlerFn
