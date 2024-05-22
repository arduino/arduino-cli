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
	"os"

	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/spf13/cobra"
)

// NewCommand created a new `lib` command
func NewCommand(srv rpc.ArduinoCoreServiceServer, defaultSettings *rpc.Configuration) *cobra.Command {
	libCommand := &cobra.Command{
		Use:   "lib",
		Short: i18n.Tr("Arduino commands about libraries."),
		Long:  i18n.Tr("Arduino commands about libraries."),
		Example: "" +
			"  " + os.Args[0] + " lib install AudioZero\n" +
			"  " + os.Args[0] + " lib update-index",
	}

	libCommand.AddCommand(initDownloadCommand(srv))
	libCommand.AddCommand(initInstallCommand(srv, defaultSettings))
	libCommand.AddCommand(initListCommand(srv))
	libCommand.AddCommand(initExamplesCommand(srv))
	libCommand.AddCommand(initSearchCommand(srv))
	libCommand.AddCommand(initUninstallCommand(srv))
	libCommand.AddCommand(initUpgradeCommand(srv))
	libCommand.AddCommand(initUpdateIndexCommand(srv))
	libCommand.AddCommand(initDepsCommand(srv))
	return libCommand
}
