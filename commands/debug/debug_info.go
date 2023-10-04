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
	"encoding/json"
	"regexp"
	"strings"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/anypb"
)

// GetDebugConfig returns metadata to start debugging with the specified board
func GetDebugConfig(ctx context.Context, req *rpc.GetDebugConfigRequest) (*rpc.GetDebugConfigResponse, error) {
	pme, release := commands.GetPackageManagerExplorer(req)
	if pme == nil {
		return nil, &arduino.InvalidInstanceError{}
	}
	defer release()
	return getDebugProperties(req, pme)
}

func getDebugProperties(req *rpc.GetDebugConfigRequest, pme *packagemanager.Explorer) (*rpc.GetDebugConfigResponse, error) {
	// TODO: make a generic function to extract sketch from request
	// and remove duplication in commands/compile.go
	if req.GetSketchPath() == "" {
		return nil, &arduino.MissingSketchPathError{}
	}
	sketchPath := paths.New(req.GetSketchPath())
	sk, err := sketch.New(sketchPath)
	if err != nil {
		return nil, &arduino.CantOpenSketchError{Cause: err}
	}

	// XXX Remove this code duplication!!
	fqbnIn := req.GetFqbn()
	if fqbnIn == "" && sk != nil {
		fqbnIn = sk.GetDefaultFQBN()
	}
	if fqbnIn == "" {
		return nil, &arduino.MissingFQBNError{}
	}
	fqbn, err := cores.ParseFQBN(fqbnIn)
	if err != nil {
		return nil, &arduino.InvalidFQBNError{Cause: err}
	}

	// Find target board and board properties
	_, platformRelease, _, boardProperties, referencedPlatformRelease, err := pme.ResolveFQBN(fqbn)
	if err != nil {
		return nil, &arduino.UnknownFQBNError{Cause: err}
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

	for _, tool := range pme.GetAllInstalledToolsReleases() {
		toolProperties.Merge(tool.RuntimeProperties())
	}
	if requiredTools, err := pme.FindToolsRequiredForBuild(platformRelease, referencedPlatformRelease); err == nil {
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
			return nil, &arduino.ProgrammerNotFoundError{Programmer: req.GetProgrammer()}
		}
	}

	var importPath *paths.Path
	if importDir := req.GetImportDir(); importDir != "" {
		importPath = paths.New(importDir)
	} else {
		importPath = sk.DefaultBuildPath()
	}
	if !importPath.Exist() {
		return nil, &arduino.NotFoundError{Message: tr("Compiled sketch not found in %s", importPath)}
	}
	if !importPath.IsDir() {
		return nil, &arduino.NotFoundError{Message: tr("Expected compiled sketch in directory %s, but is a file instead", importPath)}
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
		return nil, &arduino.FailedDebugError{Message: tr("Debugging not supported for board %s", req.GetFqbn())}
	}

	server := debugProperties.Get("server")
	toolchain := debugProperties.Get("toolchain")

	var serverConfiguration anypb.Any
	switch server {
	case "openocd":
		openocdProperties := debugProperties.SubTree("server." + server)
		scripts := openocdProperties.ExtractSubIndexLists("scripts")
		if s := openocdProperties.Get("script"); s != "" {
			// backward compatibility
			scripts = append(scripts, s)
		}
		openocdConf := &rpc.DebugOpenOCDServerConfiguration{
			Path:       openocdProperties.Get("path"),
			ScriptsDir: openocdProperties.Get("scripts_dir"),
			Scripts:    scripts,
		}
		if err := serverConfiguration.MarshalFrom(openocdConf); err != nil {
			return nil, err
		}
	}

	var toolchainConfiguration anypb.Any
	switch toolchain {
	case "gcc":
		gccConf := &rpc.DebugGCCToolchainConfiguration{}
		if err := toolchainConfiguration.MarshalFrom(gccConf); err != nil {
			return nil, err
		}
	}

	cortexDebugCustomJson := ""
	if cortexDebugProps := debugProperties.SubTree("cortex-debug.custom"); cortexDebugProps.Size() > 0 {
		cortexDebugCustomJson = convertToJsonMap(cortexDebugProps)
	}
	return &rpc.GetDebugConfigResponse{
		Executable:             debugProperties.Get("executable"),
		Server:                 server,
		ServerPath:             debugProperties.Get("server." + server + ".path"),
		ServerConfiguration:    &serverConfiguration,
		Toolchain:              toolchain,
		ToolchainPath:          debugProperties.Get("toolchain.path"),
		ToolchainPrefix:        debugProperties.Get("toolchain.prefix"),
		ToolchainConfiguration: &toolchainConfiguration,
		CortexDebugCustomJson:  cortexDebugCustomJson,
	}, nil
}

// Extract a JSON from a given properies.Map and converts key-indexed arrays
// like:
//
//	my.indexed.array.0=first
//	my.indexed.array.1=second
//	my.indexed.array.2=third
//
// into the corresponding JSON arrays.
func convertToJsonMap(in *properties.Map) string {
	// XXX: Maybe this method could be a good candidate for propertis.Map?

	// Find the values that should be kept as is, and the indexed arrays
	// that should be later converted into arrays.
	arraysKeys := map[string]bool{}
	stringKeys := []string{}
	trailingNumberMatcher := regexp.MustCompile(`^(.*)\.[0-9]+$`)
	for _, k := range in.Keys() {
		match := trailingNumberMatcher.FindAllStringSubmatch(k, -1)
		if len(match) > 0 && len(match[0]) > 1 {
			arraysKeys[match[0][1]] = true
		} else {
			stringKeys = append(stringKeys, k)
		}
	}

	// Compose a map that can be later marshaled into JSON keeping
	// the arrays where they are expected to be.
	res := map[string]any{}
	for _, k := range stringKeys {
		res[k] = in.Get(k)
	}
	for k := range arraysKeys {
		res[k] = in.ExtractSubIndexLists(k)
	}

	data, _ := json.MarshalIndent(res, "", "  ")
	return string(data)
}
