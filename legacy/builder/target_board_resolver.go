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
	"strings"

	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/i18n"
	"github.com/arduino/arduino-cli/legacy/builder/types"
)

type TargetBoardResolver struct{}

func (s *TargetBoardResolver) Run(ctx *types.Context) error {
	logger := ctx.GetLogger()

	targetPackage, targetPlatform, targetBoard, buildProperties, actualPlatform, err := ctx.PackageManager.ResolveFQBN(ctx.FQBN)
	if err != nil {
		return i18n.ErrorfWithLogger(logger, "Error resolving FQBN: {0}", err)
	}

	targetBoard.Properties = buildProperties // FIXME....

	core := targetBoard.Properties.Get("build.core")
	if core == "" {
		core = "arduino"
	}
	// select the core name in case of "package:core" format
	core = core[strings.Index(core, ":")+1:]

	if ctx.Verbose {
		logger.Println(constants.LOG_LEVEL_INFO, constants.MSG_USING_BOARD, targetBoard.BoardID, targetPlatform.InstallDir)
		logger.Println(constants.LOG_LEVEL_INFO, constants.MSG_USING_CORE, core, actualPlatform.InstallDir)
	}

	ctx.BuildCore = core
	ctx.TargetBoard = targetBoard
	ctx.TargetPlatform = targetPlatform
	ctx.TargetPackage = targetPackage
	ctx.ActualPlatform = actualPlatform
	return nil
}
