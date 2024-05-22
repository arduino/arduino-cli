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

	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/arduino/arduino-cli/version"
	"github.com/fatih/color"
)

// NotifyNewVersionIsAvailable prints information about the new latestVersion
func NotifyNewVersionIsAvailable(latestVersion string) {
	msg := fmt.Sprintf("\n\n%s %s → %s\n%s",
		color.YellowString(i18n.Tr("A new release of Arduino CLI is available:")),
		color.CyanString(version.VersionInfo.VersionString),
		color.CyanString(latestVersion),
		color.YellowString("https://arduino.github.io/arduino-cli/latest/installation/#latest-packages"))
	feedback.Warning(msg)
}
