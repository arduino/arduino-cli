// This file is part of arduino-cli.
//
// Copyright 2023 ARDUINO SA (http://www.arduino.cc/)
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

package assert

import (
	"os/exec"
	"testing"

	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/stretchr/testify/require"
)

func CmdExitCode(t *testing.T, expected feedback.ExitCode, err error) {
	var cmdErr *exec.ExitError
	require.ErrorAs(t, err, &cmdErr)
	require.Equal(t, expected, feedback.ExitCode(cmdErr.ExitCode()))
}
