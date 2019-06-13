/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package output

import (
	"fmt"
	"unicode/utf8"

	"github.com/fatih/color"
)

var red = color.New(color.FgRed).SprintfFunc()
var blue = color.New(color.FgBlue).SprintfFunc()
var green = color.New(color.FgGreen).SprintfFunc()
var yellow = color.New(color.FgYellow).SprintfFunc()
var white = color.New(color.FgWhite).SprintfFunc()
var hiWhite = color.New(color.FgHiWhite).SprintfFunc()

// Red FIXMEDOC
func Red(in string) *Text {
	return &Text{raw: red(in), clean: in}
}

// Blue FIXMEDOC
func Blue(in string) *Text {
	return &Text{raw: blue(in), clean: in}
}

// Green FIXMEDOC
func Green(in string) *Text {
	return &Text{raw: green(in), clean: in}
}

// Yellow FIXMEDOC
func Yellow(in string) *Text {
	return &Text{raw: yellow(in), clean: in}
}

// White FIXMEDOC
func White(in string) *Text {
	return &Text{raw: white(in), clean: in}
}

// HiWhite FIXMEDOC
func HiWhite(in string) *Text {
	return &Text{raw: hiWhite(in), clean: in}
}

// TextBox FIXMEDOC
type TextBox interface {
	Len() int
	Pad(availableWidth int) string
}

// Text FIXMEDOC
type Text struct {
	clean   string
	raw     string
	justify int
}

// Len FIXMEDOC
func (t *Text) Len() int {
	return utf8.RuneCountInString(t.clean)
}

// func (t *Text) String() string {
// 	return t.raw
// }

// JustifyLeft FIXMEDOC
func (t *Text) JustifyLeft() {
	t.justify = 0
}

// JustifyCenter FIXMEDOC
func (t *Text) JustifyCenter() {
	t.justify = 1
}

// JustifyRight FIXMEDOC
func (t *Text) JustifyRight() {
	t.justify = 2
}

// Pad FIXMEDOC
func (t *Text) Pad(totalLen int) string {
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

// Sprintf FIXMEDOC
func Sprintf(format string, args ...interface{}) TextBox {
	cleanArgs := make([]interface{}, len(args))
	for i, arg := range args {
		if text, ok := arg.(*Text); ok {
			cleanArgs[i], args[i] = text.clean, text.raw
		} else {
			cleanArgs[i] = args[i]
		}
	}

	return &Text{
		clean: fmt.Sprintf(format, cleanArgs...),
		raw:   fmt.Sprintf(format, args...),
	}
}
