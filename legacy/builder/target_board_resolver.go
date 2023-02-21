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
	"fmt"
	"strings"

	"github.com/arduino/arduino-cli/legacy/builder/types"
)

type TargetBoardResolver struct{}

func (s *TargetBoardResolver) Run(ctx *types.Context) error {
	targetPackage, targetPlatform, targetBoard, buildProperties, buildPlatform, err := ctx.PackageManager.ResolveFQBN(ctx.FQBN)
	if err != nil {
		return fmt.Errorf("%s: %w", tr("Error resolving FQBN"), err)
	}

	if ctx.Verbose {
		core := buildProperties.Get("build.core")
		if core == "" {
			core = "arduino"
		}
		// select the core name in case of "package:core" format
		core = core[strings.Index(core, ":")+1:]
		ctx.Info(tr("Using board '%[1]s' from platform in folder: %[2]s", targetBoard.BoardID, targetPlatform.InstallDir))
		ctx.Info(tr("Using core '%[1]s' from platform in folder: %[2]s", core, buildPlatform.InstallDir))
	}

	if buildProperties.Get("build.board") == "" {
		architecture := targetBoard.PlatformRelease.Platform.Architecture
		defaultBuildBoard := strings.ToUpper(architecture + "_" + targetBoard.BoardID)
		buildProperties.Set("build.board", defaultBuildBoard)
		ctx.Info(tr("Warning: Board %[1]s doesn't define a %[2]s preference. Auto-set to: %[3]s",
			targetBoard.String(), "'build.board'", defaultBuildBoard))
	}

	requiredTools, err := ctx.PackageManager.FindToolsRequiredForBuild(targetPlatform, buildPlatform)
	if err != nil {
		return err
	}

	ctx.TargetBoard = targetBoard
	ctx.TargetBoardBuildProperties = buildProperties
	ctx.TargetPlatform = targetPlatform
	ctx.TargetPackage = targetPackage
	ctx.ActualPlatform = buildPlatform
	ctx.RequiredTools = requiredTools
	return nil
}
