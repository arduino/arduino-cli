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

package test

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	bldr "github.com/arduino/arduino-cli/arduino/builder"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/phases"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func cleanUpBuilderTestContext(t *testing.T, ctx *types.Context) {
	if ctx.BuildPath != nil {
		err := ctx.BuildPath.RemoveAll()
		require.NoError(t, err)
	}
}

func prepareBuilderTestContext(t *testing.T, ctx *types.Context, sketchPath *paths.Path, fqbn string) *types.Context {
	DownloadCoresAndToolsAndLibraries(t)

	if ctx == nil {
		ctx = &types.Context{}
	}
	if ctx.HardwareDirs.Len() == 0 {
		ctx.HardwareDirs = paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware")
		ctx.BuiltInToolsDirs = paths.NewPathList("downloaded_tools")
		ctx.BuiltInLibrariesDirs = paths.New("downloaded_libraries")
		ctx.OtherLibrariesDirs = paths.NewPathList("libraries")
	}
	if ctx.BuildPath == nil {
		buildPath, err := paths.MkTempDir("", "test_build_path")
		NoError(t, err)
		ctx.BuildPath = buildPath
	}

	// Create a Package Manager from the given context
	// This should happen only on legacy arduino-builder.
	// Hopefully this piece will be removed once the legacy package will be cleanedup.
	pmb := packagemanager.NewBuilder(nil, nil, nil, nil, "arduino-builder")
	for _, err := range pmb.LoadHardwareFromDirectories(ctx.HardwareDirs) {
		// NoError(t, err)
		fmt.Println(err)
	}
	if !ctx.CanUseCachedTools {
		if ctx.BuiltInToolsDirs != nil {
			pmb.LoadToolsFromBundleDirectories(ctx.BuiltInToolsDirs)
		}
		ctx.CanUseCachedTools = true
	}
	pm := pmb.Build()
	pme, _ /* never release... */ := pm.NewExplorer()
	ctx.PackageManager = pme

	if sketchPath != nil {
		sk, err := sketch.New(sketchPath)
		require.NoError(t, err)
		ctx.Sketch = sk
	}

	if fqbn != "" {
		ctx.FQBN = parseFQBN(t, fqbn)
		targetPackage, targetPlatform, targetBoard, buildProperties, buildPlatform, err := pme.ResolveFQBN(ctx.FQBN)
		require.NoError(t, err)
		requiredTools, err := pme.FindToolsRequiredForBuild(targetPlatform, buildPlatform)
		require.NoError(t, err)

		buildProperties = bldr.SetupBuildProperties(buildProperties, ctx.BuildPath, ctx.Sketch, false /*OptimizeForDebug*/)
		ctx.PackageManager = pme
		ctx.TargetBoard = targetBoard
		ctx.BuildProperties = buildProperties
		ctx.TargetPlatform = targetPlatform
		ctx.TargetPackage = targetPackage
		ctx.ActualPlatform = buildPlatform
		ctx.RequiredTools = requiredTools
	}

	return ctx
}

func TestBuilderEmptySketch(t *testing.T) {
	ctx := prepareBuilderTestContext(t, nil, paths.New("sketch1", "sketch1.ino"), "arduino:avr:uno")
	defer cleanUpBuilderTestContext(t, ctx)

	// Run builder
	command := builder.Builder{}
	err := command.Run(ctx)
	NoError(t, err)

	buildPath := ctx.BuildPath
	exist, err := buildPath.Join(constants.FOLDER_CORE, "HardwareSerial.cpp.o").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join(constants.FOLDER_SKETCH, "sketch1.ino.cpp.o").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join("sketch1.ino.elf").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
	exist, err = buildPath.Join("sketch1.ino.hex").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
}

func TestBuilderSketchBuildPathContainsUnusedPreviouslyCompiledLibrary(t *testing.T) {
	ctx := prepareBuilderTestContext(t, nil, paths.New("downloaded_libraries", "Bridge", "examples", "Bridge", "Bridge.ino"), "arduino:avr:leonardo")
	defer cleanUpBuilderTestContext(t, ctx)

	buildPath := ctx.BuildPath
	NoError(t, buildPath.Join("libraries", "SPI").MkdirAll())

	// Run builder
	command := builder.Builder{}
	err := command.Run(ctx)
	NoError(t, err)

	exist, err := buildPath.Join("libraries", "SPI").ExistCheck()
	NoError(t, err)
	require.False(t, exist)
	exist, err = buildPath.Join("libraries", "Bridge").ExistCheck()
	NoError(t, err)
	require.True(t, exist)
}

func TestBuilderWithBuildPathInSketchDir(t *testing.T) {
	buildPath, err := paths.New("sketch1", "build").Abs()
	NoError(t, err)
	ctx := prepareBuilderTestContext(t, &types.Context{BuildPath: buildPath}, paths.New("sketch1", "sketch1.ino"), "arduino:avr:uno")
	defer cleanUpBuilderTestContext(t, ctx)

	// Run build
	command := builder.Builder{}
	err = command.Run(ctx)
	NoError(t, err)

	// Run build twice, to verify the build still works when the
	// build directory is present at the start
	err = command.Run(ctx)
	NoError(t, err)
}

func TestBuilderCacheCoreAFile(t *testing.T) {
	ctx := prepareBuilderTestContext(t, nil, paths.New("sketch1", "sketch1.ino"), "arduino:avr:uno")
	defer cleanUpBuilderTestContext(t, ctx)

	SetupBuildCachePath(t, ctx)
	defer ctx.CoreBuildCachePath.RemoveAll()

	// Run build
	bldr := builder.Builder{}
	err := bldr.Run(ctx)
	NoError(t, err)

	// Pick timestamp of cached core
	coreFolder := paths.New("downloaded_hardware", "arduino", "avr")
	coreFileName := phases.GetCachedCoreArchiveDirName(ctx.FQBN.String(), ctx.BuildProperties.Get("compiler.optimization_flags"), coreFolder)
	cachedCoreFile := ctx.CoreBuildCachePath.Join(coreFileName, "core.a")
	coreStatBefore, err := cachedCoreFile.Stat()
	require.NoError(t, err)
	lastUsedFile := ctx.CoreBuildCachePath.Join(coreFileName, ".last-used")
	_, err = lastUsedFile.Stat()
	require.NoError(t, err)

	// Run build again, to verify that the builder skips rebuilding core.a
	err = bldr.Run(ctx)
	NoError(t, err)

	coreStatAfterRebuild, err := cachedCoreFile.Stat()
	require.NoError(t, err)
	require.Equal(t, coreStatBefore.ModTime(), coreStatAfterRebuild.ModTime())

	// Touch a file of the core and check if the builder invalidate the cache
	time.Sleep(time.Second)
	now := time.Now().Local()
	err = coreFolder.Join("cores", "arduino", "Arduino.h").Chtimes(now, now)
	require.NoError(t, err)

	// Run build again, to verify that the builder rebuilds core.a
	err = bldr.Run(ctx)
	NoError(t, err)

	coreStatAfterTouch, err := cachedCoreFile.Stat()
	require.NoError(t, err)
	require.NotEqual(t, coreStatBefore.ModTime(), coreStatAfterTouch.ModTime())
}
