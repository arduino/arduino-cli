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
	"strings"

	"github.com/arduino/arduino-cli/arduino"
	bldr "github.com/arduino/arduino-cli/arduino/builder"
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/configuration"
	"github.com/arduino/arduino-cli/i18n"
	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	paths "github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
)

var tr = i18n.Tr

// Compile FIXMEDOC
func Compile(ctx context.Context, req *rpc.CompileRequest, outStream, errStream io.Writer, progressCB rpc.TaskProgressCB, debug bool) (r *rpc.CompileResponse, e error) {

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

	pme, release := commands.GetPackageManagerExplorer(req)
	if pme == nil {
		return nil, &arduino.InvalidInstanceError{}
	}
	defer release()

	lm := commands.GetLibraryManager(req)
	if lm == nil {
		return nil, &arduino.InvalidInstanceError{}
	}

	logrus.Tracef("Compile %s for %s started", req.GetSketchPath(), req.GetFqbn())
	if req.GetSketchPath() == "" {
		return nil, &arduino.MissingSketchPathError{}
	}
	sketchPath := paths.New(req.GetSketchPath())
	sk, err := sketch.New(sketchPath)
	if err != nil {
		return nil, &arduino.CantOpenSketchError{Cause: err}
	}

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

	targetPlatform := pme.FindPlatform(&packagemanager.PlatformReference{
		Package:              fqbn.Package,
		PlatformArchitecture: fqbn.PlatformArch,
	})
	if targetPlatform == nil || pme.GetInstalledPlatformRelease(targetPlatform) == nil {
		return nil, &arduino.PlatformNotFoundError{
			Platform: fmt.Sprintf("%s:%s", fqbn.Package, fqbn.PlatformArch),
			Cause:    fmt.Errorf(tr("platform not installed")),
		}
	}

	// At the current time we do not have a way of knowing if a board supports the secure boot or not,
	// so, if the flags to override the default keys are used, we try override the corresponding platform property nonetheless.
	// It's not possible to use the default name for the keys since there could be more tools to sign and encrypt.
	// So it's mandatory to use all three flags to sign and encrypt the binary
	securityKeysOverride := []string{}
	if req.KeysKeychain != "" && req.SignKey != "" && req.EncryptKey != "" {
		securityKeysOverride = append(securityKeysOverride, "build.keys.keychain="+req.KeysKeychain, "build.keys.sign_key="+req.GetSignKey(), "build.keys.encrypt_key="+req.EncryptKey)
	}

	builderCtx := &types.Context{}
	builderCtx.PackageManager = pme
	if pme.GetProfile() != nil {
		builderCtx.LibrariesManager = lm
	}
	builderCtx.UseCachedLibrariesResolution = req.GetSkipLibrariesDiscovery()
	builderCtx.FQBN = fqbn
	builderCtx.SketchLocation = sk.FullPath
	builderCtx.ProgressCB = progressCB

	// FIXME: This will be redundant when arduino-builder will be part of the cli
	builderCtx.HardwareDirs = configuration.HardwareDirectories(configuration.Settings)
	builderCtx.BuiltInToolsDirs = configuration.BuiltinToolsDirectories(configuration.Settings)

	// FIXME: This will be redundant when arduino-builder will be part of the cli
	builderCtx.OtherLibrariesDirs = paths.NewPathList(req.GetLibraries()...)
	builderCtx.OtherLibrariesDirs.Add(configuration.LibrariesDir(configuration.Settings))
	builderCtx.LibraryDirs = paths.NewPathList(req.Library...)
	if req.GetBuildPath() == "" {
		builderCtx.BuildPath = sk.BuildPath
	} else {
		builderCtx.BuildPath = paths.New(req.GetBuildPath()).Canonical()
	}
	if err = builderCtx.BuildPath.MkdirAll(); err != nil {
		return nil, &arduino.PermissionDeniedError{Message: tr("Cannot create build directory"), Cause: err}
	}
	builderCtx.CompilationDatabase = bldr.NewCompilationDatabase(
		builderCtx.BuildPath.Join("compile_commands.json"),
	)

	builderCtx.Verbose = req.GetVerbose()

	// Optimize for debug
	builderCtx.OptimizeForDebug = req.GetOptimizeForDebug()

	builderCtx.CoreBuildCachePath = paths.TempDir().Join("arduino", "core-cache")

	builderCtx.Jobs = int(req.GetJobs())

	builderCtx.USBVidPid = req.GetVidPid()
	builderCtx.WarningsLevel = req.GetWarnings()

	builderCtx.CustomBuildProperties = append(req.GetBuildProperties(), "build.warn_data_percentage=75")
	builderCtx.CustomBuildProperties = append(req.GetBuildProperties(), securityKeysOverride...)

	if req.GetBuildCachePath() != "" {
		builderCtx.BuildCachePath = paths.New(req.GetBuildCachePath())
		err = builderCtx.BuildCachePath.MkdirAll()
		if err != nil {
			return nil, &arduino.PermissionDeniedError{Message: tr("Cannot create build cache directory"), Cause: err}
		}
	}

	builderCtx.BuiltInLibrariesDirs = configuration.IDEBuiltinLibrariesDir(configuration.Settings)

	// Will be deprecated.
	builderCtx.ArduinoAPIVersion = "10607"

	builderCtx.Stdout = outStream
	builderCtx.Stderr = errStream
	builderCtx.Clean = req.GetClean()
	builderCtx.OnlyUpdateCompilationDatabase = req.GetCreateCompilationDatabaseOnly()
	builderCtx.SourceOverride = req.GetSourceOverride()

	r = &rpc.CompileResponse{}
	defer func() {
		if p := builderCtx.BuildPath; p != nil {
			r.BuildPath = p.String()
		}
		if pl := builderCtx.TargetPlatform; pl != nil {
			r.BoardPlatform = pl.ToRPCPlatformReference()
		}
		if pl := builderCtx.ActualPlatform; pl != nil {
			r.BuildPlatform = pl.ToRPCPlatformReference()
		}
	}()

	// if --preprocess or --show-properties were passed, we can stop here
	if req.GetShowProperties() {
		compileErr := builder.RunParseHardwareAndDumpBuildProperties(builderCtx)
		if compileErr != nil {
			compileErr = &arduino.CompileFailedError{Message: err.Error()}
		}
		return r, compileErr
	} else if req.GetPreprocess() {
		compileErr := builder.RunPreprocess(builderCtx)
		if compileErr != nil {
			compileErr = &arduino.CompileFailedError{Message: err.Error()}
		}
		return r, compileErr
	}

	defer func() {
		importedLibs := []*rpc.Library{}
		for _, lib := range builderCtx.ImportedLibraries {
			rpcLib, err := lib.ToRPCLibrary()
			if err != nil {
				msg := tr("Error getting information for library %s", lib.Name) + ": " + err.Error() + "\n"
				errStream.Write([]byte(msg))
			}
			importedLibs = append(importedLibs, rpcLib)
		}
		r.UsedLibraries = importedLibs
	}()

	// if it's a regular build, go on...
	if err := builder.RunBuilder(builderCtx); err != nil {
		return r, &arduino.CompileFailedError{Message: err.Error()}
	}

	// If the export directory is set we assume you want to export the binaries
	if req.GetExportDir() != "" {
		exportBinaries = true
	}
	// If CreateCompilationDatabaseOnly is set, we do not need to export anything
	if req.GetCreateCompilationDatabaseOnly() {
		exportBinaries = false
	}
	if exportBinaries {
		var exportPath *paths.Path
		if exportDir := req.GetExportDir(); exportDir != "" {
			exportPath = paths.New(exportDir)
		} else {
			// Add FQBN (without configs part) to export path
			fqbnSuffix := strings.Replace(fqbn.StringWithoutConfig(), ":", ".", -1)
			exportPath = sk.FullPath.Join("build", fqbnSuffix)
		}
		logrus.WithField("path", exportPath).Trace("Saving sketch to export path.")
		if err := exportPath.MkdirAll(); err != nil {
			return r, &arduino.PermissionDeniedError{Message: tr("Error creating output dir"), Cause: err}
		}

		// Copy all "sketch.ino.*" artifacts to the export directory
		baseName, ok := builderCtx.BuildProperties.GetOk("build.project_name") // == "sketch.ino"
		if !ok {
			return r, &arduino.MissingPlatformPropertyError{Property: "build.project_name"}
		}
		buildFiles, err := builderCtx.BuildPath.ReadDir()
		if err != nil {
			return r, &arduino.PermissionDeniedError{Message: tr("Error reading build directory"), Cause: err}
		}
		buildFiles.FilterPrefix(baseName)
		for _, buildFile := range buildFiles {
			exportedFile := exportPath.Join(buildFile.Base())
			logrus.
				WithField("src", buildFile).
				WithField("dest", exportedFile).
				Trace("Copying artifact.")
			if err = buildFile.CopyTo(exportedFile); err != nil {
				return r, &arduino.PermissionDeniedError{Message: tr("Error copying output file %s", buildFile), Cause: err}
			}
		}
	}

	r.ExecutableSectionsSize = builderCtx.ExecutableSectionsSize.ToRPCExecutableSectionSizeArray()

	logrus.Tracef("Compile %s for %s successful", sk.Name, fqbnIn)

	return r, nil
}
