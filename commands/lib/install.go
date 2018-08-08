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

func initInstallCommand() *cobra.Command {
	installCommand := &cobra.Command{
		Use:   "install LIBRARY[@VERSION_NUMBER](S)",
		Short: "Installs one of more specified libraries into the system.",
		Long:  "Installs one or more specified libraries into the system.",
		Example: "" +
			"arduino lib install AudioZero       # for the latest version.\n" +
			"arduino lib install AudioZero@1.0.0 # for the specific version.",
		Args: cobra.MinimumNArgs(1),
		Run:  runInstallCommand,
	}
	return installCommand
}

func runInstallCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino lib install`")
	lm := commands.InitLibraryManager(nil)

	refs, err := librariesindex.ParseArgs(args)
	if err != nil {
		formatter.PrintError(err, "Arguments error")
		os.Exit(commands.ErrBadArgument)
	}
	downloadLibrariesFromReferences(lm, refs)
	installLibrariesFromReferences(lm, refs)
}

func installLibrariesFromReferences(lm *librariesmanager.LibrariesManager, refs []*librariesindex.Reference) {
	libReleases := []*librariesindex.Release{}
	for _, ref := range refs {
		rel := lm.Index.FindRelease(ref)
		if rel == nil {
			formatter.PrintErrorMessage("Error: library " + ref.String() + " not found")
			os.Exit(commands.ErrBadCall)
		}
		libReleases = append(libReleases, rel)
	}
	installLibraries(lm, libReleases)
}

func installLibraries(lm *librariesmanager.LibrariesManager, libReleases []*librariesindex.Release) {
	for _, libRelease := range libReleases {
		logrus.WithField("library", libRelease).Info("Installing library")

		if _, err := lm.Install(libRelease); err != nil {
			logrus.WithError(err).Warn("Error installing library ", libRelease)
			formatter.PrintError(err, "Error installing library: "+libRelease.String())
			os.Exit(commands.ErrGeneric)
		}

		formatter.Print("Installed " + libRelease.String())
	}
}
