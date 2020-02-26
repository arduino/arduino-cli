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
	"strings"
	"time"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/sketches"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/executils"
	dbg "github.com/arduino/arduino-cli/rpc/debug"
	"github.com/arduino/go-paths-helper"
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

	cmd, err := executils.Command(commandLine)
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
	cmd.Stdout = out
	cmd.Stderr = out

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
					cmd.Process.Signal(sig)
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
		cmd.Process.Kill()
	}()

	// Wait for process to finish
	if err := cmd.Wait(); err != nil {
		return &dbg.DebugResp{Error: err.Error()}, nil
	}
	return &dbg.DebugResp{}, nil
}

// getCommandLine compose a debug command represented by a core recipe
func getCommandLine(req *dbg.DebugConfigReq, pm *packagemanager.PackageManager) ([]string, error) {
	// TODO: make a generic function to extract sketch from request
	// and remove duplication in commands/compile.go
	if req.GetSketchPath() == "" {
		return nil, fmt.Errorf("missing sketchPath")
	}
	sketchPath := paths.New(req.GetSketchPath())
	sketch, err := sketches.NewSketchFromPath(sketchPath)
	if err != nil {
		return nil, errors.Wrap(err, "opening sketch")
	}

	// FIXME: make a specification on how a port is specified via command line
	port := req.GetPort()
	if port == "" {
		return nil, fmt.Errorf("no debug port provided")
	}

	fqbnIn := req.GetFqbn()
	if fqbnIn == "" && sketch != nil && sketch.Metadata != nil {
		fqbnIn = sketch.Metadata.CPU.Fqbn
	}
	if fqbnIn == "" {
		return nil, fmt.Errorf("no Fully Qualified Board Name provided")
	}
	fqbn, err := cores.ParseFQBN(fqbnIn)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing FQBN")
	}

	// Find target board and board properties
	_, _, board, boardProperties, _, err := pm.ResolveFQBN(fqbn)
	if err != nil {
		return nil, errors.Wrap(err, "error resolving FQBN")
	}

	// Load programmer tool
	toolName, have := boardProperties.GetOk("debug.tool")
	if !have || toolName == "" {
		return nil, fmt.Errorf("cannot get programmer tool: undefined 'debug.tool' property")
	}

	var referencedPlatformRelease *cores.PlatformRelease
	if split := strings.Split(toolName, ":"); len(split) > 2 {
		return nil, fmt.Errorf("invalid 'debug.tool' property: %s", toolName)
	} else if len(split) == 2 {
		referencedPackageName := split[0]
		toolName = split[1]
		architecture := board.PlatformRelease.Platform.Architecture

		if referencedPackage := pm.Packages[referencedPackageName]; referencedPackage == nil {
			return nil, fmt.Errorf("required platform %s:%s not installed", referencedPackageName, architecture)
		} else if referencedPlatform := referencedPackage.Platforms[architecture]; referencedPlatform == nil {
			return nil, fmt.Errorf("required platform %s:%s not installed", referencedPackageName, architecture)
		} else {
			referencedPlatformRelease = pm.GetInstalledPlatformRelease(referencedPlatform)
		}
	}

	// Build configuration for debug
	toolProperties := properties.NewMap()
	if referencedPlatformRelease != nil {
		toolProperties.Merge(referencedPlatformRelease.Properties)
	}
	toolProperties.Merge(board.PlatformRelease.Properties)
	toolProperties.Merge(board.PlatformRelease.RuntimeProperties())
	toolProperties.Merge(boardProperties)

	requestedToolProperties := toolProperties.SubTree("tools." + toolName)
	toolProperties.Merge(requestedToolProperties)
	if requiredTools, err := pm.FindToolsRequiredForBoard(board); err == nil {
		for _, requiredTool := range requiredTools {
			logrus.WithField("tool", requiredTool).Info("Tool required for debug")
			toolProperties.Merge(requiredTool.RuntimeProperties())
		}
	}

	// Set path to compiled binary
	// Make the filename without the FQBN configs part
	fqbn.Configs = properties.NewMap()
	fqbnSuffix := strings.Replace(fqbn.String(), ":", ".", -1)

	var importPath *paths.Path
	var importFile string
	if req.GetImportFile() == "" {
		importPath = sketch.FullPath
		importFile = sketch.Name + "." + fqbnSuffix
	} else {
		importPath = paths.New(req.GetImportFile()).Parent()
		importFile = paths.New(req.GetImportFile()).Base()
	}

	outputTmpFile, ok := toolProperties.GetOk("recipe.output.tmp_file")
	outputTmpFile = toolProperties.ExpandPropsInString(outputTmpFile)
	if !ok {
		return nil, fmt.Errorf("property 'recipe.output.tmp_file' not defined")
	}
	ext := filepath.Ext(outputTmpFile)
	if strings.HasSuffix(importFile, ext) {
		importFile = importFile[:len(importFile)-len(ext)]
	}

	toolProperties.SetPath("build.path", importPath)
	toolProperties.Set("build.project_name", importFile)
	uploadFile := importPath.Join(importFile + ext)
	if _, err := uploadFile.Stat(); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("compiled sketch %s not found", uploadFile.String())
		}
		return nil, errors.Wrap(err, "cannot open sketch")
	}

	// Set debug port property
	toolProperties.Set("debug.port", port)
	if strings.HasPrefix(port, "/dev/") {
		toolProperties.Set("debug.port.file", port[5:])
	} else {
		toolProperties.Set("debug.port.file", port)
	}

	// Build recipe for tool
	recipe := toolProperties.Get("debug.pattern")
	// REMOVEME: hotfix for samd core 1.8.5
	if recipe == `"{path}/{cmd}" --interpreter=mi2 -ex "set pagination off" -ex 'target extended-remote | {tools.openocd.path}/{tools.openocd.cmd} -s "{tools.openocd.path}/share/openocd/scripts/" --file "{runtime.platform.path}/variants/{build.variant}/{build.openocdscript}" -c "gdb_port pipe" -c "telnet_port 0"' {build.path}/{build.project_name}.elf` {
		recipe = `"{path}/{cmd}" --interpreter=mi2 -ex "set remotetimeout 5" -ex "set pagination off" -ex 'target extended-remote | "{tools.openocd.path}/{tools.openocd.cmd}" -s "{tools.openocd.path}/share/openocd/scripts/" --file "{runtime.platform.path}/variants/{build.variant}/{build.openocdscript}" -c "gdb_port pipe" -c "telnet_port 0"' "{build.path}/{build.project_name}.elf"`
	}
	cmdLine := toolProperties.ExpandPropsInString(recipe)
	cmdArgs, err := properties.SplitQuotedString(cmdLine, `"'`, false)
	if err != nil {
		return nil, fmt.Errorf("invalid recipe '%s': %s", recipe, err)
	}
	return cmdArgs, nil
}
