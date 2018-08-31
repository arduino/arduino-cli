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

package phases

import (
	"github.com/arduino/arduino-builder/builder_utils"
	"github.com/arduino/arduino-builder/constants"
	"github.com/arduino/arduino-builder/i18n"
	"github.com/arduino/arduino-builder/types"
	"github.com/arduino/arduino-builder/utils"
)

type SketchBuilder struct{}

func (s *SketchBuilder) Run(ctx *types.Context) error {
	sketchBuildPath := ctx.SketchBuildPath
	buildProperties := ctx.BuildProperties
	includes := utils.Map(ctx.IncludeFolders.AsStrings(), utils.WrapWithHyphenI)

	if err := sketchBuildPath.MkdirAll(); err != nil {
		return i18n.WrapError(err)
	}

	objectFiles, err := builder_utils.CompileFiles(ctx, sketchBuildPath, false, sketchBuildPath, buildProperties, includes)
	if err != nil {
		return i18n.WrapError(err)
	}

	// The "src/" subdirectory of a sketch is compiled recursively
	sketchSrcPath := sketchBuildPath.Join(constants.SKETCH_FOLDER_SRC)
	if sketchSrcPath.IsDir() {
		srcObjectFiles, err := builder_utils.CompileFiles(ctx, sketchSrcPath, true, sketchSrcPath, buildProperties, includes)
		if err != nil {
			return i18n.WrapError(err)
		}
		objectFiles.AddAll(srcObjectFiles)
	}

	ctx.SketchObjectFiles = objectFiles

	return nil
}
