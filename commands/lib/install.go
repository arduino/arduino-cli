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
	"os"

	"github.com/bcmi-labs/arduino-cli/commands"
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
	libsToDownload, notFoundFailOutputs := status.Process(pairs)
	outputResults := output.LibProcessResults{
		Libraries: notFoundFailOutputs,
	}

	logrus.Info("Downloading")
	releases.ParallelDownload(libsToDownload, false, "Installed", &outputResults.Libraries, commands.GenerateDownloadProgressFormatter())
	logrus.Info("Download finished")

	logrus.Info("Installing")
	if err != nil {
		formatter.PrintError(err, "Cannot get default lib install path.")
		os.Exit(commands.ErrCoreConfig)
	}

	for libName, item := range libsToDownload {
		// FIXME: the library is installed again even if it's already installed

		if libPath, err := libraries.Install(libName, item); err != nil {
			logrus.WithError(err).Warn("Library", libName, "errored")
			outputResults.Libraries = append(outputResults.Libraries, output.ProcessResult{
				ItemName: libName,
				Error:    err.Error(),
			})
		} else {
			// FIXME: this outputResults mess really needs a revamp
			for i := range outputResults.Libraries {
				if outputResults.Libraries[i].ItemName == libName {
					outputResults.Libraries[i].Path = libPath
				}
			}
		}
	}

	formatter.Print(outputResults)
	logrus.Info("Done")
}
