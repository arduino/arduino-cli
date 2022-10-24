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
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/cli/arguments"
	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands/core"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initUpgradeCommand() *cobra.Command {
	var postInstallFlags arguments.PostInstallFlags
	upgradeCommand := &cobra.Command{
		Use:   fmt.Sprintf("upgrade [%s:%s] ...", tr("PACKAGER"), tr("ARCH")),
		Short: tr("Upgrades one or all installed platforms to the latest version."),
		Long:  tr("Upgrades one or all installed platforms to the latest version."),
		Example: "" +
			"  # " + tr("upgrade everything to the latest version") + "\n" +
			"  " + os.Args[0] + " core upgrade\n\n" +
			"  # " + tr("upgrade arduino:samd to the latest version") + "\n" +
			"  " + os.Args[0] + " core upgrade arduino:samd",
		Run: func(cmd *cobra.Command, args []string) {
			runUpgradeCommand(args, postInstallFlags.DetectSkipPostInstallValue())
		},
	}
	postInstallFlags.AddToCommand(upgradeCommand)
	return upgradeCommand
}

func runUpgradeCommand(args []string, skipPostInstall bool) {
	inst := instance.CreateAndInit()
	logrus.Info("Executing `arduino-cli core upgrade`")
	Upgrade(inst, args, skipPostInstall)
}

// Upgrade upgrades one or all installed platforms to the latest version.
func Upgrade(inst *rpc.Instance, args []string, skipPostInstall bool) {
	// if no platform was passed, upgrade allthethings
	if len(args) == 0 {
		targets, err := core.GetPlatforms(&rpc.PlatformListRequest{
			Instance:      inst,
			UpdatableOnly: true,
		})
		if err != nil {
			feedback.Errorf(tr("Error retrieving core list: %v"), err)
			os.Exit(errorcodes.ErrGeneric)
		}

		if len(targets) == 0 {
			feedback.Print(tr("All the cores are already at the latest version"))
			return
		}

		for _, t := range targets {
			args = append(args, t.Id)
		}
	}

	// proceed upgrading, if anything is upgradable
	exitErr := false
	platformsRefs, err := arguments.ParseReferences(args)
	if err != nil {
		feedback.Errorf(tr("Invalid argument passed: %v"), err)
		os.Exit(errorcodes.ErrBadArgument)
	}

	for i, platformRef := range platformsRefs {
		if platformRef.Version != "" {
			feedback.Errorf(tr("Invalid item %s"), args[i])
			exitErr = true
			continue
		}

		r := &rpc.PlatformUpgradeRequest{
			Instance:        inst,
			PlatformPackage: platformRef.PackageName,
			Architecture:    platformRef.Architecture,
			SkipPostInstall: skipPostInstall,
		}

		if _, err := core.PlatformUpgrade(context.Background(), r, output.ProgressBar(), output.TaskProgress()); err != nil {
			if errors.Is(err, &arduino.PlatformAlreadyAtTheLatestVersionError{}) {
				feedback.Print(err.Error())
				continue
			}

			feedback.Errorf(tr("Error during upgrade: %v", err))
			os.Exit(errorcodes.ErrGeneric)
		}
	}

	if exitErr {
		os.Exit(errorcodes.ErrBadArgument)
	}
}
