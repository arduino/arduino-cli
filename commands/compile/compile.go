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
	"sort"
	"strings"

	"github.com/arduino/arduino-cli/arduino"
	bldr "github.com/arduino/arduino-cli/arduino/builder"
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/arduino-cli/buildcache"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/configuration"
	"github.com/arduino/arduino-cli/i18n"
	"github.com/arduino/arduino-cli/inventory"
	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	paths "github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
)

var tr = i18n.Tr

// Compile FIXMEDOC
func Compile(ctx context.Context, req *rpc.CompileRequest, outStream, errStream io.Writer, progressCB rpc.TaskProgressCB) (r *rpc.CompileResponse, e error) {

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
	builderCtx := &types.Context{}
	builderCtx.PackageManager = pme
	if pme.GetProfile() != nil {
		builderCtx.LibrariesManager = lm
	}

	sk, newSketchErr := sketch.New(sketchPath)

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

	builderCtx.UseCachedLibrariesResolution = req.GetSkipLibrariesDiscovery()
	builderCtx.FQBN = fqbn
	defer func() {
		appendBuildProperties(r, builderCtx)
	}()
	r = &rpc.CompileResponse{}
	if newSketchErr != nil {
		if req.GetShowProperties() {
			// Just get build properties and exit
			compileErr := builder.RunParseHardware(builderCtx)
			if compileErr != nil {
				compileErr = &arduino.CompileFailedError{Message: compileErr.Error()}
			}
			return r, compileErr
		}
		return nil, &arduino.CantOpenSketchError{Cause: err}
	}
	builderCtx.Sketch = sk
	builderCtx.ProgressCB = progressCB

	// FIXME: This will be redundant when arduino-builder will be part of the cli
	builderCtx.HardwareDirs = configuration.HardwareDirectories(configuration.Settings)
	builderCtx.BuiltInToolsDirs = configuration.BuiltinToolsDirectories(configuration.Settings)

	// FIXME: This will be redundant when arduino-builder will be part of the cli
	builderCtx.OtherLibrariesDirs = paths.NewPathList(req.GetLibraries()...)
	builderCtx.OtherLibrariesDirs.Add(configuration.LibrariesDir(configuration.Settings))
	builderCtx.LibraryDirs = paths.NewPathList(req.Library...)
	if req.GetBuildPath() == "" {
		builderCtx.BuildPath = sk.DefaultBuildPath()
	} else {
		builderCtx.BuildPath = paths.New(req.GetBuildPath()).Canonical()
		if in, err := builderCtx.BuildPath.IsInsideDir(sk.FullPath); err != nil {
			return nil, &arduino.NotFoundError{Message: tr("Cannot find build path"), Cause: err}
		} else if in && builderCtx.BuildPath.IsDir() {
			if sk.AdditionalFiles, err = removeBuildFromSketchFiles(sk.AdditionalFiles, builderCtx.BuildPath); err != nil {
				return nil, err
			}
		}
	}
	if err = builderCtx.BuildPath.MkdirAll(); err != nil {
		return nil, &arduino.PermissionDeniedError{Message: tr("Cannot create build directory"), Cause: err}
	}

	buildcache.New(builderCtx.BuildPath.Parent()).GetOrCreate(builderCtx.BuildPath.Base())
	// cache is purged after compilation to not remove entries that might be required
	defer maybePurgeBuildCache()

	builderCtx.CompilationDatabase = bldr.NewCompilationDatabase(
		builderCtx.BuildPath.Join("compile_commands.json"),
	)

	builderCtx.Verbose = req.GetVerbose()

	// Optimize for debug
	builderCtx.OptimizeForDebug = req.GetOptimizeForDebug()

	builderCtx.Jobs = int(req.GetJobs())

	builderCtx.USBVidPid = req.GetVidPid()
	builderCtx.WarningsLevel = req.GetWarnings()

	builderCtx.CustomBuildProperties = append(req.GetBuildProperties(), "build.warn_data_percentage=75")
	builderCtx.CustomBuildProperties = append(req.GetBuildProperties(), securityKeysOverride...)

	if req.GetBuildCachePath() == "" {
		builderCtx.CoreBuildCachePath = paths.TempDir().Join("arduino", "cores")
	} else {
		buildCachePath, err := paths.New(req.GetBuildCachePath()).Abs()
		if err != nil {
			return nil, &arduino.PermissionDeniedError{Message: tr("Cannot create build cache directory"), Cause: err}
		}
		if err := buildCachePath.MkdirAll(); err != nil {
			return nil, &arduino.PermissionDeniedError{Message: tr("Cannot create build cache directory"), Cause: err}
		}
		builderCtx.CoreBuildCachePath = buildCachePath.Join("core")
	}

	builderCtx.BuiltInLibrariesDirs = configuration.IDEBuiltinLibrariesDir(configuration.Settings)

	builderCtx.Stdout = outStream
	builderCtx.Stderr = errStream
	builderCtx.Clean = req.GetClean()
	builderCtx.OnlyUpdateCompilationDatabase = req.GetCreateCompilationDatabaseOnly()
	builderCtx.SourceOverride = req.GetSourceOverride()

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

	if req.GetShowProperties() {
		// Just get build properties and exit
		compileErr := builder.RunParseHardware(builderCtx)
		if compileErr != nil {
			compileErr = &arduino.CompileFailedError{Message: compileErr.Error()}
		}
		return r, compileErr
	}

	if req.GetPreprocess() {
		// Just output preprocessed source code and exit
		compileErr := builder.RunPreprocess(builderCtx)
		if compileErr != nil {
			compileErr = &arduino.CompileFailedError{Message: compileErr.Error()}
		}
		return r, compileErr
	}

	defer func() {
		appendUserLibraries(r, builderCtx, errStream)
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
		presaveHex := builder.RecipeByPrefixSuffixRunner{Prefix: "recipe.hooks.savehex.presavehex", Suffix: ".pattern"}
		if err := presaveHex.Run(builderCtx); err != nil {
			return r, err
		}

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

		postsaveHex := builder.RecipeByPrefixSuffixRunner{Prefix: "recipe.hooks.savehex.postsavehex", Suffix: ".pattern"}
		if err := postsaveHex.Run(builderCtx); err != nil {
			return r, err
		}
	}

	r.ExecutableSectionsSize = builderCtx.ExecutableSectionsSize.ToRPCExecutableSectionSizeArray()

	logrus.Tracef("Compile %s for %s successful", sk.Name, fqbnIn)

	return r, nil
}

func appendUserLibraries(r *rpc.CompileResponse, builderCtx *types.Context, errStream io.Writer) {
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
}

func appendBuildProperties(r *rpc.CompileResponse, builderCtx *types.Context) bool {
	buildProperties := builderCtx.BuildProperties
	if buildProperties == nil {
		return true
	}
	keys := buildProperties.Keys()
	sort.Strings(keys)
	for _, key := range keys {
		r.BuildProperties = append(r.BuildProperties, key+"="+buildProperties.Get(key))
	}
	return false
}

// maybePurgeBuildCache runs the build files cache purge if the policy conditions are met.
func maybePurgeBuildCache() {

	compilationsBeforePurge := configuration.Settings.GetUint("build_cache.compilations_before_purge")
	// 0 means never purge
	if compilationsBeforePurge == 0 {
		return
	}
	compilationSinceLastPurge := inventory.Store.GetUint("build_cache.compilation_count_since_last_purge")
	compilationSinceLastPurge++
	inventory.Store.Set("build_cache.compilation_count_since_last_purge", compilationSinceLastPurge)
	defer inventory.WriteStore()
	if compilationsBeforePurge == 0 || compilationSinceLastPurge < compilationsBeforePurge {
		return
	}
	inventory.Store.Set("build_cache.compilation_count_since_last_purge", 0)
	cacheTTL := configuration.Settings.GetDuration("build_cache.ttl").Abs()
	buildcache.New(paths.TempDir().Join("arduino", "cores")).Purge(cacheTTL)
	buildcache.New(paths.TempDir().Join("arduino", "sketches")).Purge(cacheTTL)
}

// removeBuildFromSketchFiles removes the files contained in the build directory from
// the list of the sketch files
func removeBuildFromSketchFiles(files paths.PathList, build *paths.Path) (paths.PathList, error) {
	var res paths.PathList
	ignored := false
	for _, file := range files {
		if in, err := file.IsInsideDir(build); err != nil {
			return nil, &arduino.NotFoundError{Message: tr("Cannot find build path"), Cause: err}
		} else if !in {
			res = append(res, file)
		} else if !ignored {
			ignored = true
		}
	}
	// log only if at least a file is ignored
	if ignored {
		logrus.Tracef("Build path %s is a child of sketch path and it is ignored for additional files.", build.String())
	}
	return res, nil
}
