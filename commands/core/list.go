/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package core

import (
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/common/formatter/output"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	semver "go.bug.st/relaxed-semver"
)

func initListCommand() *cobra.Command {
	listCommand := &cobra.Command{
		Use:     "list",
		Short:   "Shows the list of installed platforms.",
		Long:    "Shows the list of installed platforms.",
		Example: "  " + commands.AppName + " core list",
		Args:    cobra.NoArgs,
		Run:     runListCommand,
	}
	listCommand.Flags().BoolVar(&listFlags.updatableOnly, "updatable", false, "List updatable platforms.")
	return listCommand
}

var listFlags struct {
	updatableOnly bool
}

func runListCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino core list`")

	pm := commands.InitPackageManager()

	installed := []*output.InstalledPlatform{}
	for _, targetPackage := range pm.GetPackages().Packages {
		for _, platform := range targetPackage.Platforms {
			if platformRelease := pm.GetInstalledPlatformRelease(platform); platformRelease != nil {
				if listFlags.updatableOnly {
					if latest := platform.GetLatestRelease(); latest == nil || latest == platformRelease {
						continue
					}
				}
				var latestVersion *semver.Version
				if latest := platformRelease.Platform.GetLatestRelease(); latest != nil {
					latestVersion = latest.Version
				}
				installed = append(installed, &output.InstalledPlatform{
					ID:        platformRelease.String(),
					Installed: platformRelease.Version,
					Latest:    latestVersion,
					Name:      platformRelease.Platform.Name,
				})
			}
		}
	}

	if len(installed) > 0 {
		formatter.Print(output.InstalledPlatforms{Platforms: installed})
	}
}
