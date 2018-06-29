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

	"github.com/bcmi-labs/arduino-cli/arduino/libraries/librariesindex"
	"github.com/bcmi-labs/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/bcmi-labs/arduino-cli/commands"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initDownloadCommand() *cobra.Command {
	downloadCommand := &cobra.Command{
		Use:   "download [LIBRARY_NAME(S)]",
		Short: "Downloads one or more libraries without installing them.",
		Long:  "Downloads one or more libraries without installing them.",
		Example: "" +
			"arduino lib download AudioZero       # for the latest version.\n" +
			"arduino lib download AudioZero@1.0.0 # for a specific version.",
		Args: cobra.MinimumNArgs(1),
		Run:  runDownloadCommand,
	}
	return downloadCommand
}

func runDownloadCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino lib download`")

	lm := getLibraryManager()

	logrus.Info("Preparing download")
	pairs := librariesindex.ParseArgs(args)
	downloadLibraries(lm, pairs)
}

func downloadLibraries(lm *librariesmanager.StatusContext, refs []*librariesindex.Reference) {
	libsReleaseToDownload := []*librariesindex.Release{}
	for _, ref := range refs {
		if lib := lm.Index.FindRelease(ref); lib == nil {
			formatter.PrintErrorMessage("Error: library " + ref.String() + " not found")
			os.Exit(commands.ErrBadCall)
		} else {
			libsReleaseToDownload = append(libsReleaseToDownload, lib)
		}
	}

	logrus.Info("Downloading")
	formatter.Print("Downloading libraries...")
	for _, libRelease := range libsReleaseToDownload {
		resp, err := libRelease.Resource.Download()
		if err != nil {
			formatter.PrintError(err, "Error downloading "+libRelease.String())
			os.Exit(commands.ErrNetwork)
		}
		formatter.DownloadProgressBar(resp, libRelease.String())
		if resp.Err() != nil {
			formatter.PrintError(err, "Error downloading "+libRelease.String())
			os.Exit(commands.ErrNetwork)
		}
	}

	logrus.Info("Done")
}
