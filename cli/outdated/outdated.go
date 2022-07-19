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
	"github.com/arduino/arduino-cli/commands/outdated"
	"github.com/arduino/arduino-cli/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cli/table"
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

	outdatedResp, err := outdated.Outdated(context.Background(), &rpc.OutdatedRequest{
		Instance: inst,
	})
	if err != nil {
		feedback.Errorf(tr("Error retrieving outdated cores and libraries: %v"), err)
	}

	feedback.PrintResult(outdatedResult{
		Platforms: outdatedResp.OutdatedPlatforms,
		Libraries: outdatedResp.OutdatedLibraries})
}

// output from this command requires special formatting, let's create a dedicated
// feedback.Result implementation
type outdatedResult struct {
	Platforms []*rpc.Platform
	Libraries []*rpc.InstalledLibrary
}

func (or outdatedResult) Data() interface{} {
	return or
}

func (or outdatedResult) String() string {
	// Prints outdated cores
	t1 := table.New()
	if len(or.Platforms) > 0 {
		t1.SetHeader(tr("ID"), tr("Installed version"), tr("New version"), tr("Name"))
		for _, p := range or.Platforms {
			t1.AddRow(p.Id, p.Installed, p.Latest, p.Name)
		}
	}

	// Prints outdated libraries
	t2 := table.New()
	if len(or.Libraries) > 0 {
		t2.SetHeader(tr("Library name"), tr("Installed version"), tr("New version"))
		for _, l := range or.Libraries {
			t2.AddRow(l.Library.Name, l.Library.Version, l.Release.Version)
		}
	}
	if len(or.Libraries) > 0 && len(or.Platforms) > 0 {
		return t1.Render() + "\n" + t2.Render() // handle the new line between tables
	}
	return t1.Render() + t2.Render()
}
