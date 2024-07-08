package ui

import (
	"bytes"
	"fmt"
	"path/filepath"
	"runtime"
	"text/tabwriter"
	"time"

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
	{nil, gocui.KeyPgup, "PgUp", "Jump to the first line", nil},
	{nil, gocui.KeyPgdn, "PgDown", "Jump to the last line", nil},
	{nil, gocui.KeyHome, "Home", "Jump to the line start", nil},
	{nil, gocui.KeyEnd, "End", "Jump to the line end", nil},
	*getDeleteHandler(),
	{nil, gocui.KeyCtrlX, "Ctrl+x", "Clear editor content", nil},
	{nil, gocui.KeyCtrlZ, "Ctrl+z", "Restore editor content", nil},
	{nil, gocui.KeyCtrlS, "Ctrl+s", "Save diagram", onDiagramSave},
	{nil, gocui.KeyCtrlG, "Ctrl+p", "Generate diagram", onDiagramGenerate},
	{nil, gocui.KeyCtrlQ, "Ctrl+q", "Quit", onQuit},
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

// onDiagramSave is an event listener which get triggered when a save action is performed.
func onDiagramSave(ui *UI, wrap bool) Fn {
	return func(*gocui.Gui, *gocui.View) error {
		return ui.saveDiagram(editorPanel)
	}
}

// onDiagramGenerate is an event listener which get triggered when a draw action is performed.
func onDiagramGenerate(ui *UI, wrap bool) Fn {
	return func(*gocui.Gui, *gocui.View) error {
		return ui.generateDiagram(editorPanel)
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
		ui.modifyView(editorPanel)
		return nil
	}
	onUp := func(g *gocui.Gui, v *gocui.View) error {
		cx, cy := v.Cursor()

		if cy > 0 {
			v.SetCursor(cx, cy-1)
		}
		ui.modifyView(editorPanel)
		return nil
	}
	onDelete := func(g *gocui.Gui, v *gocui.View) error {
		if ui.logTimer != nil {
			ui.logTimer.Stop()
		}

		cx, cy := v.Cursor()
		cv, err := ui.gui.View(savedDiagramsPanel)
		if err != nil {
			return err
		}
		cwd, err := filepath.Abs(filepath.Dir(""))
		if err != nil {
			return err
		}
		currentFile = ui.getViewRow(cv, cy)[0]
		fn := fmt.Sprintf("%s/%s/%s", cwd, mainDir, currentFile)

		err = io.DeleteDiagram(fn)
		if err != nil {
			return err
		}
		ui.log(fmt.Sprintf("The file %s has been deleted successfully from the %s directory", currentFile, cwd), false)
		ui.updateDiagramList(savedDiagramsPanel)

		if cy > 0 {
			v.SetCursor(cx, cy-1)
		}
		ui.modifyView(editorPanel)

		// Hide log message after 4 seconds
		ui.logTimer = time.AfterFunc(4*time.Second, func() {
			ui.gui.Update(func(*gocui.Gui) error {
				return ui.clearLog()
			})
		})

		return nil
	}

	if err := g.SetKeybinding(savedDiagramsPanel, gocui.KeyArrowDown, gocui.ModNone, onDown); err != nil {
		return err
	}

	if err := g.SetKeybinding(savedDiagramsPanel, gocui.KeyArrowUp, gocui.ModNone, onUp); err != nil {
		return err
	}

	if runtime.GOOS == "darwin" {
		if err := g.SetKeybinding(savedDiagramsPanel, gocui.KeyBackspace2, gocui.ModNone, onDelete); err != nil {
			return err
		}
	} else {
		if err := g.SetKeybinding(savedDiagramsPanel, gocui.KeyDelete, gocui.ModNone, onDelete); err != nil {
			return err
		}
	}

	return g.SetKeybinding("", gocui.KeyCtrlH, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return ui.toggleHelp(handlers.HelpContent())
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
