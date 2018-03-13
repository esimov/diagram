package ui

import (
	"bytes"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"text/tabwriter"

	"github.com/esimov/diagram/io"
	"github.com/jroimartin/gocui"
)

// Fn is a generic function acting as a closure function for event handlers.
type Fn func(*gocui.Gui, *gocui.View) error

type handler struct {
	views   []string
	key     interface{}
	keyName string
	help    string
	action  func(*UI, bool) Fn
}

type handlers []handler

var keyHandlers = &handlers{
	{mainViews, gocui.KeyTab, "Tab", "Next Panel", onNextPanel},
	{mainViews, 0xFF, "Shift+Tab", "Previous Panel", nil},
	{nil, gocui.KeyPgup, "PgUp", "Jump to the top", nil},
	{nil, gocui.KeyPgdn, "PgDown", "Jump to the bottom", nil},
	{nil, gocui.KeyHome, "Home", "Jump to the start", nil},
	{nil, gocui.KeyEnd, "End", "Jump to the end", nil},
	*getDeleteHandler(),
	{nil, gocui.KeyCtrlX, "Ctrl+x", "Clear editor content", nil},
	{nil, gocui.KeyCtrlZ, "Ctrl+z", "Restore editor content", nil},
	{nil, gocui.KeyCtrlS, "Ctrl+s", "Save diagram", onSaveDiagram},
	{nil, gocui.KeyCtrlD, "Ctrl+d", "Draw diagram", onDrawDiagram},
	{nil, gocui.KeyCtrlC, "Ctrl+c", "Quit", onQuit},
}

// getDeleteHandler defines and returns a delete view handler
func getDeleteHandler() *handler {
	if runtime.GOOS == "Darwin" {
		return &handler{nil, gocui.KeyBackspace2, "Backspace", "Delete diagram", nil}
	}
	return &handler{nil, gocui.KeyDelete, "Delete", "Delete diagram", nil}
}

// onNextPanel retrieves the next panel.
func onNextPanel(ui *UI, wrap bool) Fn {
	return func(*gocui.Gui, *gocui.View) error {
		return ui.nextView(wrap)
	}
}

// onPrevPanel retrieves the previous panel.
func onPrevPanel(ui *UI, wrap bool) Fn {
	return func(*gocui.Gui, *gocui.View) error {
		return ui.prevView(wrap)
	}
}

// onQuit is an event listener which get triggered when a quit action is performed.
func onQuit(ui *UI, wrap bool) Fn {
	return func(*gocui.Gui, *gocui.View) error {
		return gocui.ErrQuit
	}
}

// onSaveDiagram is an event listener which get triggered when a save action is performed.
func onSaveDiagram(ui *UI, wrap bool) Fn {
	return func(*gocui.Gui, *gocui.View) error {
		return ui.saveDiagram(DIAGRAM_PANEL)
	}
}

// onDrawDiagram is an event listener which get triggered when a draw action is performed.
func onDrawDiagram(ui *UI, wrap bool) Fn {
	return func(*gocui.Gui, *gocui.View) error {
		return ui.drawDiagram(DIAGRAM_PANEL)
	}
}

// ApplyKeyBindings applies key bindings to panel views.
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
	onDelete := func(g *gocui.Gui, v *gocui.View) error {
		cx, cy := v.Cursor()
		cv, err := ui.gui.View(SAVED_DIAGRAMS_PANEL)
		if err != nil {
			return err
		}
		cwd, err := filepath.Abs(filepath.Dir(""))
		if err != nil {
			log.Fatal(err)
		}
		currentFile = ui.getViewRow(cv, cy)[0]
		fn := cwd + "/" + DIAGRAMS_DIR + "/" + currentFile

		io.DeleteDiagram(fn)
		ui.updateDiagramList(SAVED_DIAGRAMS_PANEL)

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

	if runtime.GOOS == "darwin" {
		if err := g.SetKeybinding(SAVED_DIAGRAMS_PANEL, gocui.KeyBackspace2, gocui.ModNone, onDelete); err != nil {
			return err
		}
	} else {
		if err := g.SetKeybinding(SAVED_DIAGRAMS_PANEL, gocui.KeyDelete, gocui.ModNone, onDelete); err != nil {
			return err
		}
	}

	return g.SetKeybinding("", gocui.KeyCtrlH, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return ui.toggleHelp(g, handlers.HelpContent())
	})
}

// HelpContent populates the help panel.
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
