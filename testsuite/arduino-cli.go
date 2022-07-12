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

package integrationtest

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/executils"
	"github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/fatih/color"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// ArduinoCLI is an Arduino CLI client.
type ArduinoCLI struct {
	path          *paths.Path
	t             *require.Assertions
	proc          *executils.Process
	cliEnvVars    []string
	cliConfigPath *paths.Path
	daemonAddr    string
	daemonConn    *grpc.ClientConn
	daemonClient  commands.ArduinoCoreServiceClient
}

// NewArduinoCliWithinEnvironment creates a new Arduino CLI client inside the given environment.
func NewArduinoCliWithinEnvironment(t *testing.T, cliPath *paths.Path, env *Environment) *ArduinoCLI {
	cli := NewArduinoCli(t, cliPath)
	cli.cliEnvVars = []string{
		fmt.Sprintf("ARDUINO_DATA_DIR=%s", env.Root().Join("arduino15")),
		fmt.Sprintf("ARDUINO_DOWNLOADS_DIR=%s", env.Root().Join("arduino15/staging")),
		fmt.Sprintf("ARDUINO_SKETCHBOOK_DIR=%s", env.Root().Join("Arduino")),
	}
	return cli
}

// NewArduinoCli creates a new Arduino CLI client.
func NewArduinoCli(t *testing.T, cliPath *paths.Path) *ArduinoCLI {
	return &ArduinoCLI{
		path: cliPath,
		t:    require.New(t),
	}
}

// CleanUp closes the Arduino CLI client.
func (cli *ArduinoCLI) CleanUp() {
	if cli.proc != nil {
		cli.proc.Kill()
		cli.proc.Wait()
	}
}

// Run executes the given arduino-cli command and returns the output.
func (cli *ArduinoCLI) Run(args ...string) ([]byte, []byte, error) {
	if cli.cliConfigPath != nil {
		args = append([]string{"--config-file", cli.cliConfigPath.String()}, args...)
	}
	fmt.Println(color.HiBlackString(">>> Running: ") + color.HiYellowString("%s %s", cli.path, strings.Join(args, " ")))
	cliProc, err := executils.NewProcessFromPath(cli.cliEnvVars, cli.path, args...)
	cli.t.NoError(err)
	stdout, err := cliProc.StdoutPipe()
	cli.t.NoError(err)
	stderr, err := cliProc.StderrPipe()
	cli.t.NoError(err)
	_, err = cliProc.StdinPipe()
	cli.t.NoError(err)

	cli.t.NoError(cliProc.Start())

	var stdoutBuf, stderrBuf bytes.Buffer
	stdoutCtx, stdoutCancel := context.WithCancel(context.Background())
	stderrCtx, stderrCancel := context.WithCancel(context.Background())
	go func() {
		io.Copy(&stdoutBuf, io.TeeReader(stdout, os.Stdout))
		stdoutCancel()
	}()
	go func() {
		io.Copy(&stderrBuf, io.TeeReader(stderr, os.Stderr))
		stderrCancel()
	}()
	cliErr := cliProc.Wait()
	<-stdoutCtx.Done()
	<-stderrCtx.Done()
	fmt.Println(color.HiBlackString("<<< Run completed (err = %v)", cliErr))

	return stdoutBuf.Bytes(), stderrBuf.Bytes(), cliErr
}
