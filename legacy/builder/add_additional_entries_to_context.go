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

package builder

import (
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/pkg/errors"
)

type AddAdditionalEntriesToContext struct{}

func (*AddAdditionalEntriesToContext) Run(ctx *types.Context) error {
	if ctx.BuildPath != nil {
		buildPath := ctx.BuildPath
		preprocPath, err := buildPath.Join(constants.FOLDER_PREPROC).Abs()
		if err != nil {
			return errors.WithStack(err)
		}
		sketchBuildPath, err := buildPath.Join(constants.FOLDER_SKETCH).Abs()
		if err != nil {
			return errors.WithStack(err)
		}
		librariesBuildPath, err := buildPath.Join("libraries").Abs()
		if err != nil {
			return errors.WithStack(err)
		}
		coreBuildPath, err := buildPath.Join(constants.FOLDER_CORE).Abs()
		if err != nil {
			return errors.WithStack(err)
		}

		ctx.PreprocPath = preprocPath
		ctx.SketchBuildPath = sketchBuildPath
		ctx.LibrariesBuildPath = librariesBuildPath
		ctx.CoreBuildPath = coreBuildPath
	}

	if ctx.BuildCachePath != nil {
		coreBuildCachePath, err := ctx.BuildCachePath.Join(constants.FOLDER_CORE).Abs()
		if err != nil {
			return errors.WithStack(err)
		}

		ctx.CoreBuildCachePath = coreBuildCachePath
	}

	if ctx.WarningsLevel == "" {
		ctx.WarningsLevel = DEFAULT_WARNINGS_LEVEL
	}

	ctx.CollectedSourceFiles = &types.UniqueSourceFileQueue{}

	ctx.LibrariesResolutionResults = map[string]types.LibraryResolutionResult{}
	ctx.HardwareRewriteResults = map[*cores.PlatformRelease][]types.PlatforKeyRewrite{}

	return nil
}
