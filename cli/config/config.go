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

package config

import (
	"os"

	"github.com/spf13/cobra"
)

// NewCommand created a new `config` command
func NewCommand() *cobra.Command {
	configCommand := &cobra.Command{
		Use:     "config",
		Short:   "Arduino configuration commands.",
		Example: "  " + os.Args[0] + " config init",
	}

	configCommand.AddCommand(initAddCommand())
	configCommand.AddCommand(initDeleteCommand())
	configCommand.AddCommand(initDumpCmd())
	configCommand.AddCommand(initInitCommand())
	configCommand.AddCommand(initRemoveCommand())
	configCommand.AddCommand(initSetCommand())

	return configCommand
}
