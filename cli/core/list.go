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

package core

import (
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands/core"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cli/table"
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
	listCommand.Flags().BoolVar(&listFlags.all, "all", false, "If set return all installable and installed cores, including manually installed.")
	return listCommand
}

var listFlags struct {
	updatableOnly bool
	all           bool
}

func runListCommand(cmd *cobra.Command, args []string) {
	inst := instance.CreateAndInit()
	logrus.Info("Executing `arduino core list`")

	platforms, err := core.GetPlatforms(&rpc.PlatformListRequest{
		Instance:      inst,
		UpdatableOnly: listFlags.updatableOnly,
		All:           listFlags.all,
	})
	if err != nil {
		feedback.Errorf("Error listing platforms: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	feedback.PrintResult(installedResult{platforms})
}

// output from this command requires special formatting, let's create a dedicated
// feedback.Result implementation
type installedResult struct {
	platforms []*rpc.Platform
}

func (ir installedResult) Data() interface{} {
	return ir.platforms
}

func (ir installedResult) String() string {
	if ir.platforms == nil || len(ir.platforms) == 0 {
		return ""
	}

	t := table.New()
	t.SetHeader("ID", "Installed", "Latest", "Name")
	for _, p := range ir.platforms {
		name := p.Name
		if p.Deprecated {
			name = fmt.Sprintf("[DEPRECATED] %s", name)
		}
		t.AddRow(p.Id, p.Installed, p.Latest, name)
	}

	return t.Render()
}
