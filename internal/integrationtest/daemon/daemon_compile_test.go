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

	"github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestDaemonCompileOptions(t *testing.T) {
	// See: https://github.com/arduino/arduino-cli/issues/1614
	// See: https://github.com/arduino/arduino-cli/pull/1820

	env, cli := createEnvForDaemon(t)
	defer env.CleanUp()

	grpcInst := cli.Create()
	require.NoError(t, grpcInst.Init("", "", func(ir *commands.InitResponse) {
		fmt.Printf("INIT> %v\n", ir.GetMessage())
	}))

	plInst, err := grpcInst.PlatformInstall(context.Background(), "arduino", "avr", "1.8.5", true)
	require.NoError(t, err)
	for {
		msg, err := plInst.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		fmt.Printf("INSTALL> %v\n", msg)
	}

	// Install boards.local.txt to trigger bug
	platformLocalTxt := paths.New("testdata", "boards.local.txt-issue1614")
	err = platformLocalTxt.CopyTo(cli.DataDir().
		Join("packages", "arduino", "hardware", "avr", "1.8.5", "boards.local.txt"))
	require.NoError(t, err)

	// Re-init instance to update changes
	require.NoError(t, grpcInst.Init("", "", func(ir *commands.InitResponse) {
		fmt.Printf("INIT> %v\n", ir.GetMessage())
	}))

	// Build sketch (with errors)
	sk := paths.New("testdata", "bare_minimum")
	compile, err := grpcInst.Compile(context.Background(), "arduino:avr:uno:some_menu=bad", sk.String())
	require.NoError(t, err)
	for {
		msg, err := compile.Recv()
		if err == io.EOF {
			require.FailNow(t, "Expected compilation failure", "compilation succeeded")
			break
		}
		if err != nil {
			fmt.Println("COMPILE ERROR>", err)
			break
		}
		if msg.ErrStream != nil {
			fmt.Printf("COMPILE> %v\n", string(msg.GetErrStream()))
		}
	}

	// Build sketch (without errors)
	compile, err = grpcInst.Compile(context.Background(), "arduino:avr:uno:some_menu=good", sk.String())
	require.NoError(t, err)
	for {
		msg, err := compile.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		if msg.ErrStream != nil {
			fmt.Printf("COMPILE> %v\n", string(msg.GetErrStream()))
		}
	}
}
