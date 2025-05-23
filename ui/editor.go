package ui

import (
	"fmt"
	"strings"

	"github.com/jroimartin/gocui"
)

type editor struct {
	ui     *UI
	editor gocui.Editor
}

var cache []byte

// NewEditor creates a new GUI editor
func NewEditor(ui *UI, handler gocui.Editor) *editor {
	if handler == nil {
		handler = gocui.DefaultEditor
	}
	return &editor{ui, handler}
}

// Editor for editable views
func (e *editor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	if ch == '[' || ch == 'Z' && mod == gocui.ModAlt {
		e.ui.prevView(true)
		return
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
		if cy == maxY-1 {
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
		totalRows := e.ui.getTotalRows(v)
		cx, cy := v.Cursor()
		_, sy := v.Size()
		_, oy := v.Origin()
		offsetY := max(0, oy-(totalRows%sy))

		if err := v.SetOrigin(cx, offsetY); err == nil {
			if err := v.SetCursor(cx, cy+offsetY); err != nil {
				return
			}
		}
	case gocui.KeyPgdn:
		totalRows := e.ui.getTotalRows(v)
		cx, cy := v.Cursor()
		_, sy := v.Size()
		_, oy := v.Origin()
		offsetY := oy + (totalRows % sy)
		cy = min(offsetY+cy, totalRows)

		if offsetY < totalRows {
			if err := v.SetOrigin(cx, offsetY); err == nil {
				if err := v.SetCursor(cx, cy-offsetY); err != nil {
					return
				}
			}
		}
	case gocui.KeyCtrlX:
		if v.Name() == editorPanel {
			cache = []byte(v.ViewBuffer())
			e.ui.ClearView(v.Name())
			v.SetCursor(0, 0)
		}
	case gocui.KeyCtrlZ:
		if len(cache) > 0 {
			fmt.Fprintf(v, "%s", cache)
			cache = []byte{}
		}
	}
	e.editor.Edit(v, key, ch, mod)
}

// Editor for static (non-editable) views
type staticViewEditor editor

// Static view editor
func (e *staticViewEditor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	_, y := v.Cursor()
	maxY := strings.Count(v.Buffer(), "\n")
	switch key {
	case gocui.KeyArrowDown:
		if y < maxY {
			v.MoveCursor(0, 1, true)
		}
	case gocui.KeyArrowUp:
		v.MoveCursor(0, -1, false)
	case gocui.KeyArrowLeft:
		v.MoveCursor(-1, 0, false)
	case gocui.KeyArrowRight:
		v.MoveCursor(1, 0, false)
	}
}

type modalViewEditor struct {
	maxWidth int
}

// Save modal editor
func (e *modalViewEditor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	x, _ := v.Cursor()
	switch {
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
		return
	case key == gocui.KeyDelete:
		v.EditDelete(false)
		return
	case key == gocui.KeyArrowLeft:
		v.MoveCursor(-1, 0, false)
		return
	case key == gocui.KeyArrowRight:
		if x < len(v.Buffer())-1 {
			v.MoveCursor(1, 0, false)
		}
		return
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
func (ui *UI) getViewRow(v *gocui.View, y int) string {
	rows := []string{}
	buffer := v.ViewBuffer()
	sb := new(strings.Builder)

	for _, char := range buffer {
		if char == '\n' {
			rows = append(rows, sb.String())
			sb.Reset()
		} else {
			sb.WriteString(string(char))
		}
	}
	if len(rows) > 0 && (y > -1 && y < len(rows)) {
		return rows[y]
	}
	return ""
}

// getViewLastRow returns the last row content
func (ui *UI) getViewLastRow(v *gocui.View) string {
	rows := []string{}
	buffer := v.ViewBuffer()
	sb := new(strings.Builder)

	for _, char := range buffer {
		if char == '\n' {
			rows = append(rows, sb.String())
			sb.Reset()
		} else {
			sb.WriteString(string(char))
		}
	}

	if len(rows) > 0 {
		// Traverse up the string slice and remove all the trailing spaces from the end of the text.
		fn := func(rows []string) int {
			var idx = 1
			for {
				current := string(rows[len(rows)-idx:][0])
				if current == "" {
					idx++
				} else {
					break
				}
			}
			return idx
		}
		index := fn(rows)
		return rows[len(rows)-index]
	}
	return ""
}

// getViewRowCount returns the number of characters in the row defined by "y"
func (ui *UI) getViewRowCount(v *gocui.View, y int) int {
	return len(ui.getViewRow(v, y))
}

// getViewLastRowCount returns the number of characters in the last row
func (ui *UI) getViewLastRowCount(v *gocui.View) int {
	return len(ui.getViewLastRow(v))
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

// getTotalRows returns the total number of rows of the entire buffer
func (ui *UI) getTotalRows(v *gocui.View) int {
	var rows int
	buffer := v.Buffer()

	for _, char := range []byte(buffer) {
		if string(char) == "\n" {
			rows++
		}
	}
	return rows
}

// getPartialViewBuffer returns the view buffer down until the row defined by "n"
func (ui *UI) getPartialViewBuffer(v *gocui.View, n int) string {
	var idx int
	var newBuffer string

	rows := []string{}
	buffer := v.ViewBuffer()
	sb := new(strings.Builder)

	for _, char := range buffer {
		if char == '\n' {
			rows = append(rows, sb.String())
			sb.Reset()
			if idx > n {
				break
			}
			idx++
		} else {
			sb.WriteString(string(char))
		}
	}
	if idx < n {
		newBuffer = strings.Join(rows[:idx], "\n")
	} else {
		newBuffer = strings.Join(rows[:n], "\n")
	}
	return newBuffer
}
