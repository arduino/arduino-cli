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

package lib

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/common"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/bcmi-labs/arduino-cli/common/formatter/output"
	"github.com/bcmi-labs/arduino-cli/common/releases"
	"github.com/bcmi-labs/arduino-cli/libraries"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	command.AddCommand(installCommand)
}

var installCommand = &cobra.Command{
	Use:   "install LIBRARY[@VERSION_NUMBER](S)",
	Short: "Installs one of more specified libraries into the system.",
	Long:  "Installs one or more specified libraries into the system.",
	Example: "" +
		"arduino lib install YoutubeApi       # for the latest version.\n" +
		"arduino lib install YoutubeApi@1.0.0 # for the specific version (in this case 1.0.0).",
	Args: cobra.MinimumNArgs(1),
	Run:  runInstallCommand,
}

func runInstallCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino lib install`")

	logrus.Info("Getting Libraries status context")
	status, err := getLibStatusContext()
	if err != nil {
		formatter.PrintError(err, "Cannot get status context.")
		os.Exit(commands.ErrGeneric)
	}

	pairs := libraries.ParseArgs(args)
	libsToDownload, failOutputs := status.Process(pairs)
	outputResults := output.LibProcessResults{
		Libraries: failOutputs,
	}

	libs := make([]releases.DownloadItem, len(libsToDownload))
	for i := range libs {
		libs[i] = releases.DownloadItem(libsToDownload[i])
	}

	logrus.Info("Downloading")
	releases.ParallelDownload(libs, false, "Installed", &outputResults.Libraries, "library")
	logrus.Info("Download finished")

	logrus.Info("Installing")
	folder, err := common.GetDefaultLibFolder()
	if err != nil {
		formatter.PrintError(err, "Cannot get default lib install path.")
		os.Exit(commands.ErrCoreConfig)
	}

	for i, item := range libsToDownload {
		err = libraries.Install(item.Name, item.Release)
		if err != nil {
			logrus.WithError(err).Warn("Library", item.Name, "errored")
			outputResults.Libraries[i] = output.ProcessResult{
				ItemName: item.Name,
				Error:    err.Error(),
			}
		} else {
			outputResults.Libraries[i].Path = filepath.Join(folder, fmt.Sprintf("%s-%s", item.Name, item.Release.VersionName()))
		}
	}

	formatter.Print(outputResults)
	logrus.Info("Done")
}
