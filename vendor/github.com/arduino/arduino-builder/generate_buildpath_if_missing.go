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
	"github.com/arduino/arduino-builder/constants"
	"github.com/arduino/arduino-builder/i18n"
	"github.com/arduino/arduino-builder/types"
	"github.com/arduino/arduino-builder/utils"
	"os"
	"path/filepath"
	"strings"
)

type GenerateBuildPathIfMissing struct{}

func (s *GenerateBuildPathIfMissing) Run(ctx *types.Context) error {
	if ctx.BuildPath != "" {
		return nil
	}

	md5sum := utils.MD5Sum([]byte(ctx.SketchLocation))

	buildPath := filepath.Join(os.TempDir(), "arduino-sketch-"+strings.ToUpper(md5sum))
	_, err := os.Stat(buildPath)
	if err != nil && !os.IsNotExist(err) {
		return i18n.WrapError(err)
	}

	if ctx.DebugLevel > 5 {
		logger := ctx.GetLogger()
		logger.Fprintln(os.Stdout, constants.LOG_LEVEL_WARN, constants.MSG_SETTING_BUILD_PATH, buildPath)
	}

	ctx.BuildPath = buildPath

	return nil
}
