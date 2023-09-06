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
	"os"

	"github.com/arduino/arduino-cli/i18n"
	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/core"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	"github.com/arduino/arduino-cli/internal/cli/lib"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var tr = i18n.Tr

// NewCommand creates a new `upgrade` command
func NewCommand() *cobra.Command {
	var postInstallFlags arguments.PrePostScriptsFlags
	upgradeCommand := &cobra.Command{
		Use:     "upgrade",
		Short:   tr("Upgrades installed cores and libraries."),
		Long:    tr("Upgrades installed cores and libraries to latest version."),
		Example: "  " + os.Args[0] + " upgrade",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runUpgradeCommand(postInstallFlags.DetectSkipPostInstallValue(), postInstallFlags.DetectSkipPreUninstallValue())
		},
	}
	postInstallFlags.AddToCommand(upgradeCommand)
	return upgradeCommand
}

func runUpgradeCommand(skipPostInstall bool, skipPreUninstall bool) {
	inst := instance.CreateAndInit()
	logrus.Info("Executing `arduino-cli upgrade`")
	lib.Upgrade(inst, []string{})
	core.Upgrade(inst, []string{}, skipPostInstall, skipPreUninstall)
}
