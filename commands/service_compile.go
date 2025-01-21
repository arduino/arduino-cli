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

package commands

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/commands/internal/instances"
	"github.com/arduino/arduino-cli/internal/arduino/builder"
	"github.com/arduino/arduino-cli/internal/arduino/builder/logger"
	"github.com/arduino/arduino-cli/internal/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/internal/arduino/sketch"
	"github.com/arduino/arduino-cli/internal/arduino/utils"
	"github.com/arduino/arduino-cli/internal/buildcache"
	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/arduino/arduino-cli/internal/inventory"
	"github.com/arduino/arduino-cli/pkg/fqbn"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	paths "github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
)

// CompilerServerToStreams creates a gRPC CompileServer that sends the responses to the provided streams.
// The returned callback function can be used to retrieve the builder result after the compilation is done.
func CompilerServerToStreams(ctx context.Context, stdOut, stderr io.Writer, progressCB rpc.TaskProgressCB) (server rpc.ArduinoCoreService_CompileServer, resultCB func() *rpc.BuilderResult) {
	var builderResult *rpc.BuilderResult
	stream := streamResponseToCallback(ctx, func(resp *rpc.CompileResponse) error {
		if out := resp.GetOutStream(); len(out) > 0 {
			if _, err := stdOut.Write(out); err != nil {
				return err
			}
		}
		if err := resp.GetErrStream(); len(err) > 0 {
			if _, err := stderr.Write(err); err != nil {
				return err
			}
		}
		if result := resp.GetResult(); result != nil {
			builderResult = result
		}
		if progress := resp.GetProgress(); progress != nil {
			if progressCB != nil {
				progressCB(progress)
			}
		}
		return nil
	})
	return stream, func() *rpc.BuilderResult { return builderResult }
}

// Compile performs a compilation of a sketch.
func (s *arduinoCoreServerImpl) Compile(req *rpc.CompileRequest, stream rpc.ArduinoCoreService_CompileServer) error {
	ctx := stream.Context()
	syncSend := NewSynchronizedSend(stream.Send)

	exportBinaries := s.settings.SketchAlwaysExportBinaries()
	if e := req.ExportBinaries; e != nil {
		exportBinaries = *e
	}

	pme, release, err := instances.GetPackageManagerExplorer(req.GetInstance())
	if err != nil {
		return err
	}
	release = sync.OnceFunc(release)
	defer release()

	if pme.Dirty() {
		return &cmderrors.InstanceNeedsReinitialization{}
	}

	lm, err := instances.GetLibraryManager(req.GetInstance())
	if err != nil {
		return err
	}

	logrus.Tracef("Compile %s for %s started", req.GetSketchPath(), req.GetFqbn())
	if req.GetSketchPath() == "" {
		return &cmderrors.MissingSketchPathError{}
	}
	sketchPath := paths.New(req.GetSketchPath())
	sk, err := sketch.New(sketchPath)
	if err != nil {
		return &cmderrors.CantOpenSketchError{Cause: err}
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
		return &cmderrors.MissingFQBNError{}
	}

	fqbn, err := fqbn.Parse(fqbnIn)
	if err != nil {
		return &cmderrors.InvalidFQBNError{Cause: err}
	}
	_, targetPlatform, targetBoard, boardBuildProperties, buildPlatform, err := pme.ResolveFQBN(fqbn)
	if err != nil {
		if targetPlatform == nil {
			return &cmderrors.PlatformNotFoundError{
				Platform: fmt.Sprintf("%s:%s", fqbn.Packager, fqbn.Architecture),
				Cause:    errors.New(i18n.Tr("platform not installed")),
			}
		}
		return &cmderrors.InvalidFQBNError{Cause: err}
	}

	r := &rpc.BuilderResult{}
	defer func() {
		syncSend.Send(&rpc.CompileResponse{
			Message: &rpc.CompileResponse_Result{Result: r},
		})
	}()
	r.BoardPlatform = targetPlatform.ToRPCPlatformReference()
	r.BuildPlatform = buildPlatform.ToRPCPlatformReference()

	// Setup sign keys if requested
	if req.GetKeysKeychain() != "" {
		boardBuildProperties.Set("build.keys.keychain", req.GetKeysKeychain())
	}
	if req.GetSignKey() != "" {
		boardBuildProperties.Set("build.keys.sign_key", req.GetSignKey())
	}
	if req.GetEncryptKey() != "" {
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
		return errors.New(i18n.Tr("Firmware encryption/signing requires all the following properties to be defined: %s", "build.keys.keychain, build.keys.sign_key, build.keys.encrypt_key"))
	}

	// Retrieve build path from user arguments
	var buildPath *paths.Path
	if buildPathArg := req.GetBuildPath(); buildPathArg != "" {
		buildPath = paths.New(req.GetBuildPath()).Canonical()
		if in, _ := buildPath.IsInsideDir(sk.FullPath); in && buildPath.IsDir() {
			if sk.AdditionalFiles, err = removeBuildFromSketchFiles(sk.AdditionalFiles, buildPath); err != nil {
				return err
			}
		}
	}

	// If no build path has been set by the user:
	// - set up the build cache directory
	// - set the sketch build path inside the build cache directory.
	var coreBuildCachePath *paths.Path
	var extraCoreBuildCachePaths paths.PathList
	if buildPath == nil {
		var buildCachePath *paths.Path
		if p := req.GetBuildCachePath(); p != "" { //nolint:staticcheck
			buildCachePath = paths.New(p)
		} else {
			buildCachePath = s.settings.GetBuildCachePath()
		}
		if err := buildCachePath.ToAbs(); err != nil {
			return &cmderrors.PermissionDeniedError{Message: i18n.Tr("Cannot create build cache directory"), Cause: err}
		}
		if err := buildCachePath.MkdirAll(); err != nil {
			return &cmderrors.PermissionDeniedError{Message: i18n.Tr("Cannot create build cache directory"), Cause: err}
		}
		coreBuildCachePath = buildCachePath.Join("cores")

		if len(req.GetBuildCacheExtraPaths()) == 0 {
			extraCoreBuildCachePaths = s.settings.GetBuildCacheExtraPaths()
		} else {
			extraCoreBuildCachePaths = paths.NewPathList(req.GetBuildCacheExtraPaths()...)
		}
		for i, p := range extraCoreBuildCachePaths {
			extraCoreBuildCachePaths[i] = p.Join("cores")
		}

		buildPath = s.getDefaultSketchBuildPath(sk, buildCachePath)
		buildcache.New(buildPath.Parent()).GetOrCreate(buildPath.Base())

		// cache is purged after compilation to not remove entries that might be required
		defer maybePurgeBuildCache(
			s.settings.GetCompilationsBeforeBuildCachePurge(),
			s.settings.GetBuildCacheTTL().Abs())
	}
	if err = buildPath.MkdirAll(); err != nil {
		return &cmderrors.PermissionDeniedError{Message: i18n.Tr("Cannot create build directory"), Cause: err}
	}

	if _, err := pme.FindToolsRequiredForBuild(targetPlatform, buildPlatform); err != nil {
		return err
	}

	actualPlatform := buildPlatform
	otherLibrariesDirs := paths.NewPathList(req.GetLibraries()...)
	otherLibrariesDirs.Add(s.settings.LibrariesDir())

	var libsManager *librariesmanager.LibrariesManager
	if pme.GetProfile() != nil {
		libsManager = lm
	}

	outStream := feedStreamTo(func(data []byte) {
		syncSend.Send(&rpc.CompileResponse{
			Message: &rpc.CompileResponse_OutStream{OutStream: data},
		})
	})
	defer outStream.Close()
	errStream := feedStreamTo(func(data []byte) {
		syncSend.Send(&rpc.CompileResponse{
			Message: &rpc.CompileResponse_ErrStream{ErrStream: data},
		})
	})
	defer errStream.Close()
	progressCB := func(p *rpc.TaskProgress) {
		syncSend.Send(&rpc.CompileResponse{
			Message: &rpc.CompileResponse_Progress{Progress: p},
		})
	}
	var verbosity logger.Verbosity = logger.VerbosityNormal
	if req.GetVerbose() {
		verbosity = logger.VerbosityVerbose
	}
	sketchBuilder, err := builder.NewBuilder(
		ctx,
		sk,
		boardBuildProperties,
		buildPath,
		req.GetOptimizeForDebug(),
		coreBuildCachePath,
		extraCoreBuildCachePaths,
		int(req.GetJobs()),
		req.GetBuildProperties(),
		s.settings.HardwareDirectories(),
		otherLibrariesDirs,
		s.settings.IDEBuiltinLibrariesDir(),
		fqbn,
		req.GetClean(),
		req.GetSourceOverride(),
		req.GetCreateCompilationDatabaseOnly(),
		targetPlatform, actualPlatform,
		req.GetSkipLibrariesDiscovery(),
		libsManager,
		paths.NewPathList(req.GetLibrary()...),
		outStream, errStream, verbosity, req.GetWarnings(),
		progressCB,
		pme.GetEnvVarsForSpawnedProcess(),
	)
	if err != nil {
		if strings.Contains(err.Error(), "invalid build properties") {
			return &cmderrors.InvalidArgumentError{Message: i18n.Tr("Invalid build properties"), Cause: err}
		}
		if errors.Is(err, builder.ErrSketchCannotBeLocatedInBuildPath) {
			return &cmderrors.CompileFailedError{
				Message: i18n.Tr("Sketch cannot be located in build path. Please specify a different build path"),
			}
		}
		return &cmderrors.CompileFailedError{Message: err.Error()}
	}

	defer func() {
		if p := sketchBuilder.GetBuildPath(); p != nil {
			r.BuildPath = p.String()
		}
	}()

	defer func() {
		r.Diagnostics = sketchBuilder.CompilerDiagnostics().ToRPC()
	}()

	defer func() {
		buildProperties := sketchBuilder.GetBuildProperties()
		if buildProperties == nil {
			return
		}
		keys := buildProperties.Keys()
		sort.Strings(keys)
		for _, key := range keys {
			r.BuildProperties = append(r.GetBuildProperties(), key+"="+buildProperties.Get(key))
		}
		if !req.GetDoNotExpandBuildProperties() {
			r.BuildProperties, _ = utils.ExpandBuildProperties(r.GetBuildProperties())
		}
	}()

	// Just get build properties and exit
	if req.GetShowProperties() {
		return nil
	}

	if req.GetPreprocess() {
		// Just output preprocessed source code and exit
		preprocessedSketch, err := sketchBuilder.Preprocess()
		if err != nil {
			err = &cmderrors.CompileFailedError{Message: err.Error()}
			return err
		}
		_, err = outStream.Write(preprocessedSketch)
		return err
	}

	defer func() {
		importedLibs := []*rpc.Library{}
		for _, lib := range sketchBuilder.ImportedLibraries() {
			rpcLib, err := lib.ToRPCLibrary()
			if err != nil {
				msg := i18n.Tr("Error getting information for library %s", lib.Name) + ": " + err.Error() + "\n"
				errStream.Write([]byte(msg))
			}
			importedLibs = append(importedLibs, rpcLib)
		}
		r.UsedLibraries = importedLibs
	}()

	// if it's a regular build, go on...

	if req.GetVerbose() {
		core := sketchBuilder.GetBuildProperties().Get("build.core")
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
		outStream.Write([]byte(i18n.Tr("Using board '%[1]s' from platform in folder: %[2]s", targetBoard.BoardID, targetPlatform.InstallDir) + "\n"))
		outStream.Write([]byte(i18n.Tr("Using core '%[1]s' from platform in folder: %[2]s", core, buildPlatform.InstallDir) + "\n"))
		outStream.Write([]byte("\n"))
	}
	if !targetBoard.Properties.ContainsKey("build.board") {
		outStream.Write([]byte(
			i18n.Tr("Warning: Board %[1]s doesn't define a %[2]s preference. Auto-set to: %[3]s",
				targetBoard.String(), "'build.board'", sketchBuilder.GetBuildProperties().Get("build.board")) + "\n"))
	}

	// Release package manager
	release()

	// Perform the actual build
	if err := sketchBuilder.Build(); err != nil {
		return &cmderrors.CompileFailedError{Message: err.Error()}
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
		if err := sketchBuilder.RunRecipe("recipe.hooks.savehex.presavehex", ".pattern", false); err != nil {
			return err
		}

		exportPath := paths.New(req.GetExportDir())
		if exportPath == nil {
			// Add FQBN (without configs part) to export path
			fqbnSuffix := strings.ReplaceAll(fqbn.StringWithoutConfig(), ":", ".")
			exportPath = sk.FullPath.Join("build", fqbnSuffix)
		}

		// Copy all "sketch.ino.*" artifacts to the export directory
		if !buildPath.EqualsTo(exportPath) {
			logrus.WithField("path", exportPath).Trace("Saving sketch to export path.")
			if err := exportPath.MkdirAll(); err != nil {
				return &cmderrors.PermissionDeniedError{Message: i18n.Tr("Error creating output dir"), Cause: err}
			}

			baseName, ok := sketchBuilder.GetBuildProperties().GetOk("build.project_name") // == "sketch.ino"
			if !ok {
				return &cmderrors.MissingPlatformPropertyError{Property: "build.project_name"}
			}
			buildFiles, err := sketchBuilder.GetBuildPath().ReadDir()
			if err != nil {
				return &cmderrors.PermissionDeniedError{Message: i18n.Tr("Error reading build directory"), Cause: err}
			}
			buildFiles.FilterPrefix(baseName)
			buildFiles.FilterOutDirs()
			for _, buildFile := range buildFiles {
				exportedFile := exportPath.Join(buildFile.Base())
				logrus.WithField("src", buildFile).WithField("dest", exportedFile).Trace("Copying artifact.")
				if err = buildFile.CopyTo(exportedFile); err != nil {
					return &cmderrors.PermissionDeniedError{Message: i18n.Tr("Error copying output file %s", buildFile), Cause: err}
				}
			}
		}

		if err = sketchBuilder.RunRecipe("recipe.hooks.savehex.postsavehex", ".pattern", false); err != nil {
			return err
		}
	}

	r.ExecutableSectionsSize = sketchBuilder.ExecutableSectionsSize().ToRPCExecutableSectionSizeArray()

	logrus.Tracef("Compile %s for %s successful", sk.Name, fqbnIn)
	return nil
}

// getDefaultSketchBuildPath generates the default build directory for a given sketch.
// The sketch build path is inside the build cache path and is unique for each sketch.
// If overriddenBuildCachePath is nil the build cache path is taken from the settings.
func (s *arduinoCoreServerImpl) getDefaultSketchBuildPath(sk *sketch.Sketch, overriddenBuildCachePath *paths.Path) *paths.Path {
	if overriddenBuildCachePath == nil {
		overriddenBuildCachePath = s.settings.GetBuildCachePath()
	}
	return overriddenBuildCachePath.Join("sketches", sk.Hash())
}

// maybePurgeBuildCache runs the build files cache purge if the policy conditions are met.
func maybePurgeBuildCache(compilationsBeforePurge uint, cacheTTL time.Duration) {
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
	buildcache.New(paths.TempDir().Join("arduino", "cores")).Purge(cacheTTL)
	buildcache.New(paths.TempDir().Join("arduino", "sketches")).Purge(cacheTTL)
}

// removeBuildFromSketchFiles removes the files contained in the build directory from
// the list of the sketch files
func removeBuildFromSketchFiles(files paths.PathList, build *paths.Path) (paths.PathList, error) {
	var res paths.PathList
	ignored := false
	for _, file := range files {
		if isInside, _ := file.IsInsideDir(build); !isInside {
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
