package ui

import (
	"github.com/jroimartin/gocui"
	"strings"
)

type editor struct {
	ui            *UI
	editor        gocui.Editor
	backTabEscape bool
}

var cache []byte

// Create a new GUI editor
func newEditor(ui *UI, handler gocui.Editor) *editor {
	if handler == nil {
		handler = gocui.DefaultEditor
	}
	return &editor{ui, handler, true}
}

// The main editor for the editable views
func (e *editor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	if ch == '[' && mod == gocui.ModAlt {
		e.backTabEscape = true
		return
	}

	if e.backTabEscape {
		if ch == 'Z' {
			e.ui.PrevView(true)
			e.backTabEscape = false
			return
		}
	}
	// Prevent infinite scrolling
	if (key == gocui.KeyArrowDown || key == gocui.KeyArrowRight) && mod == gocui.ModNone {
		_, cy := v.Cursor()
		if _, err := v.Line(cy); err != nil {
			return
		}
	}

	switch key {
	// Disable line wrapping (right arrow key at line end wraps too)
	case gocui.KeyArrowRight:
		cx, cy := v.Cursor()
		// Get the total number of rows in the current view
		maxY := strings.Count(v.Buffer(), "\n")
		// Check if the cursor is on the last row of the current view
		if cy == maxY - 1 {
			// Prevent line wrapping on last row
			if cx >= e.ui.getViewRowCount(v, cy) {
				return
			}
		}
	case gocui.KeyHome:
		_, cy := v.Cursor()
		v.SetCursor(0, cy)
	case gocui.KeyEnd:
		_, cy := v.Cursor()
		maxX := e.ui.getViewRowCount(v, cy)
		v.SetCursor(maxX, cy)
	case gocui.KeyPgup:
		vx, vy := v.Origin()
		if err := v.SetCursor(0, 0); err != nil && vy > 0 {
			if err := v.SetOrigin(vx, 0); err != nil {
				return
			}
		}
	case gocui.KeyPgdn:
		maxX := e.ui.getViewLastRowCount(v)
		maxY := strings.Count(v.ViewBuffer(), "\n")-1
		v.SetCursor(maxX, maxY)
	case gocui.KeyCtrlX:
		if v.Name() == DIAGRAM_PANEL {
			cache = []byte(v.ViewBuffer())
			e.ui.ClearView(v.Name())
			v.SetCursor(0, 0)
		}
	case gocui.KeyCtrlZ:
		if len(cache) > 0 {
			v.Write(cache)
			cache = []byte{}
		}
	}
	e.editor.Edit(v, key, ch, mod)
}

// Editor for static (non-editable) views
type staticViewEditor editor

func (e *staticViewEditor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	_, y := v.Cursor()
	maxY := strings.Count(v.Buffer(), "\n")
	switch {
	case key == gocui.KeyArrowDown:
		if y < maxY {
			v.MoveCursor(0, 1, true)
		}
	case key == gocui.KeyArrowUp:
		v.MoveCursor(0, -1, false)
	case key == gocui.KeyArrowLeft:
		v.MoveCursor(-1, 0, false)
	case key == gocui.KeyArrowRight:
		v.MoveCursor(1, 0, false)
	}
}

// Save modal editor init struct
type modalSaveEditor struct {
	maxWidth int
}

// Save modal editor
func (e *modalSaveEditor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	x, _ := v.Cursor()
	switch {
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	case key == gocui.KeyDelete:
		v.EditDelete(false)
	case key == gocui.KeyArrowLeft:
		v.MoveCursor(-1, 0, false)
	case key == gocui.KeyArrowRight:
		if x < len(v.Buffer())-1 {
			v.MoveCursor(1, 0, false)
		}
	case key == gocui.KeyArrowDown:
		return
	case key == gocui.KeyEnter:
		return
	default:
		if x > e.maxWidth {
			return
		}
	}
	gocui.DefaultEditor.Edit(v, key, ch, mod)
}

// getViewRow returns the row content defined by "y"
func (ui *UI) getViewRow(v *gocui.View, y int) []string {
	var row string
	rows := []string{}
	buffer := v.ViewBuffer()
	for _, char := range []byte(buffer) {
		if string(char) == "\n" {
			rows = append(rows, row)
			row = ""
		} else {
			row = row + string(char)
		}
	}
	if len(rows) > 0 && (y > -1 && y < len(rows)) {
		return []string{rows[y]}
	}
	return []string{""}
}

// getViewLastRow returns the last row content
func (ui *UI) getViewLastRow(v *gocui.View) []string {
	var row string
	rows := []string{}
	buffer := v.ViewBuffer()

	for _, char := range []byte(buffer) {
		if string(char) == "\n" {
			rows = append(rows, row)
			row = ""
		} else {
			row = row + string(char)
		}
	}

	if len(rows) > 0 {
		// Traverse up the string slice and remove all the trailing spaces from the end of the text.
		fn := func(rows []string) int{
			var idx int = 1
			for {
				current := string(rows[len(rows) - idx:][0])
				if current == "" {
					idx++
				} else {
					break
				}
			}
			return idx
		}
		index := fn(rows)
		return rows[len(rows)-index:]
	}
	return []string{""}
}

// getViewRowCount returns the number of characters in the row defined by "y"
func (ui *UI) getViewRowCount(v *gocui.View, y int) int {
	row := ui.getViewRow(v, y)
	return len(strings.Split(row[0], ""))
}

// getViewLastRowCount returns the number of characters in the last row
func (ui *UI) getViewLastRowCount(v *gocui.View) int {
	lastRow := ui.getViewLastRow(v)
	return len(strings.Split(lastRow[0], ""))
}

// getViewTotalRows returns the total number of rows of the current view
func (ui *UI) getViewTotalRows(v *gocui.View) int {
	var rows int
	buffer := v.ViewBuffer()

	for _, char := range []byte(buffer) {
		if string(char) == "\n" {
			rows++
		}
	}
	return rows
}

// getPartialViewBuffer returns the view buffer down until the row defined by "n"
func (ui *UI) getPartialViewBuffer(v *gocui.View, n int) string {
	var row string
	var idx int
	var newBuffer string

	rows := []string{}
	buffer := v.ViewBuffer()

	for _, char := range []byte(buffer) {
		if string(char) == "\n" {
			rows = append(rows, row)
			row = ""
			if idx > n {
				break
			}
			idx++
		} else {
			row = row + string(char)
		}
	}
	if idx < n {
		newBuffer = strings.Join(rows[:idx], "\n")
	} else {
		newBuffer = strings.Join(rows[:n], "\n")
	}
	return newBuffer
}