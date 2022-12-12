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
	"fmt"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/i18n"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/updater"
	"github.com/arduino/arduino-cli/version"
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

	info := version.VersionInfo
	if strings.Contains(info.VersionString, "git-snapshot") || strings.Contains(info.VersionString, "nightly") {
		// We're using a development version, no need to check if there's a
		// new release available
		feedback.PrintResult(info)
		return
	}

	currentVersion, err := semver.Parse(info.VersionString)
	if err != nil {
		feedback.Fatal(fmt.Sprintf("Error parsing current version: %s", err), feedback.ErrGeneric)
	}
	latestVersion := updater.ForceCheckForUpdate(currentVersion)

	if feedback.GetFormat() != feedback.Text && latestVersion != nil {
		// Set this only we managed to get the latest version
		info.LatestVersion = latestVersion.String()
	}

	feedback.PrintResult(info)

	if feedback.GetFormat() == feedback.Text && latestVersion != nil {
		updater.NotifyNewVersionIsAvailable(latestVersion.String())
	}
}
