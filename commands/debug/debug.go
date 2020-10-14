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
	"time"

	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/executils"
	dbg "github.com/arduino/arduino-cli/rpc/debug"
	"github.com/arduino/go-properties-orderedmap"
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

	// Get tool commandLine from core recipe
	pm := commands.GetPackageManager(req.GetInstance().GetId())
	commandLine, err := getCommandLine(req, pm)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot get command line for tool")
	}

	// Transform every path to forward slashes (on Windows some tools further
	// escapes the command line so the backslash "\" gets in the way).
	for i, param := range commandLine {
		commandLine[i] = filepath.ToSlash(param)
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
	if req.GetImportFile() != "" {
		return nil, errors.New("the ImportFile parameter has been deprecated, use ImportDir instead")
	}

	toolProperties, err := getDebugProperties(req, pm)
	if err != nil {
		return nil, err
	}

	// Set debugger interpreter (default value should be "console")
	interpreter := req.GetInterpreter()
	if interpreter != "" {
		toolProperties.Set("interpreter", interpreter)
	} else {
		toolProperties.Set("interpreter", "console")
	}

	// Build recipe for tool
	recipe := toolProperties.Get("debug.pattern")

	// REMOVEME: hotfix for samd core 1.8.5/1.8.6
	if recipe == `"{path}/{cmd}" --interpreter=mi2 -ex "set pagination off" -ex 'target extended-remote | {tools.openocd.path}/{tools.openocd.cmd} -s "{tools.openocd.path}/share/openocd/scripts/" --file "{runtime.platform.path}/variants/{build.variant}/{build.openocdscript}" -c "gdb_port pipe" -c "telnet_port 0"' {build.path}/{build.project_name}.elf` {
		recipe = `"{path}/{cmd}" --interpreter={interpreter} -ex "set remotetimeout 5" -ex "set pagination off" -ex 'target extended-remote | "{tools.openocd.path}/{tools.openocd.cmd}" -s "{tools.openocd.path}/share/openocd/scripts/" --file "{runtime.platform.path}/variants/{build.variant}/{build.openocdscript}" -c "gdb_port pipe" -c "telnet_port 0"' "{build.path}/{build.project_name}.elf"`
	}

	cmdLine := toolProperties.ExpandPropsInString(recipe)
	cmdArgs, err := properties.SplitQuotedString(cmdLine, `"'`, false)
	if err != nil {
		return nil, fmt.Errorf("invalid recipe '%s': %s", recipe, err)
	}
	return cmdArgs, nil
}
