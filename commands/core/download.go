/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */

package core

import (
	"os"

	"github.com/bcmi-labs/arduino-cli/arduino/cores/packagemanager"
	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
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
			"arduino core download arduino:samd       # to download the latest version of arduino SAMD core.\n" +
			"arduino core download arduino:samd=1.6.9 # for a specific version (in this case 1.6.9).",
		Args: cobra.MinimumNArgs(1),
		Run:  runDownloadCommand,
	}
	return downloadCommand
}

func runDownloadCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino core download`")

	pm := commands.InitPackageManager()
	platformsRefs := parsePlatformReferenceArgs(args)
	downloadPlatforms(pm, platformsRefs)
}

func downloadPlatforms(pm *packagemanager.PackageManager, platformsRefs []packagemanager.PlatformReference) {
	platforms, tools, err := pm.FindItemsToDownload(platformsRefs)
	if err != nil {
		formatter.PrintError(err, "Could not determine platform dependencies")
		os.Exit(commands.ErrBadCall)
	}

	// Check if all tools have a flavour available for the current OS
	for _, tool := range tools {
		if tool.GetCompatibleFlavour() == nil {
			formatter.PrintErrorMessage("The tool " + tool.String() + " is not available for the current OS")
			os.Exit(commands.ErrGeneric)
		}
	}

	// Download tools
	formatter.Print("Downloading tools...")
	for _, tool := range tools {
		resp, err := pm.DownloadToolRelease(tool)
		download(resp, err, tool.String())
	}

	// Download cores
	formatter.Print("Downloading cores...")
	for _, platform := range platforms {
		resp, err := pm.DownloadPlatformRelease(platform)
		download(resp, err, platform.String())
	}

	logrus.Info("Done")
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
