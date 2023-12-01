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
	"errors"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/arduino-cli/commands/internal/instances"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/anypb"
)

// GetDebugConfig returns metadata to start debugging with the specified board
func GetDebugConfig(ctx context.Context, req *rpc.GetDebugConfigRequest) (*rpc.GetDebugConfigResponse, error) {
	pme, release := instances.GetPackageManagerExplorer(req.GetInstance())
	if pme == nil {
		return nil, &arduino.InvalidInstanceError{}
	}
	defer release()
	return getDebugProperties(req, pme, false)
}

// IsDebugSupported checks if the given board/programmer configuration supports debugging.
func IsDebugSupported(ctx context.Context, req *rpc.IsDebugSupportedRequest) (*rpc.IsDebugSupportedResponse, error) {
	pme, release := instances.GetPackageManagerExplorer(req.GetInstance())
	if pme == nil {
		return nil, &arduino.InvalidInstanceError{}
	}
	defer release()
	configRequest := &rpc.GetDebugConfigRequest{
		Instance:    req.GetInstance(),
		Fqbn:        req.GetFqbn(),
		SketchPath:  "",
		Port:        req.GetPort(),
		Interpreter: req.GetInterpreter(),
		ImportDir:   "",
		Programmer:  req.GetProgrammer(),
	}
	expectedOutput, err := getDebugProperties(configRequest, pme, true)
	var x *arduino.FailedDebugError
	if errors.As(err, &x) {
		return &rpc.IsDebugSupportedResponse{DebuggingSupported: false}, nil
	}
	if err != nil {
		return nil, err
	}

	// Compute the minimum FQBN required to get the same debug configuration.
	// (i.e. the FQBN cleaned up of the options that do not affect the debugger configuration)
	minimumFQBN := cores.MustParseFQBN(req.GetFqbn())
	for _, config := range minimumFQBN.Configs.Keys() {
		checkFQBN := minimumFQBN.Clone()
		checkFQBN.Configs.Remove(config)
		configRequest.Fqbn = checkFQBN.String()
		checkOutput, err := getDebugProperties(configRequest, pme, true)
		if err == nil && reflect.DeepEqual(expectedOutput, checkOutput) {
			minimumFQBN.Configs.Remove(config)
		}
	}
	return &rpc.IsDebugSupportedResponse{
		DebuggingSupported: true,
		DebugFqbn:          minimumFQBN.String(),
	}, nil
}

func getDebugProperties(req *rpc.GetDebugConfigRequest, pme *packagemanager.Explorer, skipSketchChecks bool) (*rpc.GetDebugConfigResponse, error) {
	var (
		sketchName             string
		sketchDefaultFQBN      string
		sketchDefaultBuildPath *paths.Path
	)
	if !skipSketchChecks {
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
		sketchName = sk.Name
		sketchDefaultFQBN = sk.GetDefaultFQBN()
		sketchDefaultBuildPath = sk.DefaultBuildPath()
	} else {
		// Use placeholder sketch data
		sketchName = "Sketch"
		sketchDefaultFQBN = ""
		sketchDefaultBuildPath = paths.New("SketchBuildPath")
	}

	// XXX Remove this code duplication!!
	fqbnIn := req.GetFqbn()
	if fqbnIn == "" {
		fqbnIn = sketchDefaultFQBN
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

	for _, tool := range pme.GetAllInstalledToolsReleases() {
		toolProperties.Merge(tool.RuntimeProperties())
	}
	if requiredTools, err := pme.FindToolsRequiredForBuild(platformRelease, referencedPlatformRelease); err == nil {
		for _, requiredTool := range requiredTools {
			logrus.WithField("tool", requiredTool).Info("Tool required for debug")
			toolProperties.Merge(requiredTool.RuntimeProperties())
		}
	}

	if req.GetProgrammer() == "" {
		return nil, &arduino.MissingProgrammerError{}
	}
	if p, ok := platformRelease.Programmers[req.GetProgrammer()]; ok {
		toolProperties.Merge(p.Properties)
	} else if refP, ok := referencedPlatformRelease.Programmers[req.GetProgrammer()]; ok {
		toolProperties.Merge(refP.Properties)
	} else {
		return nil, &arduino.ProgrammerNotFoundError{Programmer: req.GetProgrammer()}
	}

	var importPath *paths.Path
	if importDir := req.GetImportDir(); importDir != "" {
		importPath = paths.New(importDir)
	} else {
		importPath = sketchDefaultBuildPath
	}
	if !skipSketchChecks {
		if !importPath.Exist() {
			return nil, &arduino.NotFoundError{Message: tr("Compiled sketch not found in %s", importPath)}
		}
		if !importPath.IsDir() {
			return nil, &arduino.NotFoundError{Message: tr("Expected compiled sketch in directory %s, but is a file instead", importPath)}
		}
	}
	toolProperties.SetPath("build.path", importPath)
	toolProperties.Set("build.project_name", sketchName+".ino")

	// Set debug port property
	port := req.GetPort()
	if port.GetAddress() != "" {
		toolProperties.Set("debug.port", port.GetAddress())
		portFile := strings.TrimPrefix(port.GetAddress(), "/dev/")
		toolProperties.Set("debug.port.file", portFile)
	}

	// Extract and expand all debugging properties
	debugProperties := properties.NewMap()
	for k, v := range toolProperties.SubTree("debug").AsMap() {
		debugProperties.Set(k, toolProperties.ExpandPropsInString(v))
	}
	if debugAdditionalConfig, ok := toolProperties.GetOk("debug.additional_config"); ok {
		debugAdditionalConfig = toolProperties.ExpandPropsInString(debugAdditionalConfig)
		for k, v := range toolProperties.SubTree(debugAdditionalConfig).AsMap() {
			debugProperties.Set(k, toolProperties.ExpandPropsInString(v))
		}
	}

	if !debugProperties.ContainsKey("executable") || debugProperties.Get("executable") == "" {
		return nil, &arduino.FailedDebugError{Message: tr("Debugging not supported for board %s", req.GetFqbn())}
	}

	server := debugProperties.Get("server")
	toolchain := debugProperties.Get("toolchain")

	var serverConfiguration anypb.Any
	switch server {
	case "openocd":
		openocdProperties := debugProperties.SubTree("server." + server)
		scripts := openocdProperties.ExtractSubIndexLists("scripts")
		if s := openocdProperties.Get("script"); s != "" && len(scripts) == 0 {
			// backward compatibility: use "script" property if there are no "scipts.N"
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

	toolchainPrefix := debugProperties.Get("toolchain.prefix")
	// HOTFIX: for samd (and maybe some other platforms). We should keep this for a reasonable
	// amount of time to allow seamless platforms update.
	toolchainPrefix = strings.TrimSuffix(toolchainPrefix, "-")

	customConfigs := map[string]string{}
	if cortexDebugProps := debugProperties.SubTree("cortex-debug.custom"); cortexDebugProps.Size() > 0 {
		customConfigs["cortex-debug"] = convertToJsonMap(cortexDebugProps)
	}
	return &rpc.GetDebugConfigResponse{
		Executable:             debugProperties.Get("executable"),
		Server:                 server,
		ServerPath:             debugProperties.Get("server." + server + ".path"),
		ServerConfiguration:    &serverConfiguration,
		SvdFile:                debugProperties.Get("svd_file"),
		Toolchain:              toolchain,
		ToolchainPath:          debugProperties.Get("toolchain.path"),
		ToolchainPrefix:        toolchainPrefix,
		ToolchainConfiguration: &toolchainConfiguration,
		CustomConfigs:          customConfigs,
		Programmer:             req.GetProgrammer(),
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
// If a value should be converted into a JSON type different from string, the value
// may be prefiex with "[boolean]", "[number]", or "[object]":
//
//	my.stringValue=a string
//	my.booleanValue=[boolean]true
//	my.numericValue=[number]20
func convertToJsonMap(in *properties.Map) string {
	data, _ := json.MarshalIndent(convertToRawInterface(in), "", "  ")
	return string(data)
}

func allNumerics(in []string) bool {
	for _, i := range in {
		for _, c := range i {
			if c < '0' || c > '9' {
				return false
			}
		}
	}
	return true
}

func convertToRawInterface(in *properties.Map) any {
	subtrees := in.FirstLevelOf()
	keys := in.FirstLevelKeys()

	if allNumerics(keys) {
		// Compose an array
		res := []any{}
		slices.SortFunc(keys, func(x, y string) int {
			nx, _ := strconv.Atoi(x)
			ny, _ := strconv.Atoi(y)
			return nx - ny
		})
		for _, k := range keys {
			switch {
			case subtrees[k] != nil:
				res = append(res, convertToRawInterface(subtrees[k]))
			default:
				res = append(res, convertToRawValue(in.Get(k)))
			}
		}
		return res
	}

	// Compose an object
	res := map[string]any{}
	for _, k := range keys {
		switch {
		case subtrees[k] != nil:
			res[k] = convertToRawInterface(subtrees[k])
		default:
			res[k] = convertToRawValue(in.Get(k))
		}
	}
	return res
}

func convertToRawValue(v string) any {
	switch {
	case strings.HasPrefix(v, "[boolean]"):
		v = strings.TrimSpace(strings.TrimPrefix(v, "[boolean]"))
		if strings.EqualFold(v, "true") {
			return true
		} else if strings.EqualFold(v, "false") {
			return false
		}
	case strings.HasPrefix(v, "[number]"):
		v = strings.TrimPrefix(v, "[number]")
		if i, err := strconv.Atoi(v); err == nil {
			return i
		} else if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	case strings.HasPrefix(v, "[object]"):
		v = strings.TrimPrefix(v, "[object]")
		var o interface{}
		if err := json.Unmarshal([]byte(v), &o); err == nil {
			return o
		}
	case strings.HasPrefix(v, "[string]"):
		v = strings.TrimPrefix(v, "[string]")
	}
	// default or conversion error, return string as is
	return v
}
