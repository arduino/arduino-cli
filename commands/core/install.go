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
	"path/filepath"

	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/bcmi-labs/arduino-cli/common/formatter/output"
	"github.com/bcmi-labs/arduino-cli/configs"
	"github.com/bcmi-labs/arduino-cli/cores"
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
		"arduino core install arduino:samd       # to download the latest version of arduino SAMD core.\n" +
		"arduino core install arduino:samd=1.6.9 # for a specific version (in this case 1.6.9).",
	Args: cobra.MinimumNArgs(1),
	Run:  runInstallCommand,
}

func runInstallCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino core download`")

	logrus.Info("Getting packages status context")
	status, err := getPackagesStatusContext()
	if err != nil {
		formatter.PrintError(err, "Cannot get packages status context.")
		os.Exit(commands.ErrCoreConfig)
	}

	logrus.Info("Preparing download")

	coresToDownload, toolsToDownload, failOutputs := findDownloadItems(status, parsePlatformReferenceArgs(args))
	failOutputsCount := len(failOutputs)
	outputResults := output.CoreProcessResults{
		Cores: failOutputs,
		Tools: []output.ProcessResult{},
	}

	downloadToolsArchives(toolsToDownload, &outputResults)

	downloadPlatformArchives(coresToDownload, &outputResults)

	logrus.Info("Installing tool dependencies")
	for i, item := range toolsToDownload {
		logrus.WithField("Package", item.Tool.ParentPackage.Name).
			WithField("Name", item.Tool.Name).
			WithField("Version", item.Version).
			Info("Installing tool")

		toolRoot, err := configs.ToolsFolder(item.Tool.ParentPackage.Name).Get()
		if err != nil {
			formatter.PrintError(err, "Cannot get tool install path, try again.")
			os.Exit(commands.ErrCoreConfig)
		}
		possiblePath := filepath.Join(toolRoot, item.Tool.Name, item.Version)

		err = cores.InstallTool(possiblePath, item.GetCompatibleFlavour())
		if err != nil {
			if os.IsExist(err) {
				logrus.WithError(err).Warnf("Cannot install tool `%s`, it is already installed", item.Tool.Name)
				outputResults.Tools[i] = output.ProcessResult{
					ItemName: item.Tool.Name,
					Status:   "Already Installed",
					Path:     possiblePath,
				}
			} else {
				logrus.WithError(err).Warnf("Cannot install tool `%s`", item.Tool.Name)
				outputResults.Tools[i] = output.ProcessResult{
					ItemName: item.Tool.Name,
					Status:   "",
					Error:    err.Error(),
				}
			}
		} else {
			logrus.Info("Adding installed tool to final result")
			outputResults.Tools[i] = output.ProcessResult{
				ItemName: item.Tool.Name,
				Status:   "Installed",
				Path:     possiblePath,
			}
		}
	}

	for i, item := range coresToDownload {
		logrus.WithField("Package", item.Platform.ParentPackage.Name).
			WithField("Name", item.Platform.Name).
			WithField("Version", item.Version).
			Info("Installing core")

		coreRoot, err := configs.CoresFolder(item.Platform.ParentPackage.Name).Get()
		if err != nil {
			formatter.PrintError(err, "Cannot get core install path, try again.")
			os.Exit(commands.ErrCoreConfig)
		}
		possiblePath := filepath.Join(coreRoot, item.Platform.Name, item.Version)

		err = cores.InstallPlatform(possiblePath, item.Resource)
		if err != nil {
			if os.IsExist(err) {
				logrus.WithError(err).Warnf("Cannot install core `%s`, it is already installed", item.Platform.Name)
				outputResults.Cores[i+failOutputsCount] = output.ProcessResult{
					ItemName: item.Platform.Name,
					Status:   "Already Installed",
					Path:     possiblePath,
				}
			} else {
				logrus.WithError(err).Warnf("Cannot install core `%s`", item.Platform.Name)
				outputResults.Cores[i+failOutputsCount] = output.ProcessResult{
					ItemName: item.Platform.Name,
					Status:   "",
					Error:    err.Error(),
				}
			}
		} else {
			logrus.Info("Adding installed core to final result")

			outputResults.Cores[i+failOutputsCount] = output.ProcessResult{
				ItemName: item.Platform.Name,
				Status:   "Installed",
				Path:     possiblePath,
			}
		}
	}

	formatter.Print(outputResults)
	logrus.Info("Done")
}
