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
	"fmt"
	"os"
	"sort"

	"github.com/arduino/arduino-cli/cli"
	"github.com/arduino/arduino-cli/commands/core"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/output"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initListCommand() *cobra.Command {
	listCommand := &cobra.Command{
		Use:     "list",
		Short:   "Shows the list of installed platforms.",
		Long:    "Shows the list of installed platforms.",
		Example: "  " + cli.VersionInfo.Application + " core list",
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
	instance := cli.CreateInstance()
	logrus.Info("Executing `arduino core list`")

	resp, err := core.PlatformList(context.Background(), &rpc.PlatformListReq{
		Instance:      instance,
		UpdatableOnly: listFlags.updatableOnly,
	})
	if err != nil {
		formatter.PrintError(err, "Error listing platforms")
		os.Exit(cli.ErrGeneric)
	}
	installed := resp.GetInstalledPlatform()
	if installed != nil && len(installed) > 0 {
		if cli.OutputJSONOrElse(installed) {
			outputInstalledCores(installed)
		}
	}
}

func outputInstalledCores(cores []*rpc.InstalledPlatform) {
	table := output.NewTable()
	table.AddRow("ID", "Installed", "Latest", "Name")
	sort.Slice(cores, func(i, j int) bool {
		return cores[i].ID < cores[j].ID
	})
	for _, item := range cores {
		table.AddRow(item.GetID(), item.GetInstalled(), item.GetLatest(), item.GetName())
	}
	fmt.Print(table.Render())
}
