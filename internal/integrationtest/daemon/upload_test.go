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

package daemon_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestUploadCancelation(t *testing.T) {
	env, cli := integrationtest.CreateEnvForDaemon(t)
	defer env.CleanUp()

	grpcInst := cli.Create()
	require.NoError(t, grpcInst.Init("", "", func(ir *commands.InitResponse) {
		fmt.Printf("INIT> %v\n", ir.GetMessage())
	}))

	plInst, err := grpcInst.PlatformInstall(context.Background(), "arduino", "avr", "1.8.6", true)
	require.NoError(t, err)
	for {
		msg, err := plInst.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		fmt.Printf("INSTALL> %v\n", msg)
	}

	// Mock avrdude
	cli.InstallMockedAvrdude(t)

	// Re-init instance to update changes
	require.NoError(t, grpcInst.Init("", "", func(ir *commands.InitResponse) {
		fmt.Printf("INIT> %v\n", ir.GetMessage())
	}))

	// Build sketch for upload
	sk := paths.New("testdata", "bare_minimum")
	compile, err := grpcInst.Compile(context.Background(), "arduino:avr:uno", sk.String(), "")
	require.NoError(t, err)
	for {
		msg, err := compile.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			fmt.Println("COMPILE ERROR>", err)
			require.FailNow(t, "Expected successful compile", "compilation failed")
			break
		}
		if msg.GetOutStream() != nil {
			fmt.Printf("COMPILE OUT> %v\n", string(msg.GetOutStream()))
		}
		if msg.GetErrStream() != nil {
			fmt.Printf("COMPILE ERR> %v\n", string(msg.GetErrStream()))
		}
	}

	// Try upload and interrupt the call after 1 sec
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	upload, err := grpcInst.Upload(ctx, "arduino:avr:uno", sk.String(), "/dev/ttyACM0", "serial")
	require.NoError(t, err)
	checkFile := ""
	for {
		msg, err := upload.Recv()
		if errors.Is(err, io.EOF) {
			require.FailNow(t, "Expected interrupted upload", "upload succeeded")
			break
		}
		if err != nil {
			fmt.Println("UPLOAD ERROR>", err)
			break
		}
		if out := string(msg.GetOutStream()); out != "" {
			fmt.Printf("UPLOAD OUT> %v\n", out)
			if strings.HasPrefix(out, "CHECKFILE: ") {
				checkFile = strings.TrimSpace(out[11:])
			}
		}
		if msg.GetErrStream() != nil {
			fmt.Printf("UPLOAD ERR> %v\n", string(msg.GetErrStream()))
		}
	}
	cancel()

	// Wait 5 seconds.
	// If the mocked avrdude is not killed it will create a checkfile and it will remove it after 5 seconds.
	time.Sleep(5 * time.Second)

	// Test if the checkfile is still there (if the file is there it means that mocked avrdude
	// has been correctly killed).
	require.NotEmpty(t, checkFile)
	require.FileExists(t, checkFile)
	require.NoError(t, os.Remove(checkFile))
}
