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
	"os"
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/i18n"
	"github.com/spf13/cobra"
)

var tr = i18n.Tr

// CheckFlagsConflicts is a helper function useful to report errors when more than one conflicting flag is used
func CheckFlagsConflicts(command *cobra.Command, flagNames ...string) {
	for _, flagName := range flagNames {
		if !command.Flag(flagName).Changed {
			return
		}
	}
	feedback.Errorf("Can't use %s flags at the same time.", "--"+strings.Join(flagNames, " "+tr("and")+" --"))
	os.Exit(errorcodes.ErrBadArgument)
}

// CheckFlagsMandatory is a helper function useful to report errors when at least one flag is not used in a group of "required" flags
func CheckFlagsMandatory(command *cobra.Command, flagNames ...string) {
	for _, flagName := range flagNames {
		if command.Flag(flagName).Changed {
			continue
		} else {
			feedback.Errorf("Flag %[1]s is mandatory when used in conjunction with flag %[2]s.", "--"+flagName, "--"+strings.Join(flagNames, " "+tr("and")+" --"))
			os.Exit(errorcodes.ErrBadArgument)
		}
	}
}
