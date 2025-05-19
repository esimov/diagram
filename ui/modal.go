package ui

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/esimov/diagram/io"
	"github.com/jroimartin/gocui"
)

// showHelpModal show/hide the help modal.
func (ui *UI) showHelpModal(content string) error {
	modals := slices.Concat(saveModalViews, layoutModalViews)
	if err := ui.closeModals(modals...); err != nil {
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
	modal, err := ui.openModal(helpModal, 50, panelHeight, true)
	if err != nil {
		return err
	}

	ui.gui.Cursor = false
	modal.BgColor = gocui.ColorBlack
	modal.Editor = NewEditor(ui, &staticViewEditor{})

	fmt.Fprint(modal, content)

	return nil
}

// showProgressModal shows the progress modal.
func (ui *UI) showProgressModal(name string) error {
	modals := slices.Concat(saveModalViews, layoutModalViews)
	if err := ui.closeModals(modals...); err != nil {
		return err
	}

	modal, err := ui.openModal(name, 40, 1, false)
	if err != nil {
		return err
	}
	if ui.modalTimer != nil {
		ui.modalTimer.Stop()
	}
	modal.BgColor = gocui.ColorBlack
	ui.gui.Cursor = false

	ui.gui.DeleteKeybinding("", gocui.MouseLeft, gocui.ModNone)
	ui.gui.DeleteKeybinding("", gocui.MouseRelease, gocui.ModNone)

	return nil
}

// showSaveModal show the save modal.
func (ui *UI) showSaveModal(name string) error {
	var saveBtnWidget, cancelBtnWidget ComponentHandler

	modals := slices.Concat(saveModalViews, layoutModalViews)
	if err := ui.closeModals(modals...); err != nil {
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
	modal.BgColor = ui.selectedColor
	modal.Editor = NewEditor(ui, &modalViewEditor{30})
	_ = modal.SetCursor(0, 0)

	ui.gui.DeleteKeybinding("", gocui.MouseLeft, gocui.ModNone)
	ui.gui.DeleteKeybinding("", gocui.MouseRelease, gocui.ModNone)

	// Close event handler
	onClose := func(*gocui.Gui, *gocui.View) error {
		ui.nextItem = 0 // reset modal elements counter to 0
		if err := ui.closeModals(saveModalViews...); err != nil {
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

		var hasErrors bool
		if len(strings.TrimSpace(modal.Buffer())) <= len(v.text) {
			ui.log("File name should not be empty!", true)
			hasErrors = true
		} else if res {
			file := buffer + v.text
			f, err := io.SaveFile(file, mainDir, diagram.ViewBuffer())
			if err != nil {
				return fmt.Errorf("error saving the file: %w", err)
			}
			defer f.Close()

			ui.log(fmt.Sprintf("The diagram has been saved as %q into the %q folder.", file, mainDir), false)
		} else {
			ui.log("Error saving the diagram. The file name should contain only letters, numbers and underscores!", true)
			hasErrors = true
		}

		if !hasErrors {
			if err := ui.closeModals(saveModalViews...); err != nil {
				return fmt.Errorf("could not close opened modal: %w", err)
			}
		}

		// Update diagrams directory list
		err := ui.updateDiagramList(diagramsPanel)
		if err != nil {
			return fmt.Errorf("could not update diagram list: %w", err)
		}

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

	// Tab event handler
	onNext := func(*gocui.Gui, *gocui.View) error {
		var pv *gocui.View

		if saveBtnWidget != nil {
			if err := saveBtnWidget.NextElement(saveModalViews); err != nil {
				return err
			}
		}

		if (ui.nextItem - 1) > 0 {
			pv, _ = ui.gui.View(saveModalViews[ui.nextItem-1])
		} else {
			pv, _ = ui.gui.View(saveModalViews[len(saveModalViews)-1])
		}
		pv.Highlight = false
		if ui.nextItem == 0 {
			ui.gui.Cursor = true
		}
		return nil
	}

	onClick := func(*gocui.Gui, *gocui.View) error {
		_ = modal.SetOrigin(0, 0)
		_ = modal.SetCursor(0, 0)
		return nil
	}

	// Get modal with and height
	sw, sh := ui.gui.Size()
	mw, _ := modal.Size()

	saveBtnWidget, err = NewButton[*ButtonWidget](ui, saveButton, sw/2-mw/2, sh/2, len(saveButton)+1)
	if err != nil {
		return fmt.Errorf("failed to create a new button widget: %w", err)
	}

	saveBtn, err := saveBtnWidget.Draw()
	if err != nil {
		return fmt.Errorf("failed drawing the button: %w", err)
	}

	if saveBtn != nil {
		saveBtnSize, _ := saveBtn.Size()
		//Calculate the current modal button position relative to the previous button.
		cancelBtnWidget, err = NewButton[*ButtonWidget](ui, cancelButton, (sw/2-mw/2)+saveBtnSize+4, sh/2, len(cancelButton)+1)
		if err != nil {
			return fmt.Errorf("failed to create a new button widget: %w", err)
		}

		cancelBtn, err := cancelBtnWidget.Draw()
		if err != nil {
			return fmt.Errorf("failed drawing the button: %w", err)
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
	for _, view := range saveModalViews {
		if err := ui.gui.SetKeybinding(view, gocui.KeyEsc, gocui.ModNone, onClose); err != nil {
			return err
		}
		if err := ui.gui.SetKeybinding(view, gocui.KeyTab, gocui.ModNone, onNext); err != nil {
			return err
		}
		if err := ui.gui.SetKeybinding(view, gocui.MouseLeft, gocui.ModNone, onClick); err != nil {
			return err
		}
		if err := ui.gui.SetKeybinding(view, gocui.MouseRelease, gocui.ModNone, onClick); err != nil {
			return err
		}
	}

	return nil
}

// showProgressModal shows the progress modal.
func (ui *UI) showLayoutModal(name string) error {
	modals := slices.Concat(saveModalViews, layoutModalViews)
	if err := ui.closeModals(modals...); err != nil {
		return err
	}

	if ui.modalTimer != nil {
		ui.modalTimer.Stop()
	}

	modal, err := ui.openModal(name, 60, 3, false)
	if err != nil {
		return err
	}

	modal.BgColor = gocui.ColorBlack
	modal.Editor = NewEditor(ui, &staticViewEditor{})

	ui.gui.DeleteKeybinding("", gocui.MouseLeft, gocui.ModNone)
	ui.gui.DeleteKeybinding("", gocui.MouseRelease, gocui.ModNone)

	// Get modal with and height
	sw, sh := ui.gui.Size()
	mw, mh := modal.Size()

	radioBtnWidget, err := NewRadioButton[*RadioBtnWidget](ui, layoutModal, sw/2-mw/2, sh/2-mh/2)
	if err != nil {
		return fmt.Errorf("failed to create a new radio button widget: %w", err)
	}
	radioBtnWidget.AddOptions(layoutOptions...).Draw()

	// Close event handler
	onClose := func(*gocui.Gui, *gocui.View) error {
		if err := ui.closeModals(layoutModalViews...); err != nil {
			return err
		}
		ui.ApplyLayoutColor(ui.selectedColor)
		return nil
	}

	// Activate radio button on click.
	onClick := func(*gocui.Gui, *gocui.View) error {
		for idx, opt := range layoutOptions {
			v, _ := radioBtnWidget.gui.View(opt)
			cx, _ := v.Cursor()

			if cx > 0 {
				_ = v.SetCursor(0, 0)
				radioBtnWidget.unFocus()
				radioBtnWidget.activeLayout = idx
				radioBtnWidget.focus()
				continue
			}
		}

		return nil
	}

	// Associate the close modal key binding to each modal element.
	for _, view := range layoutModalViews {
		if err := ui.gui.SetKeybinding(view, gocui.KeyEsc, gocui.ModNone, onClose); err != nil {
			return err
		}
		if err := ui.gui.SetKeybinding(view, gocui.MouseRelease, gocui.ModNone, onClick); err != nil {
			return err
		}
	}

	return nil
}

// createModal initializes and creates the modal view.
func (ui *UI) createModal(name string, w, h int) (*gocui.View, error) {
	width, height := ui.gui.Size()
	x1, y1 := width/2-w/2, float64(height/2-h/2-1)
	x2, y2 := width/2+w/2, float64(height/2+h/2+1)

	return ui.createModalView(name, x1, int(y1), x2, int(y2))
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
		// Close the modal automatically after 10 seconds
		ui.modalTimer = time.AfterFunc(10*time.Second, func() {
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

// closeModals closes the modal elements provided as parameters.
func (ui *UI) closeModals(views ...string) error {
	for _, v := range views {
		if view, _ := ui.gui.View(v); view != nil {
			_ = ui.closeModal(view.Name())
		}
	}
	return nil
}
