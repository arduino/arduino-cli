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

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/configuration"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var postInstallFlags struct {
	runPostInstall  bool
	skipPostInstall bool
}

func detectSkipPostInstallValue() bool {
	if postInstallFlags.runPostInstall && postInstallFlags.skipPostInstall {
		feedback.Errorf("The flags --run-post-install and --skip-post-install can't be both set at the same time.")
		os.Exit(errorcodes.ErrBadArgument)
	}
	if postInstallFlags.runPostInstall {
		logrus.Info("Will run post-install by user request")
		return false
	}
	if postInstallFlags.skipPostInstall {
		logrus.Info("Will skip post-install by user request")
		return true
	}

	if !configuration.IsInteractive {
		logrus.Info("Not running from console, will skip post-install by default")
		return true
	}
	logrus.Info("Running from console, will run post-install by default")
	return false
}

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

	upgradeCommand.Flags().BoolVar(&postInstallFlags.runPostInstall, "run-post-install", false, "Force run of post-install scripts (if the CLI is not running interactively).")
	upgradeCommand.Flags().BoolVar(&postInstallFlags.skipPostInstall, "skip-post-install", false, "Force skip of post-install scripts (if the CLI is running interactively).")
	return upgradeCommand
}

func runUpgradeCommand(cmd *cobra.Command, args []string) {
	inst, err := instance.CreateInstance()
	if err != nil {
		feedback.Errorf("Error upgrading: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	logrus.Info("Executing `arduino upgrade`")

	err = commands.Upgrade(context.Background(), &rpc.UpgradeReq{
		Instance:        inst,
		SkipPostInstall: detectSkipPostInstallValue(),
	}, output.NewDownloadProgressBarCB(), output.TaskProgress())

	if err != nil {
		feedback.Errorf("Error upgrading: %v", err)
	}

	logrus.Info("Done")
}
