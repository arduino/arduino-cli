//
// This file is part of arduino-cli.
//
// Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to modify or
// otherwise use the software for commercial activities involving the Arduino
// software without disclosing the source code of your own applications. To purchase
// a commercial license, send an email to license@arduino.cc.
//

package compile

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/sketches"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/configs"
	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/i18n"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	paths "github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
	"github.com/sirupsen/logrus"
)

// Compile FIXMEDOC
func Compile(ctx context.Context, req *rpc.CompileReq, outStream, errStream io.Writer, config *configs.Configuration, debug bool) (*rpc.CompileResp, error) {
	pm := commands.GetPackageManager(req)
	if pm == nil {
		return nil, errors.New("invalid instance")
	}

	logrus.Info("Executing `arduino compile`")
	if req.GetSketchPath() == "" {
		return nil, fmt.Errorf("missing sketchPath")
	}
	sketchPath := paths.New(req.GetSketchPath())
	sketch, err := sketches.NewSketchFromPath(sketchPath)
	if err != nil {
		return nil, fmt.Errorf("opening sketch: %s", err)
	}

	fqbnIn := req.GetFqbn()
	if fqbnIn == "" && sketch != nil && sketch.Metadata != nil {
		fqbnIn = sketch.Metadata.CPU.Fqbn
	}
	if fqbnIn == "" {
		return nil, fmt.Errorf("no FQBN provided")
	}
	fqbn, err := cores.ParseFQBN(fqbnIn)
	if err != nil {
		return nil, fmt.Errorf("incorrect FQBN: %s", err)
	}

	targetPlatform := pm.FindPlatform(&packagemanager.PlatformReference{
		Package:              fqbn.Package,
		PlatformArchitecture: fqbn.PlatformArch,
	})
	if targetPlatform == nil || pm.GetInstalledPlatformRelease(targetPlatform) == nil {
		// TODO: Move this error message in `cli` module
		// errorMessage := fmt.Sprintf(
		// 	"\"%[1]s:%[2]s\" platform is not installed, please install it by running \""+
		// 		version.GetAppName()+" core install %[1]s:%[2]s\".", fqbn.Package, fqbn.PlatformArch)
		// formatter.PrintErrorMessage(errorMessage)
		return nil, fmt.Errorf("platform not installed")
	}

	builderCtx := &types.Context{}
	builderCtx.PackageManager = pm
	builderCtx.FQBN = fqbn
	builderCtx.SketchLocation = sketch.FullPath

	// FIXME: This will be redundant when arduino-builder will be part of the cli
	if packagesDir, err := config.HardwareDirectories(); err == nil {
		builderCtx.HardwareDirs = packagesDir
	} else {
		return nil, fmt.Errorf("cannot get hardware directories: %s", err)
	}

	if toolsDir, err := config.BundleToolsDirectories(); err == nil {
		builderCtx.ToolsDirs = toolsDir
	} else {
		return nil, fmt.Errorf("cannot get bundled tools directories: %s", err)
	}

	builderCtx.OtherLibrariesDirs = paths.NewPathList()
	builderCtx.OtherLibrariesDirs.Add(config.LibrariesDir())

	if req.GetBuildPath() != "" {
		builderCtx.BuildPath = paths.New(req.GetBuildPath())
		err = builderCtx.BuildPath.MkdirAll()
		if err != nil {
			return nil, fmt.Errorf("cannot create build directory: %s", err)
		}
	}

	builderCtx.Verbose = req.GetVerbose()

	builderCtx.CoreBuildCachePath = paths.TempDir().Join("arduino-core-cache")

	builderCtx.USBVidPid = req.GetVidPid()
	builderCtx.WarningsLevel = req.GetWarnings()

	if debug {
		builderCtx.DebugLevel = 100
	} else {
		builderCtx.DebugLevel = 5
	}

	builderCtx.CustomBuildProperties = append(req.GetBuildProperties(), "build.warn_data_percentage=75")

	if req.GetBuildCachePath() != "" {
		builderCtx.BuildCachePath = paths.New(req.GetBuildCachePath())
		err = builderCtx.BuildCachePath.MkdirAll()
		if err != nil {
			return nil, fmt.Errorf("cannot create build cache directory: %s", err)
		}
	}

	// Will be deprecated.
	builderCtx.ArduinoAPIVersion = "10607"

	// Check if Arduino IDE is installed and get it's libraries location.
	preferencesTxt := config.DataDir.Join("preferences.txt")
	ideProperties, err := properties.LoadFromPath(preferencesTxt)
	if err == nil {
		lastIdeSubProperties := ideProperties.SubTree("last").SubTree("ide")
		// Preferences can contain records from previous IDE versions. Find the latest one.
		var pathVariants []string
		for k := range lastIdeSubProperties.AsMap() {
			if strings.HasSuffix(k, ".hardwarepath") {
				pathVariants = append(pathVariants, k)
			}
		}
		sort.Strings(pathVariants)
		ideHardwarePath := lastIdeSubProperties.Get(pathVariants[len(pathVariants)-1])
		ideLibrariesPath := filepath.Join(filepath.Dir(ideHardwarePath), "libraries")
		builderCtx.BuiltInLibrariesDirs = paths.NewPathList(ideLibrariesPath)
	}

	builderCtx.ExecStdout = outStream
	builderCtx.ExecStderr = errStream
	builderCtx.SetLogger(i18n.LoggerToCustomStreams{Stdout: outStream, Stderr: errStream})
	if req.GetShowProperties() {
		err = builder.RunParseHardwareAndDumpBuildProperties(builderCtx)
	} else if req.GetPreprocess() {
		err = builder.RunPreprocess(builderCtx)
	} else {
		err = builder.RunBuilder(builderCtx)
	}

	if err != nil {
		return nil, fmt.Errorf("build failed: %s", err)
	}

	// FIXME: Make a function to obtain these info...
	outputPath := builderCtx.BuildProperties.ExpandPropsInString("{build.path}/{recipe.output.tmp_file}")
	ext := filepath.Ext(outputPath)

	// FIXME: Make a function to produce a better name...
	// Make the filename without the FQBN configs part
	fqbn.Configs = properties.NewMap()
	fqbnSuffix := strings.Replace(fqbn.String(), ":", ".", -1)

	var exportPath *paths.Path
	var exportFile string
	if req.GetExportFile() == "" {
		if sketch.FullPath.IsDir() {
			exportPath = sketch.FullPath
		} else {
			exportPath = sketch.FullPath.Parent()
		}
		exportFile = sketch.Name + "." + fqbnSuffix
	} else {
		exportPath = paths.New(req.GetExportFile()).Parent()
		exportFile = paths.New(req.GetExportFile()).Base()
		if strings.HasSuffix(exportFile, ext) {
			exportFile = exportFile[:len(exportFile)-len(ext)]
		}
	}

	// Copy .hex file to sketch directory
	srcHex := paths.New(outputPath)
	dstHex := exportPath.Join(exportFile + ext)
	logrus.WithField("from", srcHex).WithField("to", dstHex).Print("copying sketch build output")
	if err = srcHex.CopyTo(dstHex); err != nil {
		return nil, fmt.Errorf("copying output file: %s", err)
	}

	// Copy .elf file to sketch directory
	srcElf := paths.New(outputPath[:len(outputPath)-3] + "elf")
	dstElf := exportPath.Join(exportFile + ".elf")
	logrus.WithField("from", srcElf).WithField("to", dstElf).Print("copying sketch build output")
	if err = srcElf.CopyTo(dstElf); err != nil {
		return nil, fmt.Errorf("copying elf file: %s", err)
	}

	return &rpc.CompileResp{}, nil
}
