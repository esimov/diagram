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
	{nil, gocui.KeyCtrlX, "Ctrl+x", "Clear editor content", nil},
	{nil, gocui.KeyCtrlZ, "Ctrl+z", "Restore diagram", nil},
	{nil, gocui.KeyPgup, "PgUp", "Jump to the top", nil},
	{nil, gocui.KeyPgdn, "PgDown", "Jump to the bottom", nil},
	{nil, gocui.KeyHome, "Home", "Jump to the start", nil},
	{nil, gocui.KeyEnd, "End", "Jump to the end", nil},
	{nil, gocui.KeyCtrlS, "Ctrl+s", "Save diagram", onSaveDiagram},
	{nil, gocui.KeyCtrlC, "Ctrl+c", "Quit", onQuit},
}

func onNextPanel(ui *UI, wrap bool) Fn {
	return func(*gocui.Gui, *gocui.View) error {
		return ui.nextView(wrap)
	}
}

func onPrevPanel(ui *UI, wrap bool) Fn {
	return func(*gocui.Gui, *gocui.View) error {
		return ui.prevView(wrap)
	}
}

func onQuit(ui *UI, wrap bool) Fn {
	return func(*gocui.Gui, *gocui.View) error {
		return gocui.ErrQuit
	}
}

func onSaveDiagram(ui *UI, wrap bool) Fn {
	return func(*gocui.Gui, *gocui.View) error {
		return ui.saveDiagram(DIAGRAM_PANEL)
	}
}

// Apply key bindings to panel views
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
	onDown := func(g *gocui.Gui, v *gocui.View) error {
		cx, cy := v.Cursor()

		if cy < ui.getViewTotalRows(v)-1 {
			v.SetCursor(cx, cy+1)
		}
		ui.modifyView(DIAGRAM_PANEL)
		return nil
	}
	onUp := func(g *gocui.Gui, v *gocui.View) error {
		cx, cy := v.Cursor()

		if cy > 0 {
			v.SetCursor(cx, cy-1)
		}
		ui.modifyView(DIAGRAM_PANEL)
		return nil
	}

	if err := g.SetKeybinding(SAVED_DIAGRAMS_PANEL, gocui.KeyArrowDown, gocui.ModNone, onDown); err != nil {
		return err
	}

	if err := g.SetKeybinding(SAVED_DIAGRAMS_PANEL, gocui.KeyArrowUp, gocui.ModNone, onUp); err != nil {
		return err
	}

	return g.SetKeybinding("", gocui.KeyCtrlH, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return ui.toggleHelp(g, handlers.HelpContent())
	})
}

// Populate the help panel
func (handlers handlers) HelpContent() string {
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