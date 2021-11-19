// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package version

import (
	"os"
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/arduino/arduino-cli/cli/updater"
	"github.com/arduino/arduino-cli/i18n"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	semver "go.bug.st/relaxed-semver"
)

var tr = i18n.Tr

// NewCommand created a new `version` command
func NewCommand() *cobra.Command {
	versionCommand := &cobra.Command{
		Use:     "version",
		Short:   tr("Shows version number of Arduino CLI."),
		Long:    tr("Shows the version number of Arduino CLI which is installed on your system."),
		Example: "  " + os.Args[0] + " version",
		Args:    cobra.NoArgs,
		Run:     runVersionCommand,
	}
	return versionCommand
}

func runVersionCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Executing `arduino-cli version`")
	if strings.Contains(globals.VersionInfo.VersionString, "git-snapshot") || strings.Contains(globals.VersionInfo.VersionString, "nightly") {
		// We're using a development version, no need to check if there's a
		// new release available
		feedback.Print(globals.VersionInfo)
		return
	}

	currentVersion, err := semver.Parse(globals.VersionInfo.VersionString)
	if err != nil {
		feedback.Errorf("Error parsing current version: %s", err)
		os.Exit(errorcodes.ErrGeneric)
	}
	latestVersion := updater.ForceCheckForUpdate(currentVersion)

	versionInfo := globals.VersionInfo
	if feedback.GetFormat() == feedback.JSON && latestVersion != nil {
		// Set this only we managed to get the latest version
		versionInfo.LatestVersion = latestVersion.String()
	}

	feedback.Print(versionInfo)

	if feedback.GetFormat() == feedback.Text && latestVersion != nil {
		updater.NotifyNewVersionIsAvailable(latestVersion.String())
	}
}
