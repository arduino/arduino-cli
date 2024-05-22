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
	"os"

	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/spf13/cobra"
)

// NewCommand created a new `core` command
func NewCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	coreCommand := &cobra.Command{
		Use:     "core",
		Short:   i18n.Tr("Arduino core operations."),
		Long:    i18n.Tr("Arduino core operations."),
		Example: "  " + os.Args[0] + " core update-index",
	}

	coreCommand.AddCommand(initDownloadCommand(srv))
	coreCommand.AddCommand(initInstallCommand(srv))
	coreCommand.AddCommand(initListCommand(srv))
	coreCommand.AddCommand(initUpdateIndexCommand(srv))
	coreCommand.AddCommand(initUpgradeCommand(srv))
	coreCommand.AddCommand(initUninstallCommand(srv))
	coreCommand.AddCommand(initSearchCommand(srv))

	return coreCommand
}
