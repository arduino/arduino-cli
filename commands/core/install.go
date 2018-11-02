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

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initInstallCommand() *cobra.Command {
	installCommand := &cobra.Command{
		Use:   "install PACKAGER:ARCH[@VERSION] ...",
		Short: "Installs one or more cores and corresponding tool dependencies.",
		Long:  "Installs one or more cores and corresponding tool dependencies.",
		Example: "  # download the latest version of arduino SAMD core.\n" +
			"  " + commands.AppName + " core install arduino:samd\n\n" +
			"  # download a specific version (in this case 1.6.9).\n" +
			"  " + commands.AppName + " core install arduino:samd=1.6.9",
		Args: cobra.MinimumNArgs(1),
		Run:  runInstallCommand,
	}
	return installCommand
}

func runInstallCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino core download`")

	platformsRefs := parsePlatformReferenceArgs(args)
	pm := commands.InitPackageManagerWithoutBundles()

	for _, platformRef := range platformsRefs {
		installPlatformByRef(pm, platformRef)
	}

	// TODO: Cleanup unused tools
}

func installPlatformByRef(pm *packagemanager.PackageManager, platformRef *packagemanager.PlatformReference) {
	platform, tools, err := pm.FindPlatformReleaseDependencies(platformRef)
	if err != nil {
		formatter.PrintError(err, "Could not determine platform dependencies")
		os.Exit(commands.ErrBadCall)
	}

	installPlatform(pm, platform, tools)
}

func installPlatform(pm *packagemanager.PackageManager, platformRelease *cores.PlatformRelease, requiredTools []*cores.ToolRelease) {
	log := pm.Log.WithField("platform", platformRelease)

	// Prerequisite checks before install
	if platformRelease.IsInstalled() {
		log.Warn("Platform already installed")
		formatter.Print("Platform " + platformRelease.String() + " already installed")
		return
	}
	toolsToInstall := []*cores.ToolRelease{}
	for _, tool := range requiredTools {
		if tool.IsInstalled() {
			log.WithField("tool", tool).Warn("Tool already installed")
			formatter.Print("Tool " + tool.String() + " already installed")
		} else {
			toolsToInstall = append(toolsToInstall, tool)
		}
	}

	// Package download
	for _, tool := range toolsToInstall {
		downloadTool(pm, tool)
	}
	downloadPlatform(pm, platformRelease)

	for _, tool := range toolsToInstall {
		InstallToolRelease(pm, tool)
	}

	// Are we installing or upgrading?
	platform := platformRelease.Platform
	installed := pm.GetInstalledPlatformRelease(platform)
	if installed == nil {
		log.Info("Installing platform")
		formatter.Print("Installing " + platformRelease.String() + "...")
	} else {
		log.Info("Updating platform " + installed.String())
		formatter.Print("Updating " + installed.String() + " with " + platformRelease.String() + "...")
	}

	// Install
	err := pm.InstallPlatform(platformRelease)
	if err != nil {
		log.WithError(err).Error("Cannot install platform")
		formatter.PrintError(err, "Cannot install platform")
		os.Exit(commands.ErrGeneric)
	}

	// If upgrading remove previous release
	if installed != nil {
		err := pm.UninstallPlatform(installed)

		// In case of error try to rollback
		if err != nil {
			log.WithError(err).Error("Error updating platform.")
			formatter.PrintError(err, "Error updating platform")

			// Rollback
			if err := pm.UninstallPlatform(platformRelease); err != nil {
				log.WithError(err).Error("Error rolling-back changes.")
				formatter.PrintError(err, "Error rolling-back changes.")
			}
			os.Exit(commands.ErrGeneric)
		}
	}

	log.Info("Platform installed")
	formatter.Print(platformRelease.String() + " installed")
}

// InstallToolRelease installs a ToolRelease
func InstallToolRelease(pm *packagemanager.PackageManager, toolRelease *cores.ToolRelease) {
	log := pm.Log.WithField("Tool", toolRelease)

	if toolRelease.IsInstalled() {
		log.Warn("Tool already installed")
		formatter.Print("Tool " + toolRelease.String() + " already installed")
		return
	}

	log.Info("Installing tool")
	formatter.Print("Installing " + toolRelease.String() + "...")
	err := pm.InstallTool(toolRelease)
	if err != nil {
		log.WithError(err).Warn("Cannot install tool")
		formatter.PrintError(err, "Cannot install tool: "+toolRelease.String())
		os.Exit(commands.ErrGeneric)
	}

	log.Info("Tool installed")
	formatter.Print(toolRelease.String() + " installed")
}
