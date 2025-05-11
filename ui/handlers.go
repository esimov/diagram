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
	{nil, gocui.KeyPgup, "PgUp", "Jump to the next page", nil},
	{nil, gocui.KeyPgdn, "PgDown", "Jump to the previous page", nil},
	{nil, gocui.KeyHome, "Home", "Jump to the start of the line", nil},
	{nil, gocui.KeyEnd, "End", "Jump to the end of the line", nil},
	*getDeleteHandler(),
	{nil, gocui.KeyCtrlX, "Ctrl+x", "Clear editor content", nil},
	{nil, gocui.KeyCtrlZ, "Ctrl+z", "Restore editor content", nil},
	{nil, gocui.KeyCtrlS, "Ctrl+s", "Save diagram", onDiagramSave},
	{nil, gocui.KeyCtrlG, "Ctrl+g", "Generate diagram", onDiagramGenerate},
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
		err := ui.generateDiagram(editorPanel)
		if err != nil {
			return ui.log(fmt.Sprintf("Error saving the ASCII diagram: %v", err), true)
		}

		return nil
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

		return ui.loadContent(editorPanel)
	}
	onUp := func(g *gocui.Gui, v *gocui.View) error {
		cx, cy := v.Cursor()

		if cy > 0 {
			v.SetCursor(cx, cy-1)
		}

		return ui.loadContent(editorPanel)
	}
	onDelete := func(g *gocui.Gui, v *gocui.View) error {
		if ui.logTimer != nil {
			ui.logTimer.Stop()
		}

		cx, cy := v.Cursor()
		cv, err := ui.gui.View(diagramsPanel)
		if err != nil {
			return err
		}
		cwd, err := filepath.Abs(filepath.Dir(""))
		if err != nil {
			return err
		}

		currentFile = ui.getViewRow(cv, cy)
		if len(currentFile) == 0 { // this means that the file list is empty!
			return nil
		}
		fn := fmt.Sprintf("%s/%s/%s", cwd, mainDir, currentFile)

		err = io.DeleteFile(fn)
		if err != nil {
			return err
		}
		ui.log(fmt.Sprintf("The file %q has been deleted successfully from the %q directory", currentFile, cwd), false)

		if err := ui.updateDiagramList(diagramsPanel); err != nil {
			return fmt.Errorf("cannot update the diagram list: %w", err)
		}

		if cy > 0 {
			v.SetCursor(cx, cy-1)
		}

		_ = ui.loadContent(editorPanel)

		// Hide log message after 4 seconds
		defer func() {
			ui.logTimer = time.AfterFunc(4*time.Second, func() {
				ui.gui.Update(func(*gocui.Gui) error {
					return ui.clearLog()
				})
			})
		}()

		return nil
	}

	// Activate key bindings
	if err := g.SetKeybinding(diagramsPanel, gocui.KeyArrowDown, gocui.ModNone, onDown); err != nil {
		return err
	}
	if err := g.SetKeybinding(diagramsPanel, gocui.KeyArrowUp, gocui.ModNone, onUp); err != nil {
		return err
	}
	if err := g.SetKeybinding(editorPanel, gocui.MouseWheelDown, gocui.ModNone, ui.scrollDown); err != nil {
		return err
	}
	if err := g.SetKeybinding(editorPanel, gocui.MouseWheelUp, gocui.ModNone, ui.scrollUp); err != nil {
		return err
	}

	if runtime.GOOS == "darwin" {
		if err := g.SetKeybinding(diagramsPanel, gocui.KeyBackspace2, gocui.ModNone, onDelete); err != nil {
			return err
		}
	} else {
		if err := g.SetKeybinding(diagramsPanel, gocui.KeyDelete, gocui.ModNone, onDelete); err != nil {
			return err
		}
	}

	return g.SetKeybinding("", gocui.KeyF1, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return ui.toggleHelpModal(handlers.helpContent())
	})
}

// helpContent populates the help panel.
func (handlers handlers) helpContent() string {
	buf := &bytes.Buffer{}
	w := tabwriter.NewWriter(buf, 0, 0, 3, ' ', tabwriter.DiscardEmptyColumns)
	for _, handler := range handlers {
		if handler.keyName == "" || handler.help == "" {
			continue
		}
		fmt.Fprintf(w, "  %s\t: %s\n", handler.keyName, handler.help)
	}

	fmt.Fprintf(w, "  %s\t: %s\n", "F1", "Show/hide help panel")
	w.Flush()

	return buf.String()
}
