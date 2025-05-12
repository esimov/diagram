package ui

import "github.com/jroimartin/gocui"

type WidgetHandler interface {
	Draw() (*gocui.View, error)
	NextElement(views []string) error
	PrevElement(views []string) error
}
