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
	command.AddCommand(downloadCommand)
}

var downloadCommand = &cobra.Command{
	Use:   "download [LIBRARY_NAME(S)]",
	Short: "Downloads one or more libraries without installing them.",
	Long:  "Downloads one or more libraries without installing them.",
	Example: "" +
		"arduino lib download YoutubeApi       # for the latest version.\n" +
		"arduino lib download YoutubeApi@1.0.0 # for a specific version (in this case 1.0.0).",
	Args: cobra.MinimumNArgs(1),
	Run:  runDownloadCommand,
}

func runDownloadCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino lib download`")

	logrus.Info("Getting Libraries status context")
	status, err := getLibStatusContext()
	if err != nil {
		logrus.WithError(err).Warn("Cannot get status context.")
		os.Exit(commands.ErrGeneric)
	}

	logrus.Info("Preparing download")
	pairs := libraries.ParseArgs(args)
	downloadResources, notFoundFailOutputs := status.Process(pairs)

	logrus.Info("Downloading")
	downloadResults := releases.ParallelDownload(downloadResources, false, commands.GenerateDownloadProgressFormatter())
	logrus.Info("Download finished")

	downloadOutputs := formatter.ExtractProcessResultsFromDownloadResults(downloadResources, downloadResults, "Downloaded")
	out := output.LibProcessResults{
		Libraries: map[string]output.ProcessResult{},
	}
	for name, value := range notFoundFailOutputs {
		out.Libraries[name] = value
	}
	for name, value := range downloadOutputs {
		out.Libraries[name] = value
	}
	formatter.Print(out)
	logrus.Info("Done")
}
