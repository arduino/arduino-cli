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

package update

import (
	"context"
	"os"

	"github.com/arduino/arduino-cli/internal/cli/core"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	"github.com/arduino/arduino-cli/internal/cli/lib"
	"github.com/arduino/arduino-cli/internal/cli/outdated"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var tr = i18n.Tr

// NewCommand creates a new `update` command
func NewCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	var showOutdated bool
	updateCommand := &cobra.Command{
		Use:     "update",
		Short:   tr("Updates the index of cores and libraries"),
		Long:    tr("Updates the index of cores and libraries to the latest versions."),
		Example: "  " + os.Args[0] + " update",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runUpdateCommand(srv, showOutdated)
		},
	}
	updateCommand.Flags().BoolVar(&showOutdated, "show-outdated", false, tr("Show outdated cores and libraries after index update"))
	return updateCommand
}

func runUpdateCommand(srv rpc.ArduinoCoreServiceServer, showOutdated bool) {
	logrus.Info("Executing `arduino-cli update`")
	ctx := context.Background()
	inst := instance.CreateAndInit(ctx, srv)

	lib.UpdateIndex(inst)
	core.UpdateIndex(ctx, srv, inst)
	instance.Init(ctx, srv, inst)
	if showOutdated {
		outdated.Outdated(ctx, srv, inst)
	}
}
