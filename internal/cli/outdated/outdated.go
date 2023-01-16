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

package outdated

import (
	"os"

	"github.com/arduino/arduino-cli/i18n"
	"github.com/arduino/arduino-cli/internal/cli/core"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	"github.com/arduino/arduino-cli/internal/cli/lib"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var tr = i18n.Tr

// NewCommand creates a new `outdated` command
func NewCommand() *cobra.Command {
	outdatedCommand := &cobra.Command{
		Use:   "outdated",
		Short: tr("Lists cores and libraries that can be upgraded"),
		Long: tr(`This commands shows a list of installed cores and/or libraries
that can be upgraded. If nothing needs to be updated the output is empty.`),
		Example: "  " + os.Args[0] + " outdated\n",
		Args:    cobra.NoArgs,
		Run:     runOutdatedCommand,
	}
	return outdatedCommand
}

func runOutdatedCommand(cmd *cobra.Command, args []string) {
	inst := instance.CreateAndInit()
	logrus.Info("Executing `arduino-cli outdated`")
	Outdated(inst)
}

// Outdated returns a list of outdated platforms and libraries
func Outdated(inst *rpc.Instance) {
	core.List(inst, false, true)
	lib.List(inst, []string{}, false, true)
}
