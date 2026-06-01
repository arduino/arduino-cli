// This file is part of arduino-cli.
//
// Copyright 2025 ARDUINO SA (http://www.arduino.cc/)
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

package profile

import (
	"os"

	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/spf13/cobra"
)

func initProfileLibCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	libCommand := &cobra.Command{
		Use:   "lib",
		Short: i18n.Tr("Commands to manage libraries in sketch profiles."),
		Example: "" +
			"  " + os.Args[0] + " profile lib add AudioZero -m my_profile\n" +
			"  " + os.Args[0] + " profile lib remove Arduino_JSON --profile my_profile\n",
	}
	libCommand.AddCommand(initLibAddCommand(srv))
	libCommand.AddCommand(initLibRemoveCommand(srv))
	return libCommand
}
