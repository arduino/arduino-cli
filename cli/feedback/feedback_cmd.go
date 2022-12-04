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

package feedback

import (
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/spf13/cobra"
)

// NewCommand creates a new `feedback` command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "feedback",
		Short:  "Test the feedback functions of the arduino CLI.",
		Long:   "This command is for testing purposes only, it is not intended for use by end users.",
		Args:   cobra.NoArgs,
		Hidden: true,
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "input",
		Short: "Test the input functions",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			user, err := InputUserField("User name", false)
			if err != nil {
				Errorf("Error reading input: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
			pass, err := InputUserField("Password", true)
			if err != nil {
				Errorf("Error reading input: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
			nick, err := InputUserField("Nickname", false)
			if err != nil {
				Errorf("Error reading input: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
			Print("Hello " + user + " (a.k.a " + nick + ")!")
			Print("Your password is " + pass + "!")
		},
	})
	return cmd
}
