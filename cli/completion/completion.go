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

package completion

import (
	"os"

	"github.com/spf13/cobra"
)

// NewCommand created a new `version` command
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "completion",
		Short: "Generates bash completion scripts",
		Long: `To load completion run
				# arduino-cli completion > arduino-cli.sh
				and then
				# source arduino-cli.sh
				or add the script to /etc/bash_completion.d/
				`,
		Run: run,
	}
}

func run(cmd *cobra.Command, args []string) {
	cmd.Root().GenBashCompletion(os.Stdout)
}
