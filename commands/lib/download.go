/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package lib

import (
	"os"

	"github.com/arduino/arduino-cli/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
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

	lm := commands.InitLibraryManager(nil)

	logrus.Info("Preparing download")
	pairs, err := librariesindex.ParseArgs(args)
	if err != nil {
		formatter.PrintError(err, "Arguments error")
		os.Exit(commands.ErrBadArgument)
	}
	downloadLibrariesFromReferences(lm, pairs)
}

func downloadLibrariesFromReferences(lm *librariesmanager.LibrariesManager, refs []*librariesindex.Reference) {
	libReleases := []*librariesindex.Release{}
	for _, ref := range refs {
		if lib := lm.Index.FindRelease(ref); lib == nil {
			formatter.PrintErrorMessage("Error: library " + ref.String() + " not found")
			os.Exit(commands.ErrBadCall)
		} else {
			libReleases = append(libReleases, lib)
		}
	}
	downloadLibraries(lm, libReleases)
}

func downloadLibraries(lm *librariesmanager.LibrariesManager, libReleases []*librariesindex.Release) {
	logrus.Info("Downloading libraries")
	for _, libRelease := range libReleases {
		resp, err := libRelease.Resource.Download(lm.DownloadsDir)
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
