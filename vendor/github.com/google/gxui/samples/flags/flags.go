// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package flags holds command line options common to all GXUI samples.
package flags

import (
	"github.com/google/gxui"
	"github.com/google/gxui/themes/dark"
)

var DefaultScaleFactor float32
var FlagTheme string

func init() {
	DefaultScaleFactor = 1.0
	FlagTheme = "dark"
}

// CreateTheme creates and returns the theme specified on the command line.
// The default theme is dark.
func CreateTheme(driver gxui.Driver) gxui.Theme {
	return dark.CreateTheme(driver)
}
