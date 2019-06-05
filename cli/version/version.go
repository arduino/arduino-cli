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

package version

import (
	"fmt"

	"github.com/arduino/arduino-cli/cli"
	"github.com/arduino/arduino-cli/global"
	"github.com/spf13/cobra"
)

// InitCommand prepares the command.
func InitCommand() *cobra.Command {
	versionCommand := &cobra.Command{
		Use:     "version",
		Short:   "Shows version number of arduino CLI.",
		Long:    "Shows version number of arduino CLI which is installed on your system.",
		Example: "  " + global.GetAppName() + " version",
		Args:    cobra.NoArgs,
		Run:     run,
	}
	return versionCommand
}

type versionOutput struct {
	Command string `json:"command"`
	Version string `json:"version"`
}

func run(cmd *cobra.Command, args []string) {
	res := &versionOutput{
		Command: cmd.Parent().Name(),
		Version: global.GetVersion(),
	}
	if cli.OutputJSONOrElse(res) {
		fmt.Printf("%s version %s\n", res.Command, res.Version)
	}
}
