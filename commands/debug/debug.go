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

package debug

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/executils"
	dbg "github.com/arduino/arduino-cli/rpc/debug"
	"github.com/arduino/go-paths-helper"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Debug command launches a debug tool for a sketch.
// It also implements streams routing:
// gRPC In -> tool stdIn
// grpc Out <- tool stdOut
// grpc Out <- tool stdErr
// It also implements tool process lifecycle management
func Debug(ctx context.Context, req *dbg.DebugConfigReq, inStream io.Reader, out io.Writer, interrupt <-chan os.Signal) (*dbg.DebugResp, error) {

	// Get debugging command line to run debugger
	pm := commands.GetPackageManager(req.GetInstance().GetId())
	commandLine, err := getCommandLine(req, pm)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot get command line for tool")
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

	cmd, err := executils.NewProcess(commandLine...)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot execute debug tool")
	}

	// Get stdIn pipe from tool
	in, err := cmd.StdinPipe()
	if err != nil {
		return &dbg.DebugResp{Error: err.Error()}, nil
	}
	defer in.Close()

	// Merge tool StdOut and StdErr to stream them in the io.Writer passed stream
	cmd.RedirectStdoutTo(out)
	cmd.RedirectStderrTo(out)

	// Start the debug command
	if err := cmd.Start(); err != nil {
		return &dbg.DebugResp{Error: err.Error()}, nil
	}

	if interrupt != nil {
		go func() {
			for {
				if sig, ok := <-interrupt; !ok {
					break
				} else {
					cmd.Signal(sig)
				}
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
		return &dbg.DebugResp{Error: err.Error()}, nil
	}
	return &dbg.DebugResp{}, nil
}

// getCommandLine compose a debug command represented by a core recipe
func getCommandLine(req *dbg.DebugConfigReq, pm *packagemanager.PackageManager) ([]string, error) {
	debugInfo, err := getDebugProperties(req, pm)
	if err != nil {
		return nil, err
	}

	cmdArgs := []string{}
	add := func(s string) { cmdArgs = append(cmdArgs, s) }

	// Add path to GDB Client to command line
	var gdbPath *paths.Path
	switch debugInfo.GetToolchain() {
	case "gcc":
		gdbexecutable := debugInfo.ToolchainPrefix + "gdb"
		if runtime.GOOS == "windows" {
			gdbexecutable += ".exe"
		}
		gdbPath = paths.New(debugInfo.ToolchainPath).Join(gdbexecutable)
	default:
		return nil, errors.Errorf("unsupported toolchain '%s'", debugInfo.GetToolchain())
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
		serverCmd := fmt.Sprintf(`target extended-remote | "%s"`, debugInfo.ServerPath)

		if cfg := debugInfo.ServerConfiguration["scripts_dir"]; cfg != "" {
			serverCmd += fmt.Sprintf(` -s "%s"`, cfg)
		}

		if script := debugInfo.ServerConfiguration["script"]; script != "" {
			serverCmd += fmt.Sprintf(` --file "%s"`, script)
		}

		serverCmd += ` -c "gdb_port pipe"`
		serverCmd += ` -c "telnet_port 0"`

		add("-ex")
		add(serverCmd)

	default:
		return nil, errors.Errorf("unsupported gdb server '%s'", debugInfo.GetServer())
	}

	// Add executable
	add(debugInfo.Executable)

	// Transform every path to forward slashes (on Windows some tools further
	// escapes the command line so the backslash "\" gets in the way).
	for i, param := range cmdArgs {
		cmdArgs[i] = filepath.ToSlash(param)
	}

	return cmdArgs, nil
}
