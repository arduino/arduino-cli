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
	"github.com/cavaliercoder/grab"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initDownloadCommand() *cobra.Command {
	downloadCommand := &cobra.Command{
		Use:   "download [PACKAGER:ARCH[=VERSION]](S)",
		Short: "Downloads one or more cores and corresponding tool dependencies.",
		Long:  "Downloads one or more cores and corresponding tool dependencies.",
		Example: "" +
			"  " + commands.AppName + "core download arduino:samd       # to download the latest version of arduino SAMD core.\n" +
			"  " + commands.AppName + "core download arduino:samd=1.6.9 # for a specific version (in this case 1.6.9).",
		Args: cobra.MinimumNArgs(1),
		Run:  runDownloadCommand,
	}
	return downloadCommand
}

func runDownloadCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino core download`")

	platformsRefs := parsePlatformReferenceArgs(args)
	pm := commands.InitPackageManager()
	for _, platformRef := range platformsRefs {
		downloadPlatformByRef(pm, platformRef)
	}
}

func downloadPlatformByRef(pm *packagemanager.PackageManager, platformsRef *packagemanager.PlatformReference) {
	platform, tools, err := pm.FindPlatformReleaseDependencies(platformsRef)
	if err != nil {
		formatter.PrintError(err, "Could not determine platform dependencies")
		os.Exit(commands.ErrBadCall)
	}

	// Check if all tools have a flavor available for the current OS
	for _, tool := range tools {
		if tool.GetCompatibleFlavour() == nil {
			formatter.PrintErrorMessage("The tool " + tool.String() + " is not available for the current OS")
			os.Exit(commands.ErrGeneric)
		}
	}

	// Download tools
	for _, tool := range tools {
		DownloadToolRelease(pm, tool)
	}

	// Download cores
	formatter.Print("Downloading " + platform.String() + "...")
	resp, err := pm.DownloadPlatformRelease(platform)
	download(resp, err, platform.String())

	logrus.Info("Done")
}

// DownloadToolRelease downloads a ToolRelease
func DownloadToolRelease(pm *packagemanager.PackageManager, toolRelease *cores.ToolRelease) {
	formatter.Print("Downloading " + toolRelease.String() + "...")
	resp, err := pm.DownloadToolRelease(toolRelease)
	download(resp, err, toolRelease.String())
}

func download(resp *grab.Response, err error, label string) {
	if err != nil {
		formatter.PrintError(err, "Error downloading "+label)
		os.Exit(commands.ErrNetwork)
	}
	if resp == nil {
		formatter.Print(label + " already downloaded")
		return
	}
	formatter.DownloadProgressBar(resp, label)
	if resp.Err() != nil {
		formatter.PrintError(resp.Err(), "Error downloading "+label)
		os.Exit(commands.ErrNetwork)
	}
}
