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
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// AskForUserFields prompts the user to input the provided user fields.
// If there is an error reading input it panics.
func AskForUserFields(userFields []*rpc.UserField) (map[string]string, error) {
	fields := map[string]string{}
	for _, f := range userFields {
		value, err := feedback.InputUserField(f.GetLabel(), f.GetSecret())
		if err != nil {
			return nil, err
		}
		fields[f.GetName()] = value
	}

	return fields, nil
}
