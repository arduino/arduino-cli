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
	"fmt"
	"os"

	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/bcmi-labs/arduino-cli/common/formatter/output"
	"github.com/bcmi-labs/arduino-cli/common/releases"
	"github.com/bcmi-labs/arduino-cli/cores"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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

	logrus.Info("Getting packages status context")
	status, err := getPackagesStatusContext()
	if err != nil {
		formatter.PrintError(err, "Cannot get packages status context.")
		os.Exit(commands.ErrCoreConfig)
	}

	logrus.Info("Preparing download")

	coresToDownload, toolsToDownload, failOutputs := findItemsToDownload(status, parsePlatformReferenceArgs(args))
	outputResults := output.CoreProcessResults{
		Cores: failOutputs,
		Tools: []output.ProcessResult{},
	}

	downloadToolsArchives(toolsToDownload, &outputResults)

	downloadPlatformArchives(coresToDownload, &outputResults)

	formatter.Print(outputResults)
	logrus.Info("Done")
}

func downloadToolsArchives(tools []*cores.ToolRelease, results *output.CoreProcessResults) {
	downloads := map[string]*releases.DownloadResource{}
	for _, tool := range tools {
		resource := tool.GetCompatibleFlavour()
		if resource == nil {
			formatter.PrintError(fmt.Errorf("missing tool %s", tool), "A release of the tool is not available for your OS")
		}
		downloads[tool.Tool.Name+"@"+tool.Version] = tool.GetCompatibleFlavour()
	}
	logrus.Info("Downloading tools")
	downloadRes := releases.ParallelDownload(downloads, false, commands.GenerateDownloadProgressFormatter())

	for name, res := range downloadRes {
		path, err := downloads[name].ArchivePath()
		if err != nil {
			// FIXME: do something!!
			logrus.Error("Could not determine library archive path:", err)
		}
		status := ""
		if res.Error == nil {
			status = "Downloaded"
		}
		results.Tools = append(results.Tools, output.ProcessResult{
			ItemName: name,
			Path:     path,
			Error:    res.Error.Error(),
			Status:   status,
		})
	}
}

func downloadPlatformArchives(platforms []*cores.PlatformRelease, results *output.CoreProcessResults) {
	downloads := map[string]*releases.DownloadResource{}
	for _, platform := range platforms {
		downloads[platform.Platform.Package.Name+":"+platform.Platform.Name+"@"+platform.Version] = platform.Resource
	}
	logrus.Info("Downloading cores")
	downloadRes := releases.ParallelDownload(downloads, false, commands.GenerateDownloadProgressFormatter())

	for name, res := range downloadRes {
		path, err := downloads[name].ArchivePath()
		if err != nil {
			// FIXME: do something!!
			logrus.Error("Could not determine library archive path:", err)
		}
		status := ""
		if res.Error == nil {
			status = "Downloaded"
		}
		results.Cores = append(results.Cores, output.ProcessResult{
			ItemName: name,
			Path:     path,
			Error:    res.Error.Error(),
			Status:   status,
		})
	}
}
