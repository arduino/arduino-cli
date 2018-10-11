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

	colorable "github.com/mattn/go-colorable"
)

var colorStdout = colorable.NewColorableStdout()

type Output interface {
	EmitTerminal() string
	EmitJSON() string
}

type Kind int

const (
	Terminal Kind = iota
	JSON
)

var defaultOutputKind = Terminal

func SetOutputKind(kind Kind) {
	defaultOutputKind = kind
}

func Emit(data Output) {
	switch defaultOutputKind {
	case Terminal:
		fmt.Fprint(colorStdout, data.EmitTerminal())
	case JSON:
		fmt.Print(data.EmitJSON())
	}
}
