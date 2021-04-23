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

package upgrade

import (
	"context"
	"os"

	"github.com/arduino/arduino-cli/cli/core"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewCommand creates a new `upgrade` command
func NewCommand() *cobra.Command {
	upgradeCommand := &cobra.Command{
		Use:     "upgrade",
		Short:   "Upgrades installed cores and libraries.",
		Long:    "Upgrades installed cores and libraries to latest version.",
		Example: "  " + os.Args[0] + " upgrade",
		Args:    cobra.NoArgs,
		Run:     runUpgradeCommand,
	}

	core.AddPostInstallFlagsToCommand(upgradeCommand)
	return upgradeCommand
}

func runUpgradeCommand(cmd *cobra.Command, args []string) {
	inst := instance.CreateAndInit()
	logrus.Info("Executing `arduino upgrade`")

	err := commands.Upgrade(context.Background(), &rpc.UpgradeRequest{
		Instance:        inst,
		SkipPostInstall: core.DetectSkipPostInstallValue(),
	}, output.NewDownloadProgressBarCB(), output.TaskProgress())

	if err != nil {
		feedback.Errorf("Error upgrading: %v", err)
	}

	logrus.Info("Done")
}
