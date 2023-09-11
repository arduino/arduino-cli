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
	"github.com/arduino/arduino-cli/arduino/builder/compilation"
	"github.com/arduino/arduino-cli/arduino/builder/detector"
	"github.com/arduino/arduino-cli/arduino/builder/logger"
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/arduino-cli/arduino/utils"
	"github.com/arduino/arduino-cli/buildcache"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/configuration"
	"github.com/arduino/arduino-cli/i18n"
	"github.com/arduino/arduino-cli/internal/inventory"
	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	paths "github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
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
	sk, err := sketch.New(sketchPath)
	if err != nil {
		return nil, &arduino.CantOpenSketchError{Cause: err}
	}

	fqbnIn := req.GetFqbn()
	if fqbnIn == "" && sk != nil {
		if pme.GetProfile() != nil {
			fqbnIn = pme.GetProfile().FQBN
		} else {
			fqbnIn = sk.GetDefaultFQBN()
		}
	}
	if fqbnIn == "" {
		return nil, &arduino.MissingFQBNError{}
	}

	fqbn, err := cores.ParseFQBN(fqbnIn)
	if err != nil {
		return nil, &arduino.InvalidFQBNError{Cause: err}
	}
	targetPackage, targetPlatform, targetBoard, boardBuildProperties, buildPlatform, err := pme.ResolveFQBN(fqbn)
	if err != nil {
		if targetPlatform == nil {
			return nil, &arduino.PlatformNotFoundError{
				Platform: fmt.Sprintf("%s:%s", fqbn.Package, fqbn.PlatformArch),
				Cause:    fmt.Errorf(tr("platform not installed")),
			}
		}
		return nil, &arduino.InvalidFQBNError{Cause: err}
	}

	r = &rpc.CompileResponse{}
	r.BoardPlatform = targetPlatform.ToRPCPlatformReference()
	r.BuildPlatform = buildPlatform.ToRPCPlatformReference()

	// Setup sign keys if requested
	if req.KeysKeychain != "" {
		boardBuildProperties.Set("build.keys.keychain", req.GetKeysKeychain())
	}
	if req.SignKey != "" {
		boardBuildProperties.Set("build.keys.sign_key", req.GetSignKey())
	}
	if req.EncryptKey != "" {
		boardBuildProperties.Set("build.keys.encrypt_key", req.GetEncryptKey())
	}
	// At the current time we do not have a way of knowing if a board supports the secure boot or not,
	// so, if the flags to override the default keys are used, we try override the corresponding platform property nonetheless.
	// It's not possible to use the default name for the keys since there could be more tools to sign and encrypt.
	// So it's mandatory to use all three flags to sign and encrypt the binary
	keychainProp := boardBuildProperties.ContainsKey("build.keys.keychain")
	signProp := boardBuildProperties.ContainsKey("build.keys.sign_key")
	encryptProp := boardBuildProperties.ContainsKey("build.keys.encrypt_key")
	// we verify that all the properties for the secure boot keys are defined or none of them is defined.
	if !(keychainProp == signProp && signProp == encryptProp) {
		return nil, fmt.Errorf(tr("Firmware encryption/signing requires all the following properties to be defined: %s", "build.keys.keychain, build.keys.sign_key, build.keys.encrypt_key"))
	}

	// Generate or retrieve build path
	var buildPath *paths.Path
	if buildPathArg := req.GetBuildPath(); buildPathArg != "" {
		buildPath = paths.New(req.GetBuildPath()).Canonical()
		if in := buildPath.IsInsideDir(sk.FullPath); in && buildPath.IsDir() {
			if sk.AdditionalFiles, err = removeBuildFromSketchFiles(sk.AdditionalFiles, buildPath); err != nil {
				return nil, err
			}
		}
	}
	if buildPath == nil {
		buildPath = sk.DefaultBuildPath()
	}
	if err = buildPath.MkdirAll(); err != nil {
		return nil, &arduino.PermissionDeniedError{Message: tr("Cannot create build directory"), Cause: err}
	}
	buildcache.New(buildPath.Parent()).GetOrCreate(buildPath.Base())
	// cache is purged after compilation to not remove entries that might be required
	defer maybePurgeBuildCache()

	var coreBuildCachePath *paths.Path
	if req.GetBuildCachePath() == "" {
		coreBuildCachePath = paths.TempDir().Join("arduino", "cores")
	} else {
		buildCachePath, err := paths.New(req.GetBuildCachePath()).Abs()
		if err != nil {
			return nil, &arduino.PermissionDeniedError{Message: tr("Cannot create build cache directory"), Cause: err}
		}
		if err := buildCachePath.MkdirAll(); err != nil {
			return nil, &arduino.PermissionDeniedError{Message: tr("Cannot create build cache directory"), Cause: err}
		}
		coreBuildCachePath = buildCachePath.Join("core")
	}

	sketchBuilder := bldr.NewBuilder(
		sk,
		boardBuildProperties,
		buildPath,
		req.GetOptimizeForDebug(),
		coreBuildCachePath,
		int(req.GetJobs()),
	)

	buildProperties := sketchBuilder.GetBuildProperties()

	// Add user provided custom build properties
	customBuildPropertiesArgs := append(req.GetBuildProperties(), "build.warn_data_percentage=75")
	if customBuildProperties, err := properties.LoadFromSlice(req.GetBuildProperties()); err == nil {
		buildProperties.Merge(customBuildProperties)
	} else {
		return nil, &arduino.InvalidArgumentError{Message: tr("Invalid build properties"), Cause: err}
	}

	requiredTools, err := pme.FindToolsRequiredForBuild(targetPlatform, buildPlatform)
	if err != nil {
		return nil, err
	}

	builderCtx := &types.Context{}
	builderCtx.Builder = sketchBuilder
	builderCtx.PackageManager = pme
	builderCtx.TargetBoard = targetBoard
	builderCtx.TargetPlatform = targetPlatform
	builderCtx.TargetPackage = targetPackage
	builderCtx.ActualPlatform = buildPlatform
	builderCtx.RequiredTools = requiredTools
	builderCtx.BuildProperties = buildProperties
	builderCtx.CustomBuildProperties = customBuildPropertiesArgs
	builderCtx.FQBN = fqbn
	builderCtx.BuildPath = buildPath
	builderCtx.ProgressCB = progressCB

	// FIXME: This will be redundant when arduino-builder will be part of the cli
	builderCtx.HardwareDirs = configuration.HardwareDirectories(configuration.Settings)
	builderCtx.BuiltInToolsDirs = configuration.BuiltinToolsDirectories(configuration.Settings)
	builderCtx.OtherLibrariesDirs = paths.NewPathList(req.GetLibraries()...)
	builderCtx.OtherLibrariesDirs.Add(configuration.LibrariesDir(configuration.Settings))

	builderCtx.CompilationDatabase = compilation.NewDatabase(
		builderCtx.BuildPath.Join("compile_commands.json"),
	)

	builderCtx.Verbose = req.GetVerbose()

	warningsLevel := req.GetWarnings()
	// TODO move this inside the logger
	if warningsLevel == "" {
		warningsLevel = builder.DEFAULT_WARNINGS_LEVEL
	}
	builderCtx.WarningsLevel = warningsLevel

	builderCtx.BuiltInLibrariesDirs = configuration.IDEBuiltinLibrariesDir(configuration.Settings)

	builderCtx.Stdout = outStream
	builderCtx.Stderr = errStream
	builderCtx.Clean = req.GetClean()
	builderCtx.OnlyUpdateCompilationDatabase = req.GetCreateCompilationDatabaseOnly()
	builderCtx.SourceOverride = req.GetSourceOverride()

	builderLogger := logger.New(outStream, errStream, builderCtx.Verbose, warningsLevel)
	builderCtx.BuilderLogger = builderLogger

	sketchBuildPath, err := buildPath.Join(constants.FOLDER_SKETCH).Abs()
	if err != nil {
		return r, &arduino.CompileFailedError{Message: err.Error()}
	}
	librariesBuildPath, err := buildPath.Join(constants.FOLDER_LIBRARIES).Abs()
	if err != nil {
		return r, &arduino.CompileFailedError{Message: err.Error()}
	}
	coreBuildPath, err := buildPath.Join(constants.FOLDER_CORE).Abs()
	if err != nil {
		return r, &arduino.CompileFailedError{Message: err.Error()}
	}
	builderCtx.SketchBuildPath = sketchBuildPath
	builderCtx.LibrariesBuildPath = librariesBuildPath
	builderCtx.CoreBuildPath = coreBuildPath

	if builderCtx.BuildPath.Canonical().EqualsTo(sk.FullPath.Canonical()) {
		return r, &arduino.CompileFailedError{
			Message: tr("Sketch cannot be located in build path. Please specify a different build path"),
		}
	}

	var libsManager *librariesmanager.LibrariesManager
	if pme.GetProfile() != nil {
		libsManager = lm
	}
	useCachedLibrariesResolution := req.GetSkipLibrariesDiscovery()
	libraryDir := paths.NewPathList(req.Library...)
	libsManager, libsResolver, verboseOut, err := detector.LibrariesLoader(
		useCachedLibrariesResolution, libsManager,
		builderCtx.BuiltInLibrariesDirs, libraryDir, builderCtx.OtherLibrariesDirs,
		builderCtx.ActualPlatform, builderCtx.TargetPlatform,
	)
	if err != nil {
		return r, &arduino.CompileFailedError{Message: err.Error()}
	}

	if builderCtx.Verbose {
		builderLogger.Warn(string(verboseOut))
	}

	builderCtx.SketchLibrariesDetector = detector.NewSketchLibrariesDetector(
		libsManager, libsResolver,
		useCachedLibrariesResolution,
		req.GetCreateCompilationDatabaseOnly(),
		builderLogger,
	)

	defer func() {
		if p := builderCtx.BuildPath; p != nil {
			r.BuildPath = p.String()
		}
	}()

	defer func() {
		buildProperties := builderCtx.BuildProperties
		if buildProperties == nil {
			return
		}
		keys := buildProperties.Keys()
		sort.Strings(keys)
		for _, key := range keys {
			r.BuildProperties = append(r.BuildProperties, key+"="+buildProperties.Get(key))
		}
		if !req.GetDoNotExpandBuildProperties() {
			r.BuildProperties, _ = utils.ExpandBuildProperties(r.BuildProperties)
		}
	}()

	// Just get build properties and exit
	if req.GetShowProperties() {
		return r, nil
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
		importedLibs := []*rpc.Library{}
		for _, lib := range builderCtx.SketchLibrariesDetector.ImportedLibraries() {
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

	if req.GetVerbose() {
		core := buildProperties.Get("build.core")
		if core == "" {
			core = "arduino"
		}
		// select the core name in case of "package:core" format
		normalizedFQBN, err := pme.NormalizeFQBN(fqbn)
		if err != nil {
			outStream.Write([]byte(fmt.Sprintf("Could not normalize FQBN: %s\n", err)))
			normalizedFQBN = fqbn
		}
		outStream.Write([]byte(fmt.Sprintf("FQBN: %s\n", normalizedFQBN)))
		core = core[strings.Index(core, ":")+1:]
		outStream.Write([]byte(tr("Using board '%[1]s' from platform in folder: %[2]s", targetBoard.BoardID, targetPlatform.InstallDir) + "\n"))
		outStream.Write([]byte(tr("Using core '%[1]s' from platform in folder: %[2]s", core, buildPlatform.InstallDir) + "\n"))
		outStream.Write([]byte("\n"))
	}
	if !targetBoard.Properties.ContainsKey("build.board") {
		outStream.Write([]byte(
			tr("Warning: Board %[1]s doesn't define a %[2]s preference. Auto-set to: %[3]s",
				targetBoard.String(), "'build.board'", buildProperties.Get("build.board")) + "\n"))
	}

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
		err := builder.RecipeByPrefixSuffixRunner(
			"recipe.hooks.savehex.presavehex", ".pattern", false,
			builderCtx.OnlyUpdateCompilationDatabase,
			builderCtx.BuildProperties,
			builderLogger,
		)
		if err != nil {
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

		err = builder.RecipeByPrefixSuffixRunner(
			"recipe.hooks.savehex.postsavehex", ".pattern", false,
			builderCtx.OnlyUpdateCompilationDatabase,
			builderCtx.BuildProperties, builderLogger,
		)
		if err != nil {
			return r, err
		}
	}

	r.ExecutableSectionsSize = builderCtx.ExecutableSectionsSize.ToRPCExecutableSectionSizeArray()

	logrus.Tracef("Compile %s for %s successful", sk.Name, fqbnIn)

	return r, nil
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
		if !file.IsInsideDir(build) {
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
