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

	bldr "github.com/arduino/arduino-cli/arduino/builder"
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/sketches"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/configuration"
	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/i18n"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/arduino-cli/metrics"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	paths "github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
	"github.com/segmentio/stats/v4"
	"github.com/sirupsen/logrus"
)

// Compile FIXMEDOC
func Compile(ctx context.Context, req *rpc.CompileReq, outStream, errStream io.Writer, debug bool) (r *rpc.CompileResp, e error) {

	// There is a binding between the export binaries setting and the CLI flag to explicitly set it,
	// since we want this binding to work also for the gRPC interface we must read it here in this
	// package instead of the cli/compile one, otherwise we'd lose the binding.
	exportBinaries := configuration.Settings.GetBool("sketch.always_export_binaries")
	// If we'd just read the binding in any case, even if the request sets the export binaries setting,
	// the settings value would always overwrite the request one and it wouldn't have any effect
	// setting it for individual requests. To solve this we use a wrapper.BoolValue to handle
	// the optionality of this property, otherwise we would have no way of knowing if the property
	// was set in the request or it's just the default boolean value.
	if reqExportBinaries := req.GetExportBinaries(); reqExportBinaries != nil {
		exportBinaries = reqExportBinaries.Value
	}

	tags := map[string]string{
		"fqbn":            req.Fqbn,
		"sketchPath":      metrics.Sanitize(req.SketchPath),
		"showProperties":  strconv.FormatBool(req.ShowProperties),
		"preprocess":      strconv.FormatBool(req.Preprocess),
		"buildProperties": strings.Join(req.BuildProperties, ","),
		"warnings":        req.Warnings,
		"verbose":         strconv.FormatBool(req.Verbose),
		"quiet":           strconv.FormatBool(req.Quiet),
		"vidPid":          req.VidPid,
		"exportDir":       metrics.Sanitize(req.GetExportDir()),
		"jobs":            strconv.FormatInt(int64(req.Jobs), 10),
		"libraries":       strings.Join(req.Libraries, ","),
		"clean":           strconv.FormatBool(req.GetClean()),
		"exportBinaries":  strconv.FormatBool(exportBinaries),
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
	builderCtx.HardwareDirs = configuration.HardwareDirectories(configuration.Settings)
	builderCtx.BuiltInToolsDirs = configuration.BundleToolsDirectories(configuration.Settings)

	builderCtx.OtherLibrariesDirs = paths.NewPathList(req.GetLibraries()...)
	builderCtx.OtherLibrariesDirs.Add(configuration.LibrariesDir(configuration.Settings))

	if req.GetBuildPath() == "" {
		builderCtx.BuildPath = bldr.GenBuildPath(sketch.FullPath)
	} else {
		builderCtx.BuildPath = paths.New(req.GetBuildPath())
	}
	if err = builderCtx.BuildPath.MkdirAll(); err != nil {
		return nil, fmt.Errorf("cannot create build directory: %s", err)
	}
	builderCtx.CompilationDatabase = bldr.NewCompilationDatabase(
		builderCtx.BuildPath.Join("compile_commands.json"),
	)

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
	dataDir := paths.New(configuration.Settings.GetString("directories.Data"))
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
	builderCtx.OnlyUpdateCompilationDatabase = req.GetCreateCompilationDatabaseOnly()

	builderCtx.SourceOverride = req.GetSourceOverride()

	r = &rpc.CompileResp{}
	defer func() {
		if p := builderCtx.BuildPath; p != nil {
			r.BuildPath = p.String()
		}
	}()

	// if --preprocess or --show-properties were passed, we can stop here
	if req.GetShowProperties() {
		return r, builder.RunParseHardwareAndDumpBuildProperties(builderCtx)
	} else if req.GetPreprocess() {
		return r, builder.RunPreprocess(builderCtx)
	}

	// if it's a regular build, go on...
	if err := builder.RunBuilder(builderCtx); err != nil {
		return r, err
	}

	// If the export directory is set we assume you want to export the binaries
	if exportBinaries || req.GetExportDir() != "" {
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
			return r, errors.Wrap(err, "creating output dir")
		}

		// Copy all "sketch.ino.*" artifacts to the export directory
		baseName, ok := builderCtx.BuildProperties.GetOk("build.project_name") // == "sketch.ino"
		if !ok {
			return r, errors.New("missing 'build.project_name' build property")
		}
		buildFiles, err := builderCtx.BuildPath.ReadDir()
		if err != nil {
			return r, errors.Errorf("reading build directory: %s", err)
		}
		buildFiles.FilterPrefix(baseName)
		for _, buildFile := range buildFiles {
			exportedFile := exportPath.Join(buildFile.Base())
			logrus.
				WithField("src", buildFile).
				WithField("dest", exportedFile).
				Trace("Copying artifact.")
			if err = buildFile.CopyTo(exportedFile); err != nil {
				return r, errors.Wrapf(err, "copying output file %s", buildFile)
			}
		}
	}

	importedLibs := []*rpc.Library{}
	for _, lib := range builderCtx.ImportedLibraries {
		rpcLib, err := lib.ToRPCLibrary()
		if err != nil {
			return r, fmt.Errorf("converting library %s to rpc struct: %w", lib.Name, err)
		}
		importedLibs = append(importedLibs, rpcLib)
	}

	logrus.Tracef("Compile %s for %s successful", sketch.Name, fqbnIn)

	return &rpc.CompileResp{
		UsedLibraries:          importedLibs,
		ExecutableSectionsSize: builderCtx.ExecutableSectionsSize.ToRPCExecutableSectionSizeArray(),
	}, nil
}
