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
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initUninstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "uninstall PACKAGER:ARCH[@VERSION] ...",
		Short:   "Uninstalls one or more cores and corresponding tool dependencies if no more used.",
		Long:    "Uninstalls one or more cores and corresponding tool dependencies if no more used.",
		Example: "  " + commands.AppName + " core uninstall arduino:samd\n",
		Args:    cobra.MinimumNArgs(1),
		Run:     runUninstallCommand,
	}
}

func runUninstallCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino core download`")

	platformsRefs := parsePlatformReferenceArgs(args)
	pm := commands.InitPackageManager()

	for _, platformRef := range platformsRefs {
		uninstallPlatformByRef(pm, platformRef)
	}
}

func uninstallPlatformByRef(pm *packagemanager.PackageManager, platformRef *packagemanager.PlatformReference) {
	// If no version is specified consider the installed
	if platformRef.PlatformVersion == nil {
		platform := pm.FindPlatform(platformRef)
		if platform == nil {
			formatter.PrintErrorMessage("Platform not found " + platformRef.String())
			os.Exit(commands.ErrBadCall)
		}
		platformRelease := pm.GetInstalledPlatformRelease(platform)
		if platformRelease == nil {
			formatter.PrintErrorMessage("Platform not installed " + platformRef.String())
			os.Exit(commands.ErrBadCall)
		}
		platformRef.PlatformVersion = platformRelease.Version
	}

	platform, tools, err := pm.FindPlatformReleaseDependencies(platformRef)
	if err != nil {
		formatter.PrintError(err, "Could not determine platform dependencies")
		os.Exit(commands.ErrBadCall)
	}

	uninstallPlatformRelease(pm, platform)

	for _, tool := range tools {
		if !pm.IsToolRequired(tool) {
			fmt.Printf("Tool %s is no more required\n", tool)
			uninstallToolRelease(pm, tool)
		}
	}
}

func uninstallPlatformRelease(pm *packagemanager.PackageManager, platformRelease *cores.PlatformRelease) {
	log := pm.Log.WithField("platform", platformRelease)

	log.Info("Uninstalling platform")
	formatter.Print("Uninstalling " + platformRelease.String() + "...")

	if err := pm.UninstallPlatform(platformRelease); err != nil {
		log.WithError(err).Error("Error uninstalling")
		formatter.PrintError(err, "Error uninstalling "+platformRelease.String())
		os.Exit(commands.ErrGeneric)
	}

	log.Info("Platform uninstalled")
	formatter.Print(platformRelease.String() + " uninstalled")
}

func uninstallToolRelease(pm *packagemanager.PackageManager, toolRelease *cores.ToolRelease) {
	log := pm.Log.WithField("Tool", toolRelease)

	log.Info("Uninstalling tool")
	formatter.Print("Uninstalling " + toolRelease.String())

	if err := pm.UninstallTool(toolRelease); err != nil {
		log.WithError(err).Error("Error uninstalling")
		formatter.PrintError(err, "Error uninstalling "+toolRelease.String())
		os.Exit(commands.ErrGeneric)
	}

	log.Info("Tool uninstalled")
	formatter.Print(toolRelease.String() + " uninstalled")
}
