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

package core

import (
	"context"
	"os"

	"github.com/arduino/arduino-cli/cli"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initUpdateIndexCommand() *cobra.Command {
	updateIndexCommand := &cobra.Command{
		Use:     "update-index",
		Short:   "Updates the index of cores.",
		Long:    "Updates the index of cores to the latest version.",
		Example: "  " + cli.VersionInfo.Application + " core update-index",
		Args:    cobra.NoArgs,
		Run:     runUpdateIndexCommand,
	}
	return updateIndexCommand
}

func runUpdateIndexCommand(cmd *cobra.Command, args []string) {
	instance := cli.CreateInstaceIgnorePlatformIndexErrors()
	logrus.Info("Executing `arduino core update-index`")

	_, err := commands.UpdateIndex(context.Background(), &rpc.UpdateIndexReq{
		Instance: instance,
	}, cli.OutputProgressBar())
	if err != nil {
		formatter.PrintError(err, "Error updating index")
		os.Exit(cli.ErrGeneric)
	}
}
