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

	bldr "github.com/arduino/arduino-cli/arduino/builder"
	"github.com/arduino/arduino-cli/arduino/builder/detector"
	"github.com/arduino/arduino-cli/arduino/builder/logger"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
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

type skipContextPreparationStepName string

const skipLibraries = skipContextPreparationStepName("libraries")

func prepareBuilderTestContext(t *testing.T, ctx *types.Context, sketchPath *paths.Path, fqbn string, skips ...skipContextPreparationStepName) *types.Context {
	DownloadCoresAndToolsAndLibraries(t)

	stepToSkip := map[skipContextPreparationStepName]bool{}
	for _, skip := range skips {
		stepToSkip[skip] = true
	}

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
		require.NoError(t, err)
		ctx.BuildPath = buildPath
	}

	buildPath := ctx.BuildPath
	sketchBuildPath, err := buildPath.Join(constants.FOLDER_SKETCH).Abs()
	require.NoError(t, err)
	librariesBuildPath, err := buildPath.Join(constants.FOLDER_LIBRARIES).Abs()
	require.NoError(t, err)
	coreBuildPath, err := buildPath.Join(constants.FOLDER_CORE).Abs()
	require.NoError(t, err)

	ctx.SketchBuildPath = sketchBuildPath
	ctx.LibrariesBuildPath = librariesBuildPath
	ctx.CoreBuildPath = coreBuildPath

	// Create a Package Manager from the given context
	// This should happen only on legacy arduino-builder.
	// Hopefully this piece will be removed once the legacy package will be cleanedup.
	pmb := packagemanager.NewBuilder(nil, nil, nil, nil, "arduino-builder")
	for _, err := range pmb.LoadHardwareFromDirectories(ctx.HardwareDirs) {
		// NoError(t, err)
		fmt.Println(err)
	}
	if ctx.BuiltInToolsDirs != nil {
		pmb.LoadToolsFromBundleDirectories(ctx.BuiltInToolsDirs)
	}
	pm := pmb.Build()
	pme, _ /* never release... */ := pm.NewExplorer()
	ctx.PackageManager = pme

	var sk *sketch.Sketch
	if sketchPath != nil {
		s, err := sketch.New(sketchPath)
		require.NoError(t, err)
		sk = s
	}

	builderLogger := logger.New(nil, nil, false, "")
	ctx.BuilderLogger = builderLogger
	ctx.Builder, err = bldr.NewBuilder(sk, nil, nil, false, nil, 0, nil)
	require.NoError(t, err)
	if fqbn != "" {
		ctx.FQBN = parseFQBN(t, fqbn)
		targetPackage, targetPlatform, targetBoard, boardBuildProperties, buildPlatform, err := pme.ResolveFQBN(ctx.FQBN)
		require.NoError(t, err)
		requiredTools, err := pme.FindToolsRequiredForBuild(targetPlatform, buildPlatform)
		require.NoError(t, err)

		ctx.Builder, err = bldr.NewBuilder(sk, boardBuildProperties, ctx.BuildPath, false /*OptimizeForDebug*/, nil, 0, nil)
		require.NoError(t, err)

		ctx.PackageManager = pme
		ctx.TargetBoard = targetBoard
		ctx.BuildProperties = ctx.Builder.GetBuildProperties()
		ctx.TargetPlatform = targetPlatform
		ctx.TargetPackage = targetPackage
		ctx.ActualPlatform = buildPlatform
		ctx.RequiredTools = requiredTools
	}

	if sk != nil {
		require.False(t, ctx.BuildPath.Canonical().EqualsTo(sk.FullPath.Canonical()))
	}

	if !stepToSkip[skipLibraries] {
		lm, libsResolver, _, err := detector.LibrariesLoader(
			false, nil,
			ctx.BuiltInLibrariesDirs, nil, ctx.OtherLibrariesDirs,
			ctx.ActualPlatform, ctx.TargetPlatform,
		)
		require.NoError(t, err)

		ctx.SketchLibrariesDetector = detector.NewSketchLibrariesDetector(
			lm, libsResolver,
			false,
			false,
			builderLogger,
		)
	}

	return ctx
}
