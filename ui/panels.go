package ui

import (
	"github.com/jroimartin/gocui"
	"github.com/esimov/diagram/version"
	"fmt"
	"strings"
)

type panelProperties struct {
	title    string
	text     string
	x1       float64
	y1       float64
	x2       float64
	y2       float64
	editable bool
	editor	 *UI
	modal    bool
}

const (
	LOGO_PANEL 		= "logo"
	SAVED_DIAGRAMS_PANEL 	= "saved_diagrams"
	ACTIONS_PANEL		= "actions"
	DIAGRAM_PANEL		= "diagram"
	PROGRESS_PANEL		= "progress"
)

var panelViews = map[string]panelProperties{
	LOGO_PANEL: {
		title:    "Diagram",
		text:     version.DrawLogo(),
		x1:       0.0,
		y1:       0.0,
		x2:       0.4,
		y2:       0.25,
		editable: true,
		modal:    false,
	},
	SAVED_DIAGRAMS_PANEL: {
		title:    "Saved Diagrams",
		text:     "",
		x1:       0.0,
		y1:       0.25,
		x2:       0.4,
		y2:       0.75,
		editable: true,
		modal:    false,
	},
	ACTIONS_PANEL: {
		title:    "Actions",
		text:     "",
		x1:       0.0,
		y1:       0.75,
		x2:       0.4,
		y2:       1.0,
		editable: true,
		modal:    false,
	},
	DIAGRAM_PANEL: {
		title:    "Editor",
		text:     "",
		x1:       0.4,
		y1:       0.0,
		x2:       1.0,
		y2:       1.0,
		editable: true,
		modal:    false,
	},
	PROGRESS_PANEL: {
		title:    "Progress",
		text:     "",
		x1:       0.0,
		y1:       0.7,
		x2:       1,
		y2:       0.8,
		editable: false,
		modal:    false,
	},
}

var mainViews = []string{
	LOGO_PANEL,
	SAVED_DIAGRAMS_PANEL,
	ACTIONS_PANEL,
	DIAGRAM_PANEL,
}

// Layout sets up the panel views
func (ui *UI) Layout(g *gocui.Gui) error {
	for _, view := range mainViews {
		fn := func(g *gocui.Gui, v *gocui.View) error {
			cx, cy := v.Cursor()
			line, err := v.Line(cy)
			if err != nil {
				ui.cursors.Restore(v)
				ui.setPanelView(v.Name())
			}

			if cx > len(line) {
				v.SetCursor(ui.getViewRowCount(v, cy), cy)
				ui.cursors.Set(v.Name(), ui.getViewRowCount(v, cy), cy)
			} else {
				if v.Name() != DIAGRAM_PANEL {
					v.SetCursor(0, 0)
				}
			}
			ui.currentView = ui.findViewByName(v.Name())
			ui.setPanelView(v.Name())
			return nil
		}

		if err := g.SetKeybinding(view, gocui.MouseLeft, gocui.ModNone, fn); err != nil {
			return err
		}

		if err := g.SetKeybinding(view, gocui.MouseRelease, gocui.ModNone, fn); err != nil {
			return err
		}

		if err := ui.initPanelView(view); err != nil {
			return err
		}
	}

	if err := g.SetKeybinding(DIAGRAM_PANEL, gocui.MouseWheelDown, gocui.ModNone, ui.scrollDown); err != nil {
		return err
	}

	// Activate the first panel on first run.
	if v := ui.gui.CurrentView(); v == nil {
		_, err := ui.gui.SetCurrentView(LOGO_PANEL)
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	return nil
}

func (ui *UI) scrollDown(g *gocui.Gui, v *gocui.View) error {
	maxY := strings.Count(v.Buffer(), "\n")
	if maxY < 1 {
		v.SetCursor(0, 0)
	}
	return nil
}

// initPanelView create the panel view
func (ui *UI) initPanelView(name string) error {
	maxX, maxY := ui.gui.Size()

	p := panelViews[name]
	if p.modal {
		// Don't init modals
		return nil
	}

	x1 := int(p.x1 * float64(maxX))
	y1 := int(p.y1 * float64(maxY))
	x2 := int(p.x2 * float64(maxX)) - 1
	y2 := int(p.y2 * float64(maxY)) - 1

	return  ui.createPanelView(name, x1, y1, x2, y2)
}

// createPanelView create the panel view
func (ui *UI) createPanelView(name string, x1, y1, x2, y2 int) error {
	if v, err := ui.gui.SetView(name, x1, y1, x2, y2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		p := panelViews[name]
		v.Title = p.title
		switch name {
		case DIAGRAM_PANEL:
			v.Autoscroll = true
			v.Editor = newEditor(ui, nil)
		default:
			v.Editor = newEditor(ui, &staticViewEditor{})
		}
		v.Editable = p.editable
		if err := ui.writeContent(name, p.text); err != nil {
			return err
		}
	}
	return nil
}

// activatePanelView set the focus to the view specified by id
func (ui *UI) activatePanelView(id int) error {
	if err := ui.setPanelView(mainViews[id]); err != nil {
		return err
	}
	ui.currentView = id

	return nil
}

// setPanelView activate the panel view
func (ui *UI) setPanelView(name string) error {
	// Save cursor position before switch view
	view := ui.gui.CurrentView()
	x, y := view.Cursor()
	ui.cursors.Set(view.Name(), x, y)

	if _, err := ui.gui.SetCurrentView(name); err != nil {
		return err
	}
	return nil
}

// writeContent writes string to view
func (ui *UI) writeContent(name, text string) error {
	v, err := ui.gui.View(name)
	if err != nil {
		return err
	}
	v.Clear()
	fmt.Fprintf(v, text)
	v.SetCursor(len(text), 0)
	ui.cursors.Set(name, len(text), 0)

	return nil
}

func (ui *UI) findViewByName(name string) int {
	var viewId int = -1
	for idx, v := range mainViews {
		if v == name {
			viewId = idx
			break
		}
	}
	return viewId
}

// ClearView clear the panel view
func (ui *UI) ClearView(name string) {
	v, _ := ui.gui.View(name)
	v.Clear()
}

// DeleteView delete the current view
func (ui *UI) DeleteView(name string) {
	v, _ := ui.gui.View(name)
	ui.gui.DeleteView(v.Name())
}

// NextView activate the next panel
func (ui *UI) NextView(wrap bool) error {
	var index int
	index = ui.currentView + 1
	if index > len(mainViews) - 1 {
		if wrap {
			index = 0
		} else {
			return nil
		}
	}
	ui.currentView = index % len(mainViews)
	return ui.activatePanelView(ui.currentView)
}

// PrevView activate the previous panel
func (ui *UI) PrevView(wrap bool) error {
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