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

package upload_mock_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestUploadInterrupt(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	platform := cli.SketchbookDir().Join("hardware", "foo")
	require.NoError(t, paths.New("..", "testdata", "foo").CopyDirTo(platform))
	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.5")
	require.NoError(t, err)
	sketch := cli.CopySketch("bare_minimum")
	binary := env.RootDir().Join("sketch.bin")
	require.NoError(t, binary.WriteFile(nil))

	// Model a hung uploader that ignores the interrupt sent by the terminal.
	uploader := env.RootDir().Join("uploader.sh")
	require.NoError(t, uploader.WriteFile([]byte("#!/bin/sh\ntrap '' INT\necho $$ > \"$1\"\nexec sleep 600\n")))

	for _, command := range []struct {
		name string
		args []string
	}{
		{"Upload", []string{"upload", "--input-file", binary.String()}},
		{"CompileAndUpload", []string{"compile", "--upload"}},
	} {
		t.Run(command.name, func(t *testing.T) {
			pidFile := env.RootDir().Join(command.name + ".pid")
			recipe := fmt.Sprintf("tools.baz.upload.pattern=/bin/sh \"%s\" \"%s\"\n", uploader, pidFile)
			require.NoError(t, platform.Join("avr", "platform.local.txt").WriteFile([]byte(recipe)))
			args := append(command.args, "--fqbn", "foo:avr:bar", sketch.String())
			ctx, cancel := context.WithTimeout(t.Context(), 3*time.Minute)
			defer cancel()
			cmd := exec.CommandContext(ctx, integrationtest.FindArduinoCLIPath(t).String(), args...)
			cmd.Env = os.Environ()
			for key, value := range cli.GetDefaultEnv() {
				cmd.Env = append(cmd.Env, key+"="+value)
			}
			if coverageDir := os.Getenv("INTEGRATION_GOCOVERDIR"); coverageDir != "" {
				cmd.Env = append(cmd.Env, "GOCOVERDIR="+coverageDir)
			}
			// Keep this process group separate so the simulated Ctrl+C cannot reach the test runner.
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			output, err := os.CreateTemp(t.TempDir(), "output")
			require.NoError(t, err)
			defer output.Close()
			cmd.Stdout, cmd.Stderr = output, output
			require.NoError(t, cmd.Start())
			done := make(chan struct{})
			var waitErr error
			go func() {
				waitErr = cmd.Wait()
				close(done)
			}()
			defer func() {
				// Also clean up the uploader if an assertion fails against an unfixed CLI.
				if data, err := pidFile.ReadFile(); err == nil {
					if pid, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil && pid > 0 {
						_ = syscall.Kill(pid, syscall.SIGKILL)
					}
				}
				cancel()
				<-done
				if t.Failed() {
					data, _ := os.ReadFile(output.Name())
					t.Log(string(data))
				}
			}()

			var uploaderPID int
			require.Eventually(t, func() bool {
				data, err := pidFile.ReadFile()
				if err != nil {
					return false
				}
				uploaderPID, err = strconv.Atoi(strings.TrimSpace(string(data)))
				return err == nil && uploaderPID > 0
			}, 2*time.Minute, 50*time.Millisecond, "upload tool did not start")

			require.NoError(t, syscall.Kill(-cmd.Process.Pid, syscall.SIGINT))
			select {
			case <-done:
			case <-time.After(10 * time.Second):
				t.Fatal("CLI did not exit after interrupt")
			}
			require.Error(t, waitErr)
			require.ErrorIs(t, syscall.Kill(uploaderPID, 0), syscall.ESRCH, "upload tool was left running")
		})
	}
}
