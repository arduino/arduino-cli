// This file is part of arduino-cli.
//
// Copyright 2026 ARDUINO SA (http://www.arduino.cc/)
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

package cleanup_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"

	"github.com/arduino/arduino-cli/internal/cleanup"
	"github.com/stretchr/testify/require"
)

// TestRepeatedInterrupt checks that the 1st CTRL-C cancels the context and the
// 2nd one terminates the program.
func TestRepeatedInterrupt(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("os.Interrupt cannot be sent to a process on Windows")
	}

	if os.Getenv("CLEANUP_SUBPROCESS") == "1" {
		ctx, cancel := cleanup.InterruptableContext(context.Background())
		defer cancel()
		me, err := os.FindProcess(os.Getpid())
		require.NoError(t, err)
		require.NoError(t, me.Signal(os.Interrupt))
		<-ctx.Done()
		time.Sleep(200 * time.Millisecond) // the signal handler is removed asynchronously
		require.NoError(t, me.Signal(os.Interrupt))
		time.Sleep(2 * time.Second)
		fmt.Println("SURVIVED")
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestRepeatedInterrupt")
	cmd.Env = append(os.Environ(), "CLEANUP_SUBPROCESS=1")
	out, _ := cmd.CombinedOutput()
	require.NotContains(t, string(out), "SURVIVED", "the 2nd interrupt has been swallowed")
}
