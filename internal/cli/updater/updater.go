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

package updater

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/arduino/arduino-cli/i18n"
	"github.com/arduino/arduino-cli/internal/arduino/httpclient"
	"github.com/arduino/arduino-cli/internal/cli/configuration"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/inventory"
	"github.com/arduino/arduino-cli/version"
	"github.com/fatih/color"
	semver "go.bug.st/relaxed-semver"
)

var tr = i18n.Tr

// CheckForUpdate returns the latest available version if greater than
// the one running and it makes sense to check for an update, nil in all other cases
func CheckForUpdate(currentVersion *semver.Version) *semver.Version {
	if !shouldCheckForUpdate(currentVersion) {
		return nil
	}

	return ForceCheckForUpdate(currentVersion)
}

// ForceCheckForUpdate always returns the latest available version if greater than
// the one running, nil in all other cases
func ForceCheckForUpdate(currentVersion *semver.Version) *semver.Version {
	defer func() {
		// Always save the last time we checked for updates at the end
		inventory.Store.Set("updater.last_check_time", time.Now())
		inventory.WriteStore()
	}()

	latestVersion, err := semver.Parse(getLatestRelease())
	if err != nil {
		return nil
	}

	if currentVersion.GreaterThanOrEqual(latestVersion) {
		// Current version is already good enough
		return nil
	}

	return latestVersion
}

// NotifyNewVersionIsAvailable prints information about the new latestVersion
func NotifyNewVersionIsAvailable(latestVersion string) {
	msg := fmt.Sprintf("\n\n%s %s → %s\n%s",
		color.YellowString(tr("A new release of Arduino CLI is available:")),
		color.CyanString(version.VersionInfo.VersionString),
		color.CyanString(latestVersion),
		color.YellowString("https://arduino.github.io/arduino-cli/latest/installation/#latest-packages"))
	feedback.Warning(msg)
}

// shouldCheckForUpdate return true if it actually makes sense to check for new updates,
// false in all other cases.
func shouldCheckForUpdate(currentVersion *semver.Version) bool {
	if strings.Contains(currentVersion.String(), "git-snapshot") || strings.Contains(currentVersion.String(), "nightly") {
		// This is a dev build, no need to check for updates
		return false
	}

	if !configuration.Settings.GetBool("updater.enable_notification") {
		// Don't check if the user disabled the notification
		return false
	}

	if inventory.Store.IsSet("updater.last_check_time") && time.Since(inventory.Store.GetTime("updater.last_check_time")).Hours() < 24 {
		// Checked less than 24 hours ago, let's wait
		return false
	}

	// Don't check when running on CI or on non interactive consoles
	return !isCI() && configuration.IsInteractive && configuration.HasConsole
}

// based on https://github.com/watson/ci-info/blob/HEAD/index.js
func isCI() bool {
	return os.Getenv("CI") != "" || // GitHub Actions, Travis CI, CircleCI, Cirrus CI, GitLab CI, AppVeyor, CodeShip, dsari
		os.Getenv("BUILD_NUMBER") != "" || // Jenkins, TeamCity
		os.Getenv("RUN_ID") != "" // TaskCluster, dsari
}

// getLatestRelease queries the official Arduino download server for the latest release,
// if there are no errors or issues a version string is returned, in all other case an empty string.
func getLatestRelease() string {
	client, err := httpclient.New()
	if err != nil {
		return ""
	}

	// We just use this URL to check if there's a new release available and
	// never show it to the user, so it's fine to use the Linux one for all OSs.
	URL := "https://downloads.arduino.cc/arduino-cli/arduino-cli_latest_Linux_64bit.tar.gz"
	res, err := client.Head(URL)
	if err != nil {
		// Yes, we ignore it
		return ""
	}

	// Get redirected URL
	location := res.Request.URL.String()

	// The location header points to the latest release of the CLI, it's supposed to be formatted like this:
	// https://downloads.arduino.cc/arduino-cli/arduino-cli_0.18.3_Linux_64bit.tar.gz
	// so we split it to get the version, if there are not enough splits something must have gone wrong.
	split := strings.Split(location, "_")
	if len(split) < 2 {
		return ""
	}

	return split[1]
}
