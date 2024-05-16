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

package commands

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/commands/internal/instances"
	"github.com/arduino/arduino-cli/internal/arduino/cores/packagemanager"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
)

// Debug command launches a debug tool for a sketch.
// It also implements streams routing:
// gRPC In -> tool stdIn
// grpc Out <- tool stdOut
// grpc Out <- tool stdErr
// It also implements tool process lifecycle management
func Debug(ctx context.Context, req *rpc.GetDebugConfigRequest, inStream io.Reader, out io.Writer, interrupt <-chan os.Signal) (*rpc.DebugResponse, error) {

	// Get debugging command line to run debugger
	pme, release, err := instances.GetPackageManagerExplorer(req.GetInstance())
	if err != nil {
		return nil, err
	}
	defer release()

	commandLine, err := getCommandLine(req, pme)
	if err != nil {
		return nil, err
	}

	for i, arg := range commandLine {
		fmt.Printf("%2d: %s\n", i, arg)
	}

	// Run Tool
	entry := logrus.NewEntry(logrus.StandardLogger())
	for i, param := range commandLine {
		entry = entry.WithField(fmt.Sprintf("param%d", i), param)
	}
	entry.Debug("Executing debugger")

	cmd, err := paths.NewProcess(pme.GetEnvVarsForSpawnedProcess(), commandLine...)
	if err != nil {
		return nil, &cmderrors.FailedDebugError{Message: tr("Cannot execute debug tool"), Cause: err}
	}

	// Get stdIn pipe from tool
	in, err := cmd.StdinPipe()
	if err != nil {
		return &rpc.DebugResponse{Message: &rpc.DebugResponse_Result_{
			Result: &rpc.DebugResponse_Result{Error: err.Error()},
		}}, nil
	}
	defer in.Close()

	// Merge tool StdOut and StdErr to stream them in the io.Writer passed stream
	cmd.RedirectStdoutTo(out)
	cmd.RedirectStderrTo(out)

	// Start the debug command
	if err := cmd.Start(); err != nil {
		return &rpc.DebugResponse{Message: &rpc.DebugResponse_Result_{
			Result: &rpc.DebugResponse_Result{Error: err.Error()},
		}}, nil
	}

	if interrupt != nil {
		go func() {
			for {
				sig, ok := <-interrupt
				if !ok {
					break
				}
				cmd.Signal(sig)
			}
		}()
	}

	go func() {
		// Copy data from passed inStream into command stdIn
		io.Copy(in, inStream)
		// In any case, try process termination after a second to avoid leaving
		// zombie process.
		time.Sleep(time.Second)
		cmd.Kill()
	}()

	// Wait for process to finish
	if err := cmd.Wait(); err != nil {
		return &rpc.DebugResponse{Message: &rpc.DebugResponse_Result_{
			Result: &rpc.DebugResponse_Result{Error: err.Error()},
		}}, nil
	}
	return &rpc.DebugResponse{Message: &rpc.DebugResponse_Result_{
		Result: &rpc.DebugResponse_Result{},
	}}, nil
}

// getCommandLine compose a debug command represented by a core recipe
func getCommandLine(req *rpc.GetDebugConfigRequest, pme *packagemanager.Explorer) ([]string, error) {
	debugInfo, err := getDebugProperties(req, pme, false)
	if err != nil {
		return nil, err
	}

	cmdArgs := []string{}
	add := func(s string) { cmdArgs = append(cmdArgs, s) }

	// Add path to GDB Client to command line
	var gdbPath *paths.Path
	switch debugInfo.GetToolchain() {
	case "gcc":
		gdbexecutable := debugInfo.GetToolchainPrefix() + "-gdb"
		if runtime.GOOS == "windows" {
			gdbexecutable += ".exe"
		}
		gdbPath = paths.New(debugInfo.GetToolchainPath()).Join(gdbexecutable)
	default:
		return nil, &cmderrors.FailedDebugError{Message: tr("Toolchain '%s' is not supported", debugInfo.GetToolchain())}
	}
	add(gdbPath.String())

	// Set GDB interpreter (default value should be "console")
	gdbInterpreter := req.GetInterpreter()
	if gdbInterpreter == "" {
		gdbInterpreter = "console"
	}
	add("--interpreter=" + gdbInterpreter)
	if gdbInterpreter != "console" {
		add("-ex")
		add("set pagination off")
	}

	// Add extra GDB execution commands
	add("-ex")
	add("set remotetimeout 5")

	// Extract path to GDB Server
	switch debugInfo.GetServer() {
	case "openocd":
		var openocdConf rpc.DebugOpenOCDServerConfiguration
		if err := debugInfo.GetServerConfiguration().UnmarshalTo(&openocdConf); err != nil {
			return nil, err
		}

		serverCmd := fmt.Sprintf(`target extended-remote | "%s"`, debugInfo.GetServerPath())

		if cfg := openocdConf.GetScriptsDir(); cfg != "" {
			serverCmd += fmt.Sprintf(` -s "%s"`, cfg)
		}

		for _, script := range openocdConf.GetScripts() {
			serverCmd += fmt.Sprintf(` --file "%s"`, script)
		}

		serverCmd += ` -c "gdb_port pipe"`
		serverCmd += ` -c "telnet_port 0"`

		add("-ex")
		add(serverCmd)

	default:
		return nil, &cmderrors.FailedDebugError{Message: tr("GDB server '%s' is not supported", debugInfo.GetServer())}
	}

	// Add executable
	add(debugInfo.GetExecutable())

	// Transform every path to forward slashes (on Windows some tools further
	// escapes the command line so the backslash "\" gets in the way).
	for i, param := range cmdArgs {
		cmdArgs[i] = filepath.ToSlash(param)
	}

	return cmdArgs, nil
}
