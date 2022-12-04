// This file is part of arduino-cli.
//
// Copyright 2022 ARDUINO SA (http://www.arduino.cc/)
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

package feedback

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"golang.org/x/term"
)

// InputUserField prompts the user to input the provided user field.
func InputUserField(prompt string, secret bool) (string, error) {
	if format != Text {
		return "", errors.New(tr("user input not supported for the selected output format"))
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return "", errors.New(tr("user input not supported in non interactive mode"))
	}

	fmt.Fprintf(stdOut, "%s: ", prompt)
	var value []byte
	var err error
	if secret {
		value, err = term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(stdOut)
	} else {
		value, err = bufio.NewReader(os.Stdin).ReadBytes('\n')
		if l := len(value); l > 0 {
			value = value[:l-1]
		}
	}
	if err != nil {
		panic(err)
	}

	return string(value), nil
}
