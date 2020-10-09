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

package compile

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/sketches"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/configuration"
	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/i18n"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/arduino/arduino-cli/telemetry"
	paths "github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
	"github.com/segmentio/stats/v4"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Compile FIXMEDOC
func Compile(ctx context.Context, req *rpc.CompileReq, outStream, errStream io.Writer, debug bool) (r *rpc.CompileResp, e error) {

	tags := map[string]string{
		"fqbn":            req.Fqbn,
		"sketchPath":      telemetry.Sanitize(req.SketchPath),
		"showProperties":  strconv.FormatBool(req.ShowProperties),
		"preprocess":      strconv.FormatBool(req.Preprocess),
		"buildProperties": strings.Join(req.BuildProperties, ","),
		"warnings":        req.Warnings,
		"verbose":         strconv.FormatBool(req.Verbose),
		"quiet":           strconv.FormatBool(req.Quiet),
		"vidPid":          req.VidPid,
		"exportFile":      telemetry.Sanitize(req.ExportFile), // deprecated
		"exportDir":       telemetry.Sanitize(req.GetExportDir()),
		"jobs":            strconv.FormatInt(int64(req.Jobs), 10),
		"libraries":       strings.Join(req.Libraries, ","),
		"clean":           strconv.FormatBool(req.GetClean()),
	}

	if req.GetExportFile() != "" {
		outStream.Write([]byte(fmt.Sprintln("Compile.ExportFile has been deprecated. The ExportFile parameter will be ignored, use ExportDir instead.")))
	}

	// Use defer func() to evaluate tags map when function returns
	// and set success flag inspecting the error named return parameter
	defer func() {
		tags["success"] = "true"
		if e != nil {
			tags["success"] = "false"
		}
		stats.Incr("compile", stats.M(tags)...)
	}()

	pm := commands.GetPackageManager(req.GetInstance().GetId())
	if pm == nil {
		return nil, errors.New("invalid instance")
	}

	logrus.Tracef("Compile %s for %s started", req.GetSketchPath(), req.GetFqbn())
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
		// feedback.Error(errorMessage)
		return nil, fmt.Errorf("platform not installed")
	}

	builderCtx := &types.Context{}
	builderCtx.PackageManager = pm
	builderCtx.FQBN = fqbn
	builderCtx.SketchLocation = sketch.FullPath

	// FIXME: This will be redundant when arduino-builder will be part of the cli
	builderCtx.HardwareDirs = configuration.HardwareDirectories()
	builderCtx.BuiltInToolsDirs = configuration.BundleToolsDirectories()

	builderCtx.OtherLibrariesDirs = paths.NewPathList(req.GetLibraries()...)
	builderCtx.OtherLibrariesDirs.Add(configuration.LibrariesDir())

	if req.GetBuildPath() != "" {
		builderCtx.BuildPath = paths.New(req.GetBuildPath())
		err = builderCtx.BuildPath.MkdirAll()
		if err != nil {
			return nil, fmt.Errorf("cannot create build directory: %s", err)
		}
	}

	builderCtx.Verbose = req.GetVerbose()

	// Optimize for debug
	builderCtx.OptimizeForDebug = req.GetOptimizeForDebug()

	builderCtx.CoreBuildCachePath = paths.TempDir().Join("arduino-core-cache")

	builderCtx.Jobs = int(req.GetJobs())

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
	dataDir := paths.New(viper.GetString("directories.Data"))
	preferencesTxt := dataDir.Join("preferences.txt")
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
	builderCtx.Clean = req.GetClean()

	// if --preprocess or --show-properties were passed, we can stop here
	if req.GetShowProperties() {
		return &rpc.CompileResp{}, builder.RunParseHardwareAndDumpBuildProperties(builderCtx)
	} else if req.GetPreprocess() {
		return &rpc.CompileResp{}, builder.RunPreprocess(builderCtx)
	}

	// if it's a regular build, go on...
	if err := builder.RunBuilder(builderCtx); err != nil {
		return nil, err
	}

	if !req.GetDryRun() {
		var exportPath *paths.Path
		if exportDir := req.GetExportDir(); exportDir != "" {
			exportPath = paths.New(exportDir)
		} else {
			exportPath = sketch.FullPath
			// Add FQBN (without configs part) to export path
			fqbnSuffix := strings.Replace(fqbn.StringWithoutConfig(), ":", ".", -1)
			exportPath = exportPath.Join("build").Join(fqbnSuffix)
		}
		logrus.WithField("path", exportPath).Trace("Saving sketch to export path.")
		if err := exportPath.MkdirAll(); err != nil {
			return nil, errors.Wrap(err, "creating output dir")
		}

		// Copy all "sketch.ino.*" artifacts to the export directory
		baseName, ok := builderCtx.BuildProperties.GetOk("build.project_name") // == "sketch.ino"
		if !ok {
			return nil, errors.New("missing 'build.project_name' build property")
		}
		buildFiles, err := builderCtx.BuildPath.ReadDir()
		if err != nil {
			return nil, errors.Errorf("reading build directory: %s", err)
		}
		buildFiles.FilterPrefix(baseName)
		for _, buildFile := range buildFiles {
			exportedFile := exportPath.Join(buildFile.Base())
			logrus.
				WithField("src", buildFile).
				WithField("dest", exportedFile).
				Trace("Copying artifact.")
			if err = buildFile.CopyTo(exportedFile); err != nil {
				return nil, errors.Wrapf(err, "copying output file %s", buildFile)
			}
		}
	}

	logrus.Tracef("Compile %s for %s successful", sketch.Name, fqbnIn)
	return &rpc.CompileResp{}, nil
}
