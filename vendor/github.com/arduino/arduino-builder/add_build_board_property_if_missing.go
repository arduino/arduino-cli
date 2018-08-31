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
	"os"
	"strings"

	"github.com/arduino/arduino-builder/constants"
	"github.com/arduino/arduino-builder/types"
)

type AddBuildBoardPropertyIfMissing struct{}

func (*AddBuildBoardPropertyIfMissing) Run(ctx *types.Context) error {
	packages := ctx.Hardware
	logger := ctx.GetLogger()

	for _, aPackage := range packages.Packages {
		for _, platform := range aPackage.Platforms {
			for _, platformRelease := range platform.Releases {
				for _, board := range platformRelease.Boards {
					if board.Properties["build.board"] == "" {
						board.Properties["build.board"] = strings.ToUpper(platform.Architecture + "_" + board.BoardID)
						logger.Fprintln(
							os.Stdout,
							constants.LOG_LEVEL_WARN,
							constants.MSG_MISSING_BUILD_BOARD,
							aPackage.Name,
							platform.Architecture,
							board.BoardID,
							board.Properties[constants.BUILD_PROPERTIES_BUILD_BOARD])
					}
				}
			}
		}
	}

	return nil
}
