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
	"os"

	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initUpgradeCommand() *cobra.Command {
	upgradeCommand := &cobra.Command{
		Use:   "upgrade [PACKAGER:ARCH] ...",
		Short: "Upgrades one or all installed platforms to the latest version.",
		Long:  "Upgrades one or all installed platforms to the latest version.",
		Example: "" +
			"  # upgrade everything to the latest version\n" +
			"  " + commands.AppName + " core upgrade\n\n" +
			"  # upgrade arduino:samd to the latest version\n" +
			"  " + commands.AppName + " core upgrade arduino:samd",
		Run: runUpgradeCommand,
	}
	return upgradeCommand
}

func runUpgradeCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino core upgrade`")

	pm := commands.InitPackageManager()

	platformsRefs := parsePlatformReferenceArgs(args)
	if len(platformsRefs) == 0 {
		upgradeAllPlatforms(pm)
	} else {
		upgrade(pm, platformsRefs)
	}
}

func upgradeAllPlatforms(pm *packagemanager.PackageManager) {
	// Extract all PlatformReference to platforms that have updates
	platformRefs := []*packagemanager.PlatformReference{}

	for _, targetPackage := range pm.GetPackages().Packages {
		for _, platform := range targetPackage.Platforms {
			installed := platform.GetInstalled()
			if installed == nil {
				continue
			}
			latest := platform.GetLatestRelease()
			if !latest.Version.GreaterThan(installed.Version) {
				continue
			}
			platformRefs = append(platformRefs, &packagemanager.PlatformReference{
				Package:              targetPackage.Name,
				PlatformArchitecture: platform.Architecture,
			})
		}
	}

	upgrade(pm, platformRefs)
}

func upgrade(pm *packagemanager.PackageManager, platformsRefs []*packagemanager.PlatformReference) {
	for _, platformRef := range platformsRefs {
		if platformRef.PlatformVersion != nil {
			formatter.PrintErrorMessage("Invalid item " + platformRef.String() + ", upgrade doesn't accept parameters with version")
			os.Exit(commands.ErrBadArgument)
		}
	}

	// Search the latest version for all specified platforms
	toInstallRefs := []*packagemanager.PlatformReference{}
	for _, platformRef := range platformsRefs {
		platform := pm.FindPlatform(platformRef)
		if platform == nil {
			formatter.PrintErrorMessage("Platform " + platformRef.String() + " not found")
			os.Exit(commands.ErrBadArgument)
		}
		installed := platform.GetInstalled()
		if installed == nil {
			formatter.PrintErrorMessage("Platform " + platformRef.String() + " is not installed")
			os.Exit(commands.ErrBadArgument)
		}
		latest := platform.GetLatestRelease()
		if !latest.Version.GreaterThan(installed.Version) {
			formatter.PrintResult("Platform " + platformRef.String() + " is already at the latest version.")
		} else {
			platformRef.PlatformVersion = latest.Version
			toInstallRefs = append(toInstallRefs, platformRef)
		}
	}

	for _, platformRef := range toInstallRefs {
		downloadPlatformByRef(pm, platformRef)
		installPlatformByRef(pm, platformRef)
	}
}
