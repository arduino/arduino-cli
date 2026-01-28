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
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"

	"github.com/stretchr/testify/require"
)

func TestDetectionOfChangesInCoreBeforeCompile(t *testing.T) {
	// See: https://github.com/arduino/arduino-cli/issues/2523

	env, cli := integrationtest.CreateEnvForDaemon(t)
	defer env.CleanUp()

	// Create a new instance of the daemon
	grpcInst := cli.Create()
	require.NoError(t, grpcInst.Init("", "", func(ir *commands.InitResponse) {
		fmt.Printf("INIT> %v\n", ir.GetMessage())
	}))

	// Install avr core
	installCl, err := grpcInst.PlatformInstall(context.Background(), "arduino", "avr", "1.8.6", true)
	require.NoError(t, err)
	for {
		installResp, err := installCl.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		fmt.Printf("INSTALL> %v\n", installResp)
	}
	installCl.CloseSend()

	// Utility functions: tryCompile
	sketchPath, err := paths.New("testdata", "bare_minimum").Abs()
	require.NoError(t, err)
	tryCompile := func() error {
		compileCl, err := grpcInst.Compile(context.Background(), "arduino:avr:uno", sketchPath.String(), "")
		require.NoError(t, err)
		defer compileCl.CloseSend()
		for {
			if compileResp, err := compileCl.Recv(); errors.Is(err, io.EOF) {
				return nil
			} else if err != nil {
				return err
			} else {
				fmt.Printf("COMPILE> %v\n", compileResp)
			}
		}
	}

	// Utility functions: tryTouch will touch a file and see if the compile detects the change
	tryTouch := func(fileToTouch *paths.Path) {
		time.Sleep(time.Second) // await at least one second so the timestamp of the file is different

		// touch the file
		f, err := fileToTouch.Append()
		require.NoError(t, err)
		_, err = f.WriteString("\n")
		require.NoError(t, err)
		require.NoError(t, f.Close())

		// try compile: should fail
		err = tryCompile()
		require.Error(t, err)
		require.Contains(t, err.Error(), "The instance is no longer valid and needs to be reinitialized")

		// re-init instance
		require.NoError(t, grpcInst.Init("", "", func(ir *commands.InitResponse) {
			fmt.Printf("INIT> %v\n", ir.GetMessage())
		}))

		// try compile: should succeed
		require.NoError(t, tryCompile())
	}

	avrCorePath := cli.DataDir().Join("packages", "arduino", "hardware", "avr", "1.8.6")
	tryTouch(avrCorePath.Join("installed.json"))
	tryTouch(avrCorePath.Join("platform.txt"))
	tryTouch(avrCorePath.Join("platform.local.txt"))
	tryTouch(avrCorePath.Join("programmers.txt"))
	tryTouch(avrCorePath.Join("boards.txt"))
	tryTouch(avrCorePath.Join("boards.local.txt"))

	// Delete a file and check if the change is detected
	require.NoError(t, avrCorePath.Join("programmers.txt").Remove())
	err = tryCompile()
	require.Error(t, err)
	require.Contains(t, err.Error(), "The instance is no longer valid and needs to be reinitialized")

	// Re-init instance and check again
	require.NoError(t, grpcInst.Init("", "", func(ir *commands.InitResponse) {
		fmt.Printf("INIT> %v\n", ir.GetMessage())
	}))
	require.NoError(t, tryCompile())

	// Create a file and check if the change is detected
	{
		f, err := avrCorePath.Join("programmers.txt").Create()
		require.NoError(t, err)
		require.NoError(t, f.Close())
	}
	err = tryCompile()
	require.Error(t, err)
	require.Contains(t, err.Error(), "The instance is no longer valid and needs to be reinitialized")

	// Re-init instance and check again
	require.NoError(t, grpcInst.Init("", "", func(ir *commands.InitResponse) {
		fmt.Printf("INIT> %v\n", ir.GetMessage())
	}))
	require.NoError(t, tryCompile())
}
