// This file is part of arduino-cli.
//
// Copyright 2024 ARDUINO SA (http://www.arduino.cc/)
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

package commands

import (
	"context"
	"strings"
	"time"

	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/inventory"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cli/version"
	semver "go.bug.st/relaxed-semver"
)

func (s *arduinoCoreServerImpl) CheckForArduinoCLIUpdates(ctx context.Context, req *rpc.CheckForArduinoCLIUpdatesRequest) (*rpc.CheckForArduinoCLIUpdatesResponse, error) {
	currentVersion, err := semver.Parse(version.VersionInfo.VersionString)
	if err != nil {
		return nil, err
	}

	if !s.shouldCheckForUpdate(currentVersion) && !req.GetForceCheck() {
		return &rpc.CheckForArduinoCLIUpdatesResponse{}, nil
	}

	defer func() {
		// Always save the last time we checked for updates at the end
		inventory.Store.Set("updater.last_check_time", time.Now())
		inventory.WriteStore()
	}()

	latestVersion, err := semver.Parse(s.getLatestRelease())
	if err != nil {
		return nil, err
	}

	if currentVersion.GreaterThanOrEqual(latestVersion) {
		// Current version is already good enough
		return &rpc.CheckForArduinoCLIUpdatesResponse{}, nil
	}

	return &rpc.CheckForArduinoCLIUpdatesResponse{
		NewestVersion: latestVersion.String(),
	}, nil
}

// shouldCheckForUpdate return true if it actually makes sense to check for new updates,
// false in all other cases.
func (s *arduinoCoreServerImpl) shouldCheckForUpdate(currentVersion *semver.Version) bool {
	if strings.Contains(currentVersion.String(), "1.0.0-snapshot") || strings.Contains(currentVersion.String(), "nightly") {
		// This is a dev build, no need to check for updates
		return false
	}

	if !s.settings.GetBool("updater.enable_notification") {
		// Don't check if the user disabled the notification
		return false
	}

	if inventory.Store.IsSet("updater.last_check_time") && time.Since(inventory.Store.GetTime("updater.last_check_time")).Hours() < 24 {
		// Checked less than 24 hours ago, let's wait
		return false
	}

	// Don't check when running on CI or on non interactive consoles
	return !feedback.IsCI() && feedback.IsInteractive() && feedback.HasConsole()
}

// getLatestRelease queries the official Arduino download server for the latest release,
// if there are no errors or issues a version string is returned, in all other case an empty string.
func (s *arduinoCoreServerImpl) getLatestRelease() string {
	client, err := s.settings.NewHttpClient()
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
