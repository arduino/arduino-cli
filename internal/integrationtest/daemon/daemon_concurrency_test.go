// This file is part of arduino-cli.
//
// Copyright 2022 ARDUINO SA (http://www.arduino.cc/)
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

package daemon_test

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestArduinoCliDaemonCompileWithLotOfOutput(t *testing.T) {
	// See: https://github.com/arduino/arduino-cli/issues/2169

	env, cli := createEnvForDaemon(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)

	grpcInst := cli.Create()
	require.NoError(t, grpcInst.Init("", "", func(ir *commands.InitResponse) {
		fmt.Printf("INIT> %v\n", ir.GetMessage())
	}))

	sketchPath, err := paths.New("..", "testdata", "ManyWarningsSketch").Abs()
	require.NoError(t, err)

	testCompile := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		compile, err := grpcInst.Compile(ctx, "arduino:avr:uno", sketchPath.String(), "all")
		require.NoError(t, err)
		msgCount := 0
		for {
			_, err := compile.Recv()
			if err == io.EOF {
				break
			}
			msgCount++
			require.NoError(t, err)
		}
		fmt.Println("Received", msgCount, "messages.")
	}

	// The synchronization bug doesn't always happens, try 10 times to
	// increase the chance to trigger it.
	testCompile()
	testCompile()
	testCompile()
	testCompile()
	testCompile()
	testCompile()
	testCompile()
	testCompile()
	testCompile()
	testCompile()
}
