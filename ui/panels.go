package ui

import (
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/esimov/diagram/canvas"
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

const (
	// Main panels
	logoPanel     = "logo_panel"
	diagramsPanel = "diagrams_panel"
	editorPanel   = "editor_panel"
	logPanel      = "log_panel"

	// Modals
	helpModal     = "help_modal"
	saveModal     = "save_modal"
	progressModal = "progress_modal"

	// Log messages
	errorEmpty     = "The editor should not be empty!"
	invalidContent = "Cannot display the file content!"

	mainDir = "/diagrams"
)

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
	modalElements = []string{"save_modal", "save", "cancel"}
	currentFile   string

	mouseX, mouseY int
)

// Layout initialize the panel views and associates the key bindings to them.
func (ui *UI) Layout(g *gocui.Gui) error {
	var defaultContent []byte
	diagrams, err := io.ListDiagrams(mainDir)
	if err != nil {
		return fmt.Errorf("error listing the existing diagrams: %w", err)
	}

	if len(diagrams) > 0 {
		defaultContent, err = io.ReadFile(diagrams[0])
		if err != nil {
			return fmt.Errorf("error loading the file content: %w", err)
		}
	} else {
		defaultContent, err = io.ReadFile("sample.txt")
		if err != nil {
			return fmt.Errorf("error loading the sample file content: %w", err)
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
			y2:       0.85,
			editable: true,
			cursor:   false,
		},
		logPanel: {
			title:    " Console ",
			text:     "",
			x1:       0.0,
			y1:       0.85,
			x2:       0.35,
			y2:       1.0,
			editable: true,
			cursor:   false,
		},
		editorPanel: {
			title:    " Editor ",
			text:     string(defaultContent),
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
			title:    "Key Shortcuts",
			text:     "",
			editable: false,
		},
		saveModal: {
			title:    "Save diagram",
			text:     ".txt",
			editable: true,
		},
		progressModal: {
			title:    "",
			text:     " Generating diagram...",
			editable: false,
		},
	}

	initPanel := func(g *gocui.Gui, v *gocui.View) error {
		// Obtain the cursor position once a click is detected inside the editor panel.
		if v.Name() == editorPanel {
			cx, cy := v.Cursor()
			v.SetCursor(cx, cy)
			ui.cursors.Set(editorPanel, cx, cy)
			mouseX, mouseY = cx, cy
		}

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

// scrollUp moves the cursor up to the next buffer line.
func (ui *UI) scrollUp(g *gocui.Gui, v *gocui.View) error {
	mouseY--
	ox, oy := v.Origin()
	if err := v.SetCursor(mouseX, mouseY); err != nil && oy > 0 {
		ui.cursors.Set(v.Name(), mouseX, mouseY)
		if err := v.SetOrigin(ox, oy-1); err != nil {
			return err
		}
	}
	return nil
}

// scrollDown moves the cursor down to the next buffer line.
func (ui *UI) scrollDown(g *gocui.Gui, v *gocui.View) error {
	mouseY++
	if err := v.SetCursor(mouseX, mouseY); err != nil {
		ui.cursors.Set(v.Name(), mouseX, mouseY)
		ox, oy := v.Origin()
		if err := v.SetOrigin(ox, oy+1); err != nil {
			return err
		}
	}
	return nil
}

// toggleHelpModal show or hide the help modal.
func (ui *UI) toggleHelpModal(content string) error {
	if err := ui.closeOpenedModals(modalElements); err != nil {
		return err
	}
	panelHeight := strings.Count(content, "\n")
	if ui.currentModal == helpModal {
		ui.gui.DeleteKeybinding("", gocui.MouseLeft, gocui.ModNone)
		ui.gui.DeleteKeybinding("", gocui.MouseRelease, gocui.ModNone)

		// Stop modal timer from firing in case the modal was closed manually.
		// This is needed to prevent the modal being closed before the predefined delay.
		if ui.modalTimer != nil {
			ui.modalTimer.Stop()
		}
		return ui.closeModal(ui.currentModal)
	}
	v, err := ui.openModal(helpModal, 45, panelHeight, true)
	if err != nil {
		return err
	}
	ui.gui.Cursor = false
	v.Editor = newEditor(ui, &staticViewEditor{})

	fmt.Fprint(v, content)
	return nil
}

// openModal creates and opens the modal window.
// If "autoHide" parameter is true, the modal will be automatically closed after certain seconds.
func (ui *UI) openModal(name string, w, h int, autoHide bool) (*gocui.View, error) {
	v, err := ui.createModal(name, w, h)
	if err != nil {
		return nil, err
	}

	if err := ui.setPanelView(name); err != nil {
		return nil, err
	}
	ui.currentModal = name

	if autoHide {
		// Close the modal automatically after 5 seconds
		ui.modalTimer = time.AfterFunc(5*time.Second, func() {
			ui.gui.Update(func(*gocui.Gui) error {
				if err := ui.closeModal(name); err != nil {
					return err
				}
				return nil
			})
		})
	}
	return v, nil
}

// closeModal closes the modal window and restores the focus to the last accessed panel view.
func (ui *UI) closeModal(modals ...string) error {
	for _, name := range modals {
		if _, err := ui.gui.View(name); err != nil {
			if err == gocui.ErrUnknownView {
				return nil
			}
			return err
		}
		ui.gui.DeleteView(name)
		ui.gui.DeleteKeybindings(name)
		ui.gui.Cursor = true
		ui.currentModal = ""
	}
	return ui.activatePanelView(ui.currentView)
}

// createModal initializes and creates the modal view.
func (ui *UI) createModal(name string, w, h int) (*gocui.View, error) {
	width, height := ui.gui.Size()
	x1, y1 := width/2-w/2, math.Ceil(float64(height/2-h/2-1))
	x2, y2 := width/2+w/2, math.Ceil(float64(height/2+h/2+1))

	return ui.createModalView(name, x1, int(y1), x2, int(y2))
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
		v.Autoscroll = true
		v.Wrap = true
		v.Editor = newEditor(ui, nil)
	case diagramsPanel:
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		v.Editor = newEditor(ui, &staticViewEditor{})

		// Update diagrams directory list
		if err := ui.updateDiagramList(name); err != nil {
			return nil, err
		}
	default:
		v.Editor = newEditor(ui, &staticViewEditor{})
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
	return ui.showSaveModal(saveModal)
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
		return err
	}

	ui.modalTimer = time.AfterFunc(time.Since(start), func() {
		ui.gui.Update(func(*gocui.Gui) error {
			ui.nextItem = 0 // reset modal elements counter to 0
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
	if err := ui.gioGui.Draw(srcImg); err != nil {
		return fmt.Errorf("error drawing the diagram: %w", err)
	}

	return nil
}

// showSaveModal show the save modal.
func (ui *UI) showSaveModal(name string) error {
	var saveBtn, cancelBtn *gocui.View

	if err := ui.closeModal(ui.currentModal); err != nil {
		return err
	}

	modal, err := ui.openModal(name, 40, 4, false)
	if err != nil {
		return err
	}

	if ui.modalTimer != nil {
		ui.modalTimer.Stop()
	}

	ui.gui.Cursor = true
	modal.Editor = newEditor(ui, &modalSaveEditor{30})
	modal.SetCursor(0, 0)

	ui.gui.DeleteKeybinding("", gocui.MouseLeft, gocui.ModNone)
	ui.gui.DeleteKeybinding("", gocui.MouseRelease, gocui.ModNone)

	// Close event handler
	onClose := func(*gocui.Gui, *gocui.View) error {
		ui.nextItem = 0 // reset modal elements counter to 0
		if err := ui.closeOpenedModals(modalElements); err != nil {
			return err
		}
		return nil
	}

	// Save event handler
	onSave := func(*gocui.Gui, *gocui.View) error {
		if ui.logTimer != nil {
			ui.logTimer.Stop()
		}
		diagram, _ := ui.gui.View(editorPanel)
		v := modalViews[name]

		ui.nextItem = 0 // reset modal elements counter to 0

		// Check if the file name contains only letters, numbers and underscores.
		buffer := strings.TrimSpace(strings.Replace(modal.ViewBuffer(), v.text, "", -1))
		re := regexp.MustCompile("^[a-zA-Z0-9_]*$")
		res := re.MatchString(buffer)

		if len(diagram.ViewBuffer()) == 0 {
			ui.log("Missing content on diagram save!", true)
			return nil
		}

		if len(strings.TrimSpace(modal.Buffer())) <= len(v.text) {
			ui.log("File name should not be empty!", true)
		} else if res {
			file := buffer + v.text
			f, err := io.SaveFile(file, mainDir, diagram.ViewBuffer())
			if err != nil {
				return fmt.Errorf("error saving the file: %w", err)
			}
			defer f.Close()

			ui.log(fmt.Sprintf("The file has been saved as: %s", file), false)
		} else {
			ui.log("Error saving the diagram. The file name should contain only letters, numbers and underscores!", true)
		}

		if err := ui.closeOpenedModals(modalElements); err != nil {
			return fmt.Errorf("could not close opened modal: %w", err)
		}

		// Update diagrams directory list
		err := ui.updateDiagramList(diagramsPanel)
		if err != nil {
			return fmt.Errorf("could not update diagram list: %w", err)
		}

		// Hide log message after 4 seconds
		ui.logTimer = time.AfterFunc(4*time.Second, func() {
			ui.gui.Update(func(*gocui.Gui) error {
				return ui.clearLog()
			})
		})

		return nil
	}

	// Tab event handler
	onNext := func(*gocui.Gui, *gocui.View) error {
		var pv *gocui.View

		if err := ui.nextElement(modalElements); err != nil {
			return err
		}
		if (ui.nextItem - 1) > 0 {
			pv, _ = ui.gui.View(modalElements[ui.nextItem-1])
		} else {
			pv, _ = ui.gui.View(modalElements[len(modalElements)-1])
		}
		pv.Highlight = false
		if ui.nextItem == 0 {
			ui.gui.Cursor = true
		}
		return nil
	}

	// Get modal with and height
	sw, sh := ui.gui.Size()
	mw, _ := modal.Size()

	saveBtn, err = ui.createButtonWidget("save", sw/2-mw/2, sh/2, "Save", nil)
	if err != nil {
		return err
	}

	if saveBtn != nil {
		saveBtnSize, _ := saveBtn.Size()
		//Calculate the current modal button position relative to the previous button.
		cancelBtn, err = ui.createButtonWidget("cancel", (sw/2-mw/2)+saveBtnSize+4, sh/2, "Cancel", nil)
		if err != nil {
			return err
		}
		if err := ui.gui.SetKeybinding(saveBtn.Name(), gocui.KeyEnter, gocui.ModNone, onSave); err != nil {
			return err
		}
		if err := ui.gui.SetKeybinding(cancelBtn.Name(), gocui.KeyEnter, gocui.ModNone, onClose); err != nil {
			return err
		}
	}

	keys := []gocui.Key{gocui.KeyCtrlS, gocui.KeyEnter}
	for _, k := range keys {
		if err := ui.gui.SetKeybinding(name, k, gocui.ModNone, onSave); err != nil {
			return err
		}
	}
	// Associate the close modal key binding to each modal element.
	for _, view := range modalElements {
		if err := ui.gui.SetKeybinding(view, gocui.KeyCtrlX, gocui.ModNone, onClose); err != nil {
			return err
		}
		if err := ui.gui.SetKeybinding(view, gocui.KeyTab, gocui.ModNone, onNext); err != nil {
			return err
		}
	}

	// Hide log message after 4 seconds
	ui.logTimer = time.AfterFunc(4*time.Second, func() {
		ui.gui.Update(func(*gocui.Gui) error {
			return ui.clearLog()
		})
	})

	return nil
}

// showProgressModal shows the progress modal.
func (ui *UI) showProgressModal(name string) error {
	if err := ui.closeModal(ui.currentModal); err != nil {
		return err
	}
	_, err := ui.openModal(name, 40, 1, false)
	if err != nil {
		return err
	}
	if ui.modalTimer != nil {
		ui.modalTimer.Stop()
	}

	ui.gui.Cursor = false
	ui.gui.DeleteKeybinding("", gocui.MouseLeft, gocui.ModNone)
	ui.gui.DeleteKeybinding("", gocui.MouseRelease, gocui.ModNone)

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

// closeOpenedModals closes all the opened modal elements.
func (ui *UI) closeOpenedModals(views []string) error {
	for _, v := range views {
		if view, _ := ui.gui.View(v); view != nil {
			_ = ui.closeModal(view.Name())
		}
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
