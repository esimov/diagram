package color

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// Stringf returns a color escape string with format options.
func Stringf(c int, format string, args ...interface{}) string {
	return fmt.Sprintf("\x1b[38;5;%dm%s\x1b[0m", c, fmt.Sprintf(format, args...))
}

// String returns a color escape string.
func String(c int, str string) string {
	return fmt.Sprintf("\x1b[38;5;%dm%s\x1b[0m", c, str)
}

// StringFormat returns a color escape string with extra options.
func StringFormat(c int, str string, args []string) string {
	return fmt.Sprintf("\x1b[38;5;%d;%sm%s\x1b[0m", c, strings.Join(args, ";"), str)
}

// StringFormatBoth fg and bg colors.
func StringFormatBoth(fg, bg int, str string, args []string) string {
	return fmt.Sprintf("\x1b[48;5;%dm\x1b[38;5;%d;%sm%s\x1b[0m", bg, fg, strings.Join(args, ";"), str)
}

// Random color number.
func Random(min, max int) int {
	src := rand.NewSource(time.Now().UnixNano())
	rand := rand.New(src)
	return rand.Intn(max-min) + min
}
