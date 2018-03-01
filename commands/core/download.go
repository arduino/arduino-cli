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
	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/bcmi-labs/arduino-cli/common/formatter/output"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/bcmi-labs/arduino-cli/cores/packagemanager"
)

func init() {
	command.AddCommand(downloadCommand)
}

var downloadCommand = &cobra.Command{
	Use:   "download [PACKAGER:ARCH[=VERSION]](S)",
	Short: "Downloads one or more cores and corresponding tool dependencies.",
	Long:  "Downloads one or more cores and corresponding tool dependencies.",
	Example: "" +
		"arduino core download arduino:samd       # to download the latest version of arduino SAMD core.\n" +
		"arduino core download arduino:samd=1.6.9 # for a specific version (in this case 1.6.9).",
	Args: cobra.MinimumNArgs(1),
	Run:  runDownloadCommand,
}

func runDownloadCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino core download`")

	// The usage of this depends on the specific command, so it's on-demand
	commands.InitPackageManager()
	pm := packagemanager.PackageManager()

	// FIXME: that's just a PoC; please have mercy...
	/*fmt.Printf("TESTING THE FLUENT PKGMGR\n")

	tool, err := packagemanager.PackageManager().Package("not-found").Tool("avrdude").IsInstalled()
	fmt.Printf("Tool status: %s, %v\n", tool, err)

	tool, err = packagemanager.PackageManager().Package("arduino").Tool("not-found").IsInstalled()
	fmt.Printf("Tool status: %s, %v\n", tool, err)

	tool, err = packagemanager.PackageManager().Package("arduino").Tool("avrdude").IsInstalled()
	fmt.Printf("Tool status: %s, %v\n", tool, err)*/

	logrus.Info("Preparing download")
	// FIXME: add another round of I/O marshalling
	platformReleasesToDownload, toolReleasesToDownload, failOutputs := pm.FindItemsToDownload(
		parsePlatformReferenceArgs(args))
	outputResults := output.CoreProcessResults{
		Cores: failOutputs,
		Tools: []output.ProcessResult{},
	}

	pm.DownloadToolReleaseArchives(toolReleasesToDownload, &outputResults)
	pm.DownloadPlatformReleaseArchives(platformReleasesToDownload, &outputResults)

	formatter.Print(outputResults)
	logrus.Info("Done")
}
