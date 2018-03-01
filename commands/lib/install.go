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

	logrus.Info("Downloading")
	downloadRes := releases.ParallelDownload(libsToDownload, false, commands.GenerateDownloadProgressFormatter())
	logrus.Info("Download finished")

	downloadOutputs := formatter.ExtractProcessResultsFromDownloadResults(libsToDownload, downloadRes, "Installed")
	out := output.LibProcessResults{}
	out.Libraries = append(out.Libraries, notFoundFailOutputs...)
	out.Libraries = append(out.Libraries, downloadOutputs...)

	logrus.Info("Installing")
	if err != nil {
		formatter.PrintError(err, "Cannot get default lib install path.")
		os.Exit(commands.ErrCoreConfig)
	}

	// FIXME: this outputResults mess really needs a revamp

	for libName, item := range libsToDownload {
		// FIXME: the library is installed again even if it's already installed

		if libPath, err := libraries.Install(libName, item); err != nil {
			logrus.WithError(err).Warn("Library", libName, "errored")
			out.Libraries = append(out.Libraries, output.ProcessResult{
				ItemName: libName,
				Error:    err.Error(),
			})
		} else {
			for i := range out.Libraries {
				if out.Libraries[i].ItemName == libName {
					out.Libraries[i].Path = libPath
				}
			}
		}
	}

	formatter.Print(out)
	logrus.Info("Done")
}
