package ui

import (
	"github.com/jroimartin/gocui"
	"bytes"
	"text/tabwriter"
	"fmt"
)

type Fn func(*gocui.Gui, *gocui.View) error

type handler struct {
	views	[]string
	key	interface{}
	keyName	string
	help	string
	action 	func(*UI, bool) Fn
}

type handlers []handler

var keyHandlers = &handlers{
	{mainViews, gocui.KeyTab, "Tab", "Next Panel", onNextPanel},
	{mainViews, 0xFF, "Shift+Tab", "Previous Panel", nil},
	{nil, gocui.KeyCtrlC, "Ctrl+c", "Quit", onQuit},
}

func onNextPanel(ui *UI, wrap bool) Fn {
	return func(*gocui.Gui, *gocui.View) error {
		return ui.NextView(wrap)
	}
}

func onPrevPanel(ui *UI, wrap bool) Fn {
	return func(*gocui.Gui, *gocui.View) error {
		return ui.PrevView(wrap)
	}
}

func onQuit(ui *UI, wrap bool) Fn {
	return func(*gocui.Gui, *gocui.View) error {
		return gocui.ErrQuit
	}
}

func (handlers handlers) ApplyKeyBindings(ui *UI, g *gocui.Gui) error {
	for _, h := range handlers {
		if len(h.views) == 0 {
			h.views = []string{""}
		}
		if h.action == nil {
			continue
		}

		for _, view := range h.views {
			if err := g.SetKeybinding(view, h.key, gocui.ModNone, h.action(ui, true)); err != nil {
				return err
			}
		}
	}

	return g.SetKeybinding("", gocui.KeyCtrlH, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return ui.toggleHelp(g, keyHandlers.Help())
	})
}

func (handlers handlers) Help() string {
	buf := &bytes.Buffer{}
	w := tabwriter.NewWriter(buf, 0, 0, 3, ' ', tabwriter.DiscardEmptyColumns)
	for _, handler := range handlers {
		if handler.keyName == "" || handler.help == "" {
			continue
		}
		fmt.Fprintf(w, "  %s\t: %s\n", handler.keyName, handler.help)
	}

	fmt.Fprintf(w, "  %s\t: %s\n", "Ctrl+h", "Toggle Help")
	w.Flush()
	return buf.String()
}