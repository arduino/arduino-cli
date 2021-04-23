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
	"context"
	"os"

	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cli/table"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewCommand creates a new `outdated` command
func NewCommand() *cobra.Command {
	outdatedCommand := &cobra.Command{
		Use:   "outdated",
		Short: "Lists cores and libraries that can be upgraded",
		Long: "This commands shows a list of installed cores and/or libraries\n" +
			"that can be upgraded. If nothing needs to be updated the output is empty.",
		Example: "  " + os.Args[0] + " outdated\n",
		Args:    cobra.NoArgs,
		Run:     runOutdatedCommand,
	}

	return outdatedCommand
}

func runOutdatedCommand(cmd *cobra.Command, args []string) {
	inst := instance.CreateAndInit()
	logrus.Info("Executing `arduino outdated`")

	outdatedResp, err := commands.Outdated(context.Background(), &rpc.OutdatedRequest{
		Instance: inst,
	})
	if err != nil {
		feedback.Errorf("Error retrieving outdated cores and libraries: %v", err)
	}

	// Prints outdated cores
	tab := table.New()
	tab.SetHeader("Core name", "Installed version", "New version")
	if len(outdatedResp.OutdatedPlatforms) > 0 {
		for _, p := range outdatedResp.OutdatedPlatforms {
			tab.AddRow(p.Name, p.Installed, p.Latest)
		}
		feedback.Print(tab.Render())
	}

	// Prints outdated libraries
	tab = table.New()
	tab.SetHeader("Library name", "Installed version", "New version")
	if len(outdatedResp.OutdatedLibraries) > 0 {
		for _, l := range outdatedResp.OutdatedLibraries {
			tab.AddRow(l.Library.Name, l.Library.Version, l.Release.Version)
		}
		feedback.Print(tab.Render())
	}

	logrus.Info("Done")
}
