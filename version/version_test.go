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
	"testing"

	"github.com/stretchr/testify/require"
)

// TestBuildInjectedInfo tests the Info strings passed to the binary at build time
// in order to have this test green launch your testing using the provided task (see /Taskfile.yml) or use:
//     go test -run TestBuildInjectedInfo -v ./... -ldflags '
//       -X github.com/arduino/arduino-cli/version.versionString=0.0.0-test.preview
//       -X github.com/arduino/arduino-cli/version.commit=deadbeef'
func TestBuildInjectedInfo(t *testing.T) {
	goldenAppName := "arduino-cli"
	goldenInfo := Info{
		Application:   goldenAppName,
		VersionString: "0.0.0-test.preview",
		Commit:        "deadbeef",
	}
	info := NewInfo(goldenAppName)
	require.Equal(t, goldenInfo.Application, info.Application)
	require.Equal(t, goldenInfo.VersionString, info.VersionString)
	require.Equal(t, goldenInfo.Commit, info.Commit)
}
