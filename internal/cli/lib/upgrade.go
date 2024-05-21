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

package lib

import (
	"context"
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initUpgradeCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	upgradeCommand := &cobra.Command{
		Use:   "upgrade",
		Short: i18n.Tr("Upgrades installed libraries."),
		Long:  i18n.Tr("This command upgrades an installed library to the latest available version. Multiple libraries can be passed separated by a space. If no arguments are provided, the command will upgrade all the installed libraries where an update is available."),
		Example: "  " + os.Args[0] + " lib upgrade \n" +
			"  " + os.Args[0] + " lib upgrade Audio\n" +
			"  " + os.Args[0] + " lib upgrade Audio ArduinoJson",
		Args: cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runUpgradeCommand(cmd.Context(), srv, args)
		},
	}
	return upgradeCommand
}

func runUpgradeCommand(ctx context.Context, srv rpc.ArduinoCoreServiceServer, args []string) {
	logrus.Info("Executing `arduino-cli lib upgrade`")
	instance := instance.CreateAndInit(ctx, srv)
	Upgrade(ctx, srv, instance, args)
}

// Upgrade upgrades the specified libraries
func Upgrade(ctx context.Context, srv rpc.ArduinoCoreServiceServer, instance *rpc.Instance, libraries []string) {
	var upgradeErr error
	if len(libraries) == 0 {
		req := &rpc.LibraryUpgradeAllRequest{Instance: instance}
		stream := commands.LibraryUpgradeAllStreamResponseToCallbackFunction(ctx, feedback.ProgressBar(), feedback.TaskProgress())
		upgradeErr = srv.LibraryUpgradeAll(req, stream)
	} else {
		for _, libName := range libraries {
			req := &rpc.LibraryUpgradeRequest{
				Instance: instance,
				Name:     libName,
			}
			stream := commands.LibraryUpgradeStreamResponseToCallbackFunction(ctx, feedback.ProgressBar(), feedback.TaskProgress())
			if upgradeErr = srv.LibraryUpgrade(req, stream); upgradeErr != nil {
				break
			}
		}
	}

	if upgradeErr != nil {
		feedback.Fatal(fmt.Sprintf("%s: %v", i18n.Tr("Error upgrading libraries"), upgradeErr), feedback.ErrGeneric)
	}

	logrus.Info("Done")
}
