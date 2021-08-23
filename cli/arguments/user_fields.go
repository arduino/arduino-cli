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

package arguments

import (
	"bufio"
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/cli/feedback"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"golang.org/x/crypto/ssh/terminal"
)

// AskForUserFields prompts the user to input the provided user fields.
// If there is an error reading input it panics.
func AskForUserFields(userFields []*rpc.UserField) map[string]string {
	writer := feedback.OutputWriter()
	fields := map[string]string{}
	reader := bufio.NewReader(os.Stdin)
	for _, f := range userFields {
		fmt.Fprintf(writer, "%s: ", f.Label)
		var value []byte
		var err error
		if f.Secret {
			value, err = terminal.ReadPassword(int(os.Stdin.Fd()))
		} else {
			value, err = reader.ReadBytes('\n')
		}
		if err != nil {
			panic(err)
		}
		fields[f.Name] = string(value)
	}
	fmt.Fprintln(writer, "")

	return fields
}
