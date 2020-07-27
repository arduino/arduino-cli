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
	"github.com/arduino/arduino-cli/commands/core"
	"github.com/arduino/arduino-cli/commands/lib"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
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

	return upgradeCommand
}

func runUpgradeCommand(cmd *cobra.Command, args []string) {
	inst, err := instance.CreateInstance()
	if err != nil {
		feedback.Errorf("Error upgrading: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	logrus.Info("Executing `arduino upgrade`")

	// Gets list of libraries to upgrade, cores' libraries are ignored since they're upgraded
	// when the core is
	res, err := lib.LibraryList(context.Background(), &rpc.LibraryListReq{
		Instance:  inst,
		All:       false,
		Updatable: true,
	})
	if err != nil {
		feedback.Errorf("Error retrieving library list: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}
	libraries := []string{}
	for _, l := range res.InstalledLibrary {
		libraries = append(libraries, l.Library.Name)
	}

	// Upgrades libraries
	err = lib.LibraryUpgrade(inst.Id, libraries, output.ProgressBar(), output.TaskProgress())
	if err != nil {
		feedback.Errorf("Error upgrading libraries: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	targets, err := core.GetPlatforms(inst.Id, true)
	if err != nil {
		feedback.Errorf("Error retrieving core list: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	for _, t := range targets {
		r := &rpc.PlatformUpgradeReq{
			Instance:        inst,
			PlatformPackage: t.Platform.Package.Name,
			Architecture:    t.Platform.Architecture,
		}
		_, err := core.PlatformUpgrade(context.Background(), r, output.ProgressBar(), output.TaskProgress())
		if err != nil {
			feedback.Errorf("Error during upgrade: %v", err)
			os.Exit(errorcodes.ErrGeneric)
		}
	}

	logrus.Info("Done")
}
