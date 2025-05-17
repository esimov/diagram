package ui

import "github.com/jroimartin/gocui"

type ComponentHandler interface {
	Draw() (*gocui.View, error)
	NextElement(views []string) error
	PrevElement(views []string) error
}

// Key define kye type
type Key interface{}

type Handlers map[Key]HandlerFn
