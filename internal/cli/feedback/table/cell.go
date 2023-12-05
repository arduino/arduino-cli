// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package table

import (
	"fmt"
	"unicode/utf8"

	"github.com/fatih/color"
)

// JustifyMode is used to configure text alignment on cells
type JustifyMode int

// Justify mode enumeration
const (
	JustifyLeft JustifyMode = iota
	JustifyCenter
	JustifyRight
)

// Cell represents a Table cell
type Cell struct {
	clean   string
	raw     string
	justify JustifyMode
}

// NewCell creates a new cell. Color can be nil.
func NewCell(text string, c *color.Color) *Cell {
	styled := text
	if c != nil {
		styled = c.SprintFunc()(text)
	}

	return &Cell{
		raw:     styled,
		clean:   text,
		justify: JustifyLeft,
	}
}

// Len returns the size of the cell, taking into account color codes
func (t *Cell) Len() int {
	return utf8.RuneCountInString(t.clean)
}

// Justify sets text justification
func (t *Cell) Justify(mode JustifyMode) {
	t.justify = mode
}

// Pad sets the cell padding
func (t *Cell) Pad(totalLen int) string {
	delta := totalLen - t.Len()
	switch t.justify {
	case 0:
		return t.raw + spaces(delta)
	case 1:
		return spaces(delta/2) + t.raw + spaces(delta-delta/2)
	case 2:
		return spaces(delta) + t.raw
	}
	panic(fmt.Sprintf("internal error: invalid justify %d", t.justify))
}

func spaces(n int) string {
	res := ""
	for n > 0 {
		res += " "
		n--
	}
	return res
}
