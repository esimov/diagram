package ui

import (
	_ "embed"
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/esimov/diagram/canvas"
	"github.com/esimov/diagram/gui"
	"github.com/esimov/diagram/io"
	"github.com/esimov/diagram/version"
	"github.com/jroimartin/gocui"
)

type panelProperties struct {
	title    string
	text     string
	x1       float64
	y1       float64
	x2       float64
	y2       float64
	editable bool
	cursor   bool
}

// Main views
var panelViews map[string]panelProperties

// Modal views
var modalViews map[string]panelProperties

var (
	// Panel Views
	mainViews = []string{
		logoPanel,
		diagramsPanel,
		logPanel,
		editorPanel,
	}
	layoutOptions = []string{
		defaultLayout.ToString(),
		blackLayout.ToString(),
		blueLayout.ToString(),
		greenLayout.ToString(),
		magentaLayout.ToString(),
		cyanLayout.ToString(),
	}
	layoutModalViews = slices.Concat([]string{layoutModal}, layoutOptions)

	saveOptions = []string{
		saveOption.ToString(),
		cancelOption.ToString(),
	}
	saveModalViews = slices.Concat([]string{saveModal}, saveOptions)

	currentFile string
)

// Layout initialize the panel views and associates the key bindings to them.
func (ui *UI) Layout(g *gocui.Gui) error {
	diagrams, err := io.ListDiagrams(mainDir)
	if len(diagrams) > 0 && err == nil {
		cwd, err := filepath.Abs(filepath.Dir(""))
		if err != nil {
			return err
		}

		file := fmt.Sprintf("%s/%s/%s", cwd, mainDir, diagrams[0])
		ui.defaultContent, err = io.ReadFile(file)
		if err != nil {
			log.Fatalf("error loading the file content: %v", err)
		}
	}

	panelViews = map[string]panelProperties{
		logoPanel: {
			title:    " Info ",
			text:     version.DrawLogo(),
			x1:       0.0,
			y1:       0.0,
			x2:       0.35,
			y2:       0.20,
			editable: true,
			cursor:   false,
		},
		diagramsPanel: {
			title:    " Saved Diagrams ",
			text:     "",
			x1:       0.0,
			y1:       0.20,
			x2:       0.35,
			y2:       0.90,
			editable: true,
			cursor:   false,
		},
		logPanel: {
			title:    " Console ",
			text:     "",
			x1:       0.0,
			y1:       0.90,
			x2:       0.35,
			y2:       1.0,
			editable: true,
			cursor:   false,
		},
		editorPanel: {
			title:    " Editor ",
			text:     ui.defaultContent,
			x1:       0.35,
			y1:       0.0,
			x2:       1.0,
			y2:       1.0,
			editable: true,
			cursor:   true,
		},
	}

	modalViews = map[string]panelProperties{
		helpModal: {
			title:    " Key Shortcuts ",
			text:     "",
			editable: false,
		},
		saveModal: {
			title:    " Save diagram ",
			text:     ".txt",
			editable: true,
		},
		layoutModal: {
			title:    " Layout color ",
			editable: true,
		},
		progressModal: {
			title:    "",
			text:     " Generating diagram... ",
			editable: false,
		},
	}

	initPanel := func(g *gocui.Gui, v *gocui.View) error {
		// Disable panel views selection with mouse in case the modal is activated
		if ui.currentModal == "" {
			cx, cy := v.Cursor()
			line, err := v.Line(cy)
			if err != nil {
				ui.cursors.Restore(v)
				ui.setPanelView(v.Name())
			}

			if cx > len(line) {
				v.SetCursor(ui.cursors.Get(v.Name()))
				ui.cursors.Set(v.Name(), ui.getViewRowCount(v, cy), cy)
			}
			ui.currentView = ui.findViewByName(v.Name())
			ui.setPanelView(v.Name())
			view := panelViews[v.Name()]
			ui.gui.Cursor = view.cursor
		}

		// Refresh the diagram panel with the new diagram content
		cv := ui.gui.CurrentView()
		if cv.Name() == diagramsPanel && len(cv.Buffer()) > 0 {
			if err := ui.loadContent(editorPanel); err != nil {
				return fmt.Errorf("panel error: %w", err)
			}
		}

		return nil
	}

	for _, view := range mainViews {
		if err := g.SetKeybinding(view, gocui.MouseLeft, gocui.ModNone, initPanel); err != nil {
			return err
		}
		if err := g.SetKeybinding(view, gocui.MouseRelease, gocui.ModNone, initPanel); err != nil {
			return err
		}

		if _, err := ui.initPanelView(view); err != nil {
			return err
		}
	}

	// Activate the first panel on first run
	if v := ui.gui.CurrentView(); v == nil {
		_, err := ui.gui.SetCurrentView(editorPanel)
		if err != gocui.ErrUnknownView {
			return err
		}
	}

	return nil
}

// scrollUp moves the cursor up to the previous buffer line.
func (ui *UI) scrollUp(g *gocui.Gui, v *gocui.View) error {
	ox, oy := v.Origin()
	_, cy := v.Cursor()
	if err := v.SetCursor(ox, cy-1); err == nil && oy > 0 {
		ui.cursors.Set(v.Name(), ox, 0)
		if err := v.SetOrigin(ox, oy-1); err != nil {
			return err
		}
	}

	return nil
}

// scrollDown moves the cursor down to the next buffer line.
func (ui *UI) scrollDown(g *gocui.Gui, v *gocui.View) error {
	totalRows := ui.getTotalRows(v)
	_, maxY := v.Size()

	ox, oy := v.Origin()
	_, cy := v.Cursor()
	if err := v.SetCursor(ox, cy+1); err == nil {
		ui.cursors.Set(v.Name(), ox, 0)
		if totalRows > maxY {
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
	}

	return nil
}

// initPanelView initializes the panel view.
func (ui *UI) initPanelView(name string) (*gocui.View, error) {
	maxX, maxY := ui.gui.Size()

	p := panelViews[name]

	x1 := int(p.x1 * float64(maxX))
	y1 := int(p.y1 * float64(maxY))
	x2 := int(p.x2*float64(maxX)) - 1
	y2 := int(p.y2*float64(maxY)) - 1

	return ui.createPanelView(name, x1, y1, x2, y2)
}

// createPanelView creates the panel view.
func (ui *UI) createPanelView(name string, x1, y1, x2, y2 int) (*gocui.View, error) {
	v, err := ui.gui.SetView(name, x1, y1, x2, y2)
	if err != gocui.ErrUnknownView {
		return nil, err
	}

	p := panelViews[name]
	v.Title = p.title
	v.Editable = p.editable

	if err := ui.writeContent(name, p.text); err != nil {
		return nil, err
	}

	switch name {
	case editorPanel:
		v.Highlight = false
		v.Autoscroll = false
		v.Wrap = true
		v.Editor = NewEditor(ui, nil)
	case diagramsPanel:
		v.Highlight = true
		v.Autoscroll = false
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		v.Editor = NewEditor(ui, &staticViewEditor{})

		// TODO: workaround to disable scrolling.
		_ = ui.gui.SetKeybinding(v.Name(), gocui.MouseWheelUp, gocui.ModNone,
			func(g *gocui.Gui, v *gocui.View) error {
				return v.SetCursor(0, 0)
			},
		)
		_ = ui.gui.SetKeybinding(v.Name(), gocui.MouseWheelDown, gocui.ModNone,
			func(g *gocui.Gui, v *gocui.View) error {
				return v.SetCursor(0, 0)
			},
		)

		// Update diagrams directory list
		if err := ui.updateDiagramList(name); err != nil {
			return nil, err
		}
	case logPanel:
		v.Wrap = true
		v.Editor = NewEditor(ui, &staticViewEditor{})
	default:
		v.Editor = NewEditor(ui, &staticViewEditor{})
	}

	return v, nil
}

// createModalView creates the modal view.
func (ui *UI) createModalView(name string, x1, y1, x2, y2 int) (*gocui.View, error) {
	v, err := ui.gui.SetView(name, x1, y1, x2, y2)
	if err != gocui.ErrUnknownView {
		return nil, err
	}
	m := modalViews[name]

	v.Title = m.title
	v.Editable = m.editable

	if err := ui.writeContent(name, m.text); err != nil {
		return nil, err
	}

	return v, nil
}

// activatePanelView activates the view defined by id.
func (ui *UI) activatePanelView(id int) error {
	if err := ui.setPanelView(mainViews[id]); err != nil {
		return err
	}
	v := panelViews[mainViews[id]]
	ui.gui.Cursor = v.cursor
	ui.currentView = id

	return nil
}

// setPanelView activates the panel view.
func (ui *UI) setPanelView(name string) error {
	if err := ui.closeModal(ui.currentModal); err != nil {
		return err
	}
	// Save cursor position before switch view
	view := ui.gui.CurrentView()
	x, y := view.Cursor()
	ui.cursors.Set(view.Name(), x, y)

	if _, err := ui.gui.SetCurrentView(name); err != nil {
		if err == gocui.ErrUnknownView {
			return nil
		}
		return err
	}
	return nil
}

// writeContent writes the content into the specific view and set the cursor to the buffer end.
func (ui *UI) writeContent(name, text string) error {
	v, err := ui.gui.View(name)
	if err != nil {
		return err
	}
	v.Clear()
	if err := v.SetCursor(len(text), 0); err != nil {
		if err := v.SetOrigin(0, 0); err != nil {
			return err
		}
	}
	ui.cursors.Set(name, len(text), 0)
	fmt.Fprint(v, text)

	return nil
}

// findViewByName find the view defined by name and returns the view index.
func (ui *UI) findViewByName(name string) int {
	var viewId = -1
	for idx, v := range mainViews {
		if v == name {
			viewId = idx
			break
		}
	}
	return viewId
}

// saveDiagram saves the diagram content.
func (ui *UI) saveDiagram(name string) error {
	v, err := ui.gui.View(name)
	if err != nil {
		return err
	}

	if len(v.ViewBuffer()) == 0 {
		ui.consoleLog = errorEmpty
		if err := ui.log(ui.consoleLog, true); err != nil {
			return err
		}
	}
	if err := ui.showSaveModal(saveModal); err != nil {
		return fmt.Errorf("error opening the save diagram modal: %w", err)
	}

	return nil
}

// generateDiagram converts the ASCII to the hand-drawn diagram.
func (ui *UI) generateDiagram(name string) error {
	var output string

	if ui.logTimer != nil {
		ui.logTimer.Stop()
	}

	v, err := ui.gui.View(name)
	if err != nil {
		return err
	}

	if len(v.ViewBuffer()) == 0 {
		ui.consoleLog = errorEmpty
		if err := ui.log(ui.consoleLog, true); err != nil {
			return err
		}
	}

	content := strings.ReplaceAll(v.ViewBuffer(), "\n", "")
	if content == invalidContent {
		ui.consoleLog = invalidContent
		if err := ui.log(ui.consoleLog, true); err != nil {
			return err
		}
		return fmt.Errorf("invalid file format")
	}

	if currentFile == "" {
		output = "output.png"
	} else {
		output = strings.TrimSuffix(currentFile, ".txt")
		output = output + ".png"
	}

	start := time.Now()
	// Show progress
	if err := ui.showProgressModal(progressModal); err != nil {
		return fmt.Errorf("error on showing the progress modal: %w", err)
	}

	cwd, err := filepath.Abs(filepath.Dir(""))
	if err != nil {
		return err
	}

	filePath := cwd + "/output/"
	diagram := filePath + output

	// Create output directory in case it does not exists.
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err = os.Mkdir(filePath, os.ModePerm); err != nil {
			return fmt.Errorf("cannot create the output directory: %w", err)
		}
	}

	// Generate the hand-drawn diagram.
	err = canvas.DrawDiagram(v.Buffer(), diagram, ui.fontPath)
	if err != nil {
		_ = ui.closeModal(progressModal)
		return fmt.Errorf("failed generating diagram: %w", err)
	}

	ui.modalTimer = time.AfterFunc(time.Since(start), func() {
		ui.gui.Update(func(*gocui.Gui) error {
			ui.activeModalView = 0 // reset modal elements counter to 0
			if err := ui.closeModal(progressModal); err != nil {
				return err
			}

			if err := ui.showDiagram(diagram); err != nil {
				return fmt.Errorf("error previewing the diagram: %w", err)
			}

			if ui.modalTimer != nil {
				ui.modalTimer.Stop()
			}

			return ui.log("The ASCII diagram has been successfully converted to hand drawn diagram.", false)
		})
	})

	defer func() {
		// Hide log message after 4 seconds
		ui.logTimer = time.AfterFunc(4*time.Second, func() {
			ui.gui.Update(func(*gocui.Gui) error {
				return ui.clearLog()
			})
		})
	}()

	return nil
}

func (ui *UI) showDiagram(diagram string) error {
	f, err := os.Open(diagram)
	if err != nil {
		return fmt.Errorf("failed opening the image %q: %w", diagram, err)
	}

	srcImg, _, err := image.Decode(f)
	if err != nil {
		return fmt.Errorf("failed to decode the image %q: %w", diagram, err)
	}

	// Lunch Gio GUI thread.
	gui := gui.NewGUI()
	if err := gui.Draw(srcImg); err != nil {
		return fmt.Errorf("error drawing the diagram: %w", err)
	}

	return nil
}

// updateView updates the view content.
func (ui *UI) updateView(v *gocui.View, buffer string) error {
	if err := ui.writeContent(v.Name(), buffer); err != nil {
		return err
	}

	return nil
}

// loadContent load the content of the selected file into the editor panel.
func (ui *UI) loadContent(name string) error {
	v, err := ui.gui.View(name)
	if err != nil {
		return err
	}

	cv, err := ui.gui.View(diagramsPanel)
	if err != nil {
		return err
	}
	_, cy := cv.Cursor()
	cwd, err := filepath.Abs(filepath.Dir(""))
	if err != nil {
		return err
	}

	currentFile = ui.getViewRow(cv, cy)
	file := fmt.Sprintf("%s/%s/%s", cwd, mainDir, currentFile)
	content, err := io.ReadFile(file)
	if err != nil {
		return err
	}

	buffer := string(content)
	if !utf8.ValidString(buffer) {
		buffer = invalidContent
	}

	return ui.updateView(v, buffer)
}

// updateDiagramList updates the diagram panel content.
func (ui *UI) updateDiagramList(name string) error {
	v, err := ui.gui.View(name)
	if err != nil {
		return err
	}
	v.Clear()
	diagrams, err := io.ListDiagrams(mainDir)
	if err != nil {
		return err
	}

	for idx, diagram := range diagrams {
		if idx < len(diagrams)-1 {
			fmt.Fprint(v, diagram+"\n")
		} else {
			fmt.Fprint(v, diagram)
		}
		v.SetCursor(len(diagram), 0)
		ui.cursors.Set(name, len(diagram), 0)
	}
	return nil
}

// nextView activate the next panel.
func (ui *UI) nextView(wrap bool) error {
	var index int
	index = ui.currentView + 1
	if index > len(mainViews)-1 {
		if wrap {
			index = 0
		} else {
			return nil
		}
	}
	ui.currentView = index % len(mainViews)
	return ui.activatePanelView(ui.currentView)
}

// prevView activate the previous panel.
func (ui *UI) prevView(wrap bool) error {
	var index int
	index = ui.currentView - 1
	if index < 0 {
		if wrap {
			index = len(mainViews) - 1
		} else {
			return nil
		}
	}
	ui.currentView = index % len(mainViews)
	return ui.activatePanelView(ui.currentView)
}

// ClearView clears the panel view.
func (ui *UI) ClearView(name string) {
	v, _ := ui.gui.View(name)
	v.Clear()
}

// DeleteView deletes the current view.
func (ui *UI) DeleteView(name string) error {
	v, _ := ui.gui.View(name)
	return ui.gui.DeleteView(v.Name())
}
