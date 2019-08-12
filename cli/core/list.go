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
	"os"
	"sort"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands/core"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/cheynewallace/tabby"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initListCommand() *cobra.Command {
	listCommand := &cobra.Command{
		Use:     "list",
		Short:   "Shows the list of installed platforms.",
		Long:    "Shows the list of installed platforms.",
		Example: "  " + os.Args[0] + " core list",
		Args:    cobra.NoArgs,
		Run:     runListCommand,
	}
	listCommand.Flags().BoolVar(&listFlags.updatableOnly, "updatable", false, "List updatable platforms.")
	return listCommand
}

var listFlags struct {
	updatableOnly bool
}

func runListCommand(cmd *cobra.Command, args []string) {
	instance := instance.CreateInstance()
	logrus.Info("Executing `arduino core list`")

	platforms, err := core.GetPlatforms(instance.Id, listFlags.updatableOnly)
	if err != nil {
		formatter.PrintError(err, "Error listing platforms")
		os.Exit(errorcodes.ErrGeneric)
	}

	if output.JSONOrElse(platforms) {
		outputInstalledCores(platforms)
	}
}

func outputInstalledCores(platforms []*cores.PlatformRelease) {
	if platforms == nil || len(platforms) == 0 {
		return
	}

	table := tabby.New()
	table.AddHeader("ID", "Installed", "Latest", "Name")
	sort.Slice(platforms, func(i, j int) bool {
		return platforms[i].Platform.String() < platforms[j].Platform.String()
	})
	for _, p := range platforms {
		table.AddLine(p.Platform.String(), p.Version.String(), p.Platform.GetLatestRelease().Version.String(), p.Platform.Name)
	}

	table.Print()
}
