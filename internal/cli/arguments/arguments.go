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
	"strings"

	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/spf13/cobra"
)

// CheckFlagsConflicts is a helper function useful to report errors when more than one conflicting flag is used
func CheckFlagsConflicts(command *cobra.Command, flagNames ...string) {
	var used []string
	for _, flagName := range flagNames {
		if command.Flag(flagName).Changed {
			used = append(used, flagName)
		}
	}
	if len(used) <= 1 {
		return
	}
	flags := "--" + strings.Join(used, ", --")
	msg := i18n.Tr("Can't use the following flags together: %s", flags)
	feedback.Fatal(msg, feedback.ErrBadArgument)
}

// CheckFlagsMandatory is a helper function useful to report errors when at least one flag is not used in a group of "required" flags
func CheckFlagsMandatory(command *cobra.Command, flagNames ...string) {
	for _, flagName := range flagNames {
		if command.Flag(flagName).Changed {
			continue
		}
		flags := "--" + strings.Join(flagNames, ", --")
		msg := i18n.Tr("Flag %[1]s is mandatory when used in conjunction with: %[2]s", "--"+flagName, flags)
		feedback.Fatal(msg, feedback.ErrBadArgument)
	}
}
