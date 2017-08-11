package canvas

import (
	"strings"
	"github.com/esimov/diagram/io"
	"reflect"
	"github.com/fogleman/gg"
)

// Auxiliary Point struct used during parsing.
type Point struct {
	x, y int
}

// Create a new Point struct
func NewPoint(x, y int)*Point {
	return &Point{x, y}
}

// Line struct defines the line x & y coordinates, the starting and ending type and the color.
type Line struct {
	x0, y0 	int
	start 	string
	x1, y1 	int
	end 	string
	color 	string
}

// Line from (x0, y0) to (x1, y1) with the given color at the start and end.
func NewLine(x0, y0 int, start string, x1, y1 int, end string, color string)*Line {
	return &Line{x0, y0, start, x1, y1, end, color}
}

// Text struct that contains the x & y coordinates and color.
type Text struct {
	x0, y0 	int
	text 	string
	color 	string
}

// Text annotation at (x0, y0) with the given color.
func NewText(x0, y0 int, text, color string)*Text {
	return &Text{x0, y0, text, color}
}

// Compounded struct with Line & Text
type Figures struct {
	Line
	Text
}

// Empty struct
type Diagram struct {}

// Parses given ASCII art string into a list of figures.
func (d *Diagram) ParseASCIIArt(str string)[]*Figures {
	var figures []*Figures

	lines := strings.Split(str, "\n")
	height := len(lines)

	// Get diagram widest line.
	width := func(lines []string) int {
		var max int
		for _, line := range lines {
			if len(line) > max {
				max = len(line)
			}
		}
		return max
	}(lines)

	data := make([][]string, height)

	// Convert strings into a mutable matrix of characters.
	for y := 0; y < height; y++ {
		line := lines[y]
		data[y] = make([]string, width)
		for x := 0; x < len(line); x++ {
			data[y][x] = string(line[x])
		}
		for x := len(line); x < width; x++ {
			data[y][x] = " "
		}
	}

	// Get a character from the slice or 0 if we are out of bounds.
	at := func(y, x int) string {
		if 0 <= y && y < height && 0 <= x && x < width {
			return data[y][x]
		}
		return ""
	}

	// Returns true if the character can be part of the line.
	isPartOfLine := func(x, y int) bool {
		c := at(y, x)
		switch {
		case c == "|" || c == "-" || c == "+" || c == "~" || c == "!" :
			return true
		}
		return false
	}

	toColor := func(x, y int) string {
		c := at(y, x)
		switch {
		case c =="~" || c == "!":
			return "#666"
		}
		return ""
	}

	// Returns true if characters is line ending decoration.
	isLineEnding := func(x, y int) bool {
		c := at(y, x)
		switch {
		case c == "*" || c == "<" || c == ">" || c == "^" || c == "v" :
			return true
		}
		return false
	}

	// Finds a character that belongs to unextracted line.
	findLineChar := func()*Point {
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				if data[y][x] == "|" || data[y][x] == "-" {
					return NewPoint(x, y)
				}
			}
		}
		return nil
	}

	// Converts line's character to the direction of line's growth.
	dir := map[string]*Point{
		"-" : NewPoint(1, 0),
		"|" : NewPoint(0, 1),
	}

	// Erases character that belongs to the extracted line.
	eraseChar := func(x, y, dx, dy int) {
		c := at(y, x)
		switch {
		case c == "|" || c == "-" || c == "*" || c == ">" || c == "<" || c == "^" || c == "v" || c == "~" || c == "~" :
			data[y][x] = " "
			return
		case c == "+":
			dx = 1 - dx
			dy = 1 - dy
			data[y][x] = " "

			c1 := at(y - dy, x - dx)
			switch {
			case c1 == "|" || c1 == "!" || c1 == "+":
				data[y][x] = "|"
				return
			case c1 == "-" || c1 == "~" || c1 == "+":
				data[y][x] = "-"
				return
			}

			c2 := at(y + dy, x + dx)
			switch {
			case c2 == "|" || c2 == "!" || c2 == "+":
				data[y][x] = "|"
				return
			case c2 == "-" || c2 == "~" || c2 == "+":
				data[y][x] = "-"
				return
			}
			return
		}
	}

	// Erase the given extracted line.
	erase := func(line *Line) {
		var dx, dy int
		if line.x0 != line.x1 {
			dx = 1
		} else {
			dx = 0
		}
		if line.y0 != line.y1 {
			dy = 1
		} else {
			dy = 0
		}
		if dx != 0 || dy != 0 {
			x0, y0 := line.x0 + dx, line.y0 + dy
			x1, y1 := line.x1 - dx, line.y1 - dy

			for x0 <= x1 && y0 <= y1 {
				eraseChar(x0, y0, dx, dy)
				x0 += dx
				y0 += dy
			}
			eraseChar(line.x0, line.y0, dx, dy)
			eraseChar(line.x1, line.y1, dx, dy)
		} else {
			eraseChar(line.x0, line.y0, dx, dy)
		}
	}

	// Extract a single line and erase it from the ascii art matrix.
	extractLine := func() bool {
		var color, start, end string

		ch := findLineChar()
		if ch == nil {
			return false
		}

		d := dir[data[ch.y][ch.x]]
		// Find line's start by advancing in the opposite direction.
		x0 := ch.x
		y0 := ch.y
		for isPartOfLine(x0 - d.x, y0 - d.y) {
			x0 -= d.x
			y0 -= d.y
			if color == "" {
				color = toColor(x0, y0)
			}
		}
		if isLineEnding(x0 - d.x, y0 - d.y) {
			// Line has a decorated start. Extract is as well.
			x0 -= d.x
			y0 -= d.y
			if data[y0][x0] == "*" {
				start = "circle"
			} else {
				start = "arrow"
			}
		}
		// Find line's end by advancing forward in the given direction.
		x1 := ch.x
		y1 := ch.y
		for isPartOfLine(x1 + d.x, y1 + d.y) {
			x1 += d.x
			y1 += d.y
			if color == "" {
				color = toColor(x1, y1)
			}
		}
		if isLineEnding(x1 + d.x, y1 + d.y) {
			// Line has a decorated end. Extract it.
			x1 += d.x
			y1 += d.y
			if data[y1][x1] == "*" {
				end = "circle"
			} else {
				end = "arrow"
			}
		}


		// Create line object and erase line from the ascii art matrix.
		line := NewLine(x0, y0, start, x1, y1, end, color)

		figures = append(figures, &Figures{*line, Text{}})
		erase(line)

		// Adjust line start and end to accomodate for arrow endings.
		// Those should not intersect with their targets but should touch them
		// instead. Should be done after erasure to ensure that erase deletes
		// arrowheads.
		if start == "arrow" {
			line.x0 -= d.x
			line.y0 -= d.y
		}

		if end == "arrow" {
			line.x1 += d.x
			line.y1 += d.y
		}
		return true
	}
	// Extract all non space characters that were left after line extraction
	// as text objects.
	extractText := func() {
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				if data[y][x] != " " {
					start, end := x, x
					for end < width && data[y][end] != " " {
						end++
					}
					getRange := func(start, end int)[]string {
						return data[y][start:end]
					}
					//fmt.Println(getRange(start, end))
					text := strings.Join(getRange(start, end), "")

					// Check if it can be concatenated with a previously found text annotation.
					prev := figures[len(figures)-1]
					if prev.Text.x0 + len(prev.text) + 1 == start {
						// If they touch concatenate them
						prev.text = prev.text + " " + text
					} else {
						color := "#000"
						if string(text[0]) == "\\" && string(text[len(text) - 1]) == "\\" {
							text = text[1: len(text) - 1]
							color = "#666"
						}
						newtext := NewText(x, y, text, color)
						figures = append(figures, &Figures{Line{}, *newtext})
					}
					x = end
				}
			}
		}
	}

	for extractLine() {}
	extractText()

	return figures
}

// Draw a diagram from the ascii art.
func DrawDiagram(text string) {
	var width, height int

	diagram := &Diagram{}
	content := string(io.ReadFile(text))
	figures := diagram.ParseASCIIArt(content)

	for _, fig := range figures {
		if reflect.TypeOf(fig.Line).String() == "canvas.Line" {
			width = max(width, int(X(float64(fig.Line.x1) + 1)))
			height = max(height, int(Y(float64(fig.Line.y1) + 1)))
		}
	}

	ctx := gg.NewContext(width, height)
	canvas := NewCanvas(ctx, "//home/esimov/Projects/Go/src/github.com/oakmound/oak/render/default_assets/font/luxisr.ttf", 2)
	canvas.DrawRectangle(0, 0, float64(width), float64(height))
	canvas.SetRGBA(1, 1, 1, 1)
	canvas.Fill()
	for _, fig := range figures {
		fig.Line.Draw(canvas)
		fig.Text.Draw(canvas)
	}
	canvas.SavePNG("out.png")
}