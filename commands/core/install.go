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

	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/bcmi-labs/arduino-cli/common/formatter/output"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	command.AddCommand(installCommand)
}

var installCommand = &cobra.Command{
	Use:   "install [PACKAGER:ARCH[=VERSION]](S)",
	Short: "Installs one or more cores and corresponding tool dependencies.",
	Long:  "Installs one or more cores and corresponding tool dependencies.",
	Example: "" +
		"arduino core install arduino:samd       # to download the latest version of arduino SAMD core." +
		"arduino core install arduino:samd=1.6.9 # for a specific version (in this case 1.6.9).",
	Args: cobra.MinimumNArgs(1),
	Run:  runInstallCommand,
}

func runInstallCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino core download`")

	pm := commands.InitPackageManager()
	platformsRefs := parsePlatformReferenceArgs(args)
	downloadPlatforms(pm, platformsRefs)

	platforms, tools, err := pm.FindItemsToDownload(platformsRefs)
	if err != nil {
		formatter.PrintError(err, "Could not determine platform dependencies")
		os.Exit(commands.ErrBadCall)
	}

	outputResults := output.CoreProcessResults{
		Cores: map[string]output.ProcessResult{},
		Tools: map[string]output.ProcessResult{},
	}

	logrus.Info("Installing tools")
	formatter.Print("Installing tools...")

	if err := pm.InstallToolReleases(tools, &outputResults); err != nil {
		formatter.PrintError(err, "Error installing tools.")
		os.Exit(commands.ErrGeneric)
	}

	logrus.Info("Installing cores")
	formatter.Print("Installing platforms...")
	if err := pm.InstallPlatformReleases(platforms, &outputResults); err != nil {
		formatter.PrintError(err, "Error installing platform.")
		os.Exit(commands.ErrGeneric)
	}

	formatter.Print("Results:")
	formatter.Print(outputResults)
	logrus.Info("Done")
}
