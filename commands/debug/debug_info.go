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
	"strings"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/rpc/cc/arduino/cli/debug/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/sirupsen/logrus"
)

// GetDebugConfig returns metadata to start debugging with the specified board
func GetDebugConfig(ctx context.Context, req *debug.DebugConfigRequest) (*debug.GetDebugConfigResponse, error) {
	pm := commands.GetPackageManager(req.GetInstance().GetId())
	return getDebugProperties(req, pm)
}

func getDebugProperties(req *debug.DebugConfigRequest, pm *packagemanager.PackageManager) (*debug.GetDebugConfigResponse, error) {
	// TODO: make a generic function to extract sketch from request
	// and remove duplication in commands/compile.go
	if req.GetSketchPath() == "" {
		return nil, &commands.MissingSketchPathError{}
	}
	sketchPath := paths.New(req.GetSketchPath())
	sk, err := sketch.New(sketchPath)
	if err != nil {
		return nil, &commands.CantOpenSketchError{Cause: err}
	}

	// XXX Remove this code duplication!!
	fqbnIn := req.GetFqbn()
	if fqbnIn == "" && sk != nil && sk.Metadata != nil {
		fqbnIn = sk.Metadata.CPU.Fqbn
	}
	if fqbnIn == "" {
		return nil, &commands.MissingFQBNError{}
	}
	fqbn, err := cores.ParseFQBN(fqbnIn)
	if err != nil {
		return nil, &commands.InvalidFQBNError{Cause: err}
	}

	// Find target board and board properties
	_, platformRelease, board, boardProperties, referencedPlatformRelease, err := pm.ResolveFQBN(fqbn)
	if err != nil {
		return nil, &commands.UnknownFQBNError{Cause: err}
	}

	// Build configuration for debug
	toolProperties := properties.NewMap()
	if referencedPlatformRelease != nil {
		toolProperties.Merge(referencedPlatformRelease.Properties)
	}
	toolProperties.Merge(platformRelease.Properties)
	toolProperties.Merge(platformRelease.RuntimeProperties())
	toolProperties.Merge(boardProperties)

	// HOTFIX: Remove me when the `arduino:samd` core is updated
	//         (remember to remove it also in arduino/board/details.go)
	if !toolProperties.ContainsKey("debug.executable") {
		if platformRelease.String() == "arduino:samd@1.8.9" || platformRelease.String() == "arduino:samd@1.8.8" {
			toolProperties.Set("debug.executable", "{build.path}/{build.project_name}.elf")
			toolProperties.Set("debug.toolchain", "gcc")
			toolProperties.Set("debug.toolchain.path", "{runtime.tools.arm-none-eabi-gcc-7-2017q4.path}/bin/")
			toolProperties.Set("debug.toolchain.prefix", "arm-none-eabi-")
			toolProperties.Set("debug.server", "openocd")
			toolProperties.Set("debug.server.openocd.path", "{runtime.tools.openocd-0.10.0-arduino7.path}/bin/openocd")
			toolProperties.Set("debug.server.openocd.scripts_dir", "{runtime.tools.openocd-0.10.0-arduino7.path}/share/openocd/scripts/")
			toolProperties.Set("debug.server.openocd.script", "{runtime.platform.path}/variants/{build.variant}/{build.openocdscript}")
		}
	}

	for _, tool := range pm.GetAllInstalledToolsReleases() {
		toolProperties.Merge(tool.RuntimeProperties())
	}
	if requiredTools, err := pm.FindToolsRequiredForBoard(board); err == nil {
		for _, requiredTool := range requiredTools {
			logrus.WithField("tool", requiredTool).Info("Tool required for debug")
			toolProperties.Merge(requiredTool.RuntimeProperties())
		}
	}

	if req.GetProgrammer() != "" {
		if p, ok := platformRelease.Programmers[req.GetProgrammer()]; ok {
			toolProperties.Merge(p.Properties)
		} else if refP, ok := referencedPlatformRelease.Programmers[req.GetProgrammer()]; ok {
			toolProperties.Merge(refP.Properties)
		} else {
			return nil, &commands.ProgrammerNotFoundError{Programmer: req.GetProgrammer()}
		}
	}

	importPath := sk.BuildPath
	if importDir := req.GetImportDir(); importDir != "" {
		importPath = paths.New(importDir)
	}
	if !importPath.Exist() {
		return nil, &commands.NotFoundError{Message: tr("Compiled sketch not found in %s", importPath)}
	}
	if !importPath.IsDir() {
		return nil, &commands.NotFoundError{Message: tr("Expected compiled sketch in directory %s, but is a file instead", importPath)}
	}
	toolProperties.SetPath("build.path", importPath)
	toolProperties.Set("build.project_name", sk.Name+".ino")

	// Set debug port property
	port := req.GetPort()
	if port.GetAddress() != "" {
		toolProperties.Set("debug.port", port.Address)
		portFile := strings.TrimPrefix(port.Address, "/dev/")
		toolProperties.Set("debug.port.file", portFile)
	}

	// Extract and expand all debugging properties
	debugProperties := properties.NewMap()
	for k, v := range toolProperties.SubTree("debug").AsMap() {
		debugProperties.Set(k, toolProperties.ExpandPropsInString(v))
	}

	if !debugProperties.ContainsKey("executable") {
		return nil, &commands.FailedDebugError{Message: tr("Debugging not supported for board %s", req.GetFqbn())}
	}

	server := debugProperties.Get("server")
	toolchain := debugProperties.Get("toolchain")
	return &debug.GetDebugConfigResponse{
		Executable:             debugProperties.Get("executable"),
		Server:                 server,
		ServerPath:             debugProperties.Get("server." + server + ".path"),
		ServerConfiguration:    debugProperties.SubTree("server." + server).AsMap(),
		Toolchain:              toolchain,
		ToolchainPath:          debugProperties.Get("toolchain.path"),
		ToolchainPrefix:        debugProperties.Get("toolchain.prefix"),
		ToolchainConfiguration: debugProperties.SubTree("toolchain." + toolchain).AsMap(),
	}, nil
}
