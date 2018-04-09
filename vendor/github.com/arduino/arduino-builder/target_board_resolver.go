/*
 * This file is part of Arduino Builder.
 *
 * Arduino Builder is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2015 Arduino LLC (http://www.arduino.cc/)
 */

package builder

import (
	"strings"

	"github.com/arduino/arduino-builder/constants"
	"github.com/arduino/arduino-builder/i18n"
	"github.com/arduino/arduino-builder/types"
	"github.com/bcmi-labs/arduino-cli/cores"
)

type TargetBoardResolver struct{}

func (s *TargetBoardResolver) Run(ctx *types.Context) error {
	logger := ctx.GetLogger()

	fqbn := ctx.FQBN

	fqbnParts := strings.Split(fqbn, ":")
	if len(fqbnParts) < 3 {
		return i18n.ErrorfWithLogger(logger, constants.MSG_FQBN_INVALID, fqbn)
	}
	targetPackageName := fqbnParts[0]
	targetPlatformName := fqbnParts[1]
	targetBoardName := fqbnParts[2]

	packages := ctx.Hardware

	targetPackage := packages.Packages[targetPackageName]
	if targetPackage == nil {
		return i18n.ErrorfWithLogger(logger, constants.MSG_PACKAGE_UNKNOWN, targetPackageName)
	}

	targetPlatforms := targetPackage.Platforms[targetPlatformName]
	if targetPlatforms == nil {
		return i18n.ErrorfWithLogger(logger, constants.MSG_PLATFORM_UNKNOWN, targetPlatformName, targetPackageName)
	}
	targetPlatform := targetPlatforms.GetInstalled()
	if targetPlatform == nil {
		return i18n.ErrorfWithLogger(logger, constants.MSG_PLATFORM_UNKNOWN, targetPlatformName, targetPackageName)
	}

	targetBoard := targetPlatform.Boards[targetBoardName]
	if targetBoard == nil {
		return i18n.ErrorfWithLogger(logger, constants.MSG_BOARD_UNKNOWN, targetBoardName, targetPlatformName, targetPackageName)
	}

	ctx.TargetPlatform = targetPlatform
	ctx.TargetPackage = targetPackage
	ctx.TargetBoard = targetBoard

	if len(fqbnParts) > 3 {
		if props, err := targetBoard.GeneratePropertiesForConfiguration(fqbnParts[3]); err != nil {
			i18n.ErrorfWithLogger(logger, "Error in FQBN: %s", err)
		} else {
			targetBoard.Properties = props
		}
	}

	core := targetBoard.Properties["build.core"]
	if core == "" {
		core = "arduino"
	}

	var corePlatform *cores.PlatformRelease
	coreParts := strings.Split(core, ":")
	if len(coreParts) > 1 {
		core = coreParts[1]
		if packages.Packages[coreParts[0]] == nil {
			return i18n.ErrorfWithLogger(logger, constants.MSG_MISSING_CORE_FOR_BOARD, coreParts[0])

		}
		corePlatform = packages.Packages[coreParts[0]].Platforms[targetPlatforms.Architecture].GetInstalled()
	}

	var actualPlatform *cores.PlatformRelease
	if corePlatform != nil {
		actualPlatform = corePlatform
	} else {
		actualPlatform = targetPlatform
	}

	if ctx.Verbose {
		logger.Println(constants.LOG_LEVEL_INFO, constants.MSG_USING_BOARD, targetBoard.BoardId, targetPlatform.Folder)
		logger.Println(constants.LOG_LEVEL_INFO, constants.MSG_USING_CORE, core, actualPlatform.Folder)
	}

	ctx.BuildCore = core
	ctx.ActualPlatform = actualPlatform

	return nil
}
