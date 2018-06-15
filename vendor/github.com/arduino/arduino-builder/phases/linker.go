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
	"strings"

	"github.com/arduino/arduino-builder/builder_utils"
	"github.com/arduino/arduino-builder/constants"
	"github.com/arduino/arduino-builder/i18n"
	"github.com/arduino/arduino-builder/types"
	"github.com/arduino/arduino-builder/utils"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-map"
)

type Linker struct{}

func (s *Linker) Run(ctx *types.Context) error {
	objectFilesSketch := ctx.SketchObjectFiles
	objectFilesLibraries := ctx.LibrariesObjectFiles
	objectFilesCore := ctx.CoreObjectsFiles

	objectFiles := paths.NewPathList()
	objectFiles.AddAll(objectFilesSketch)
	objectFiles.AddAll(objectFilesLibraries)
	objectFiles.AddAll(objectFilesCore)

	coreArchiveFilePath := ctx.CoreArchiveFilePath
	buildPath := ctx.BuildPath
	coreDotARelPath, err := buildPath.RelTo(coreArchiveFilePath)
	if err != nil {
		return i18n.WrapError(err)
	}

	buildProperties := ctx.BuildProperties

	err = link(ctx, objectFiles, coreDotARelPath, coreArchiveFilePath, buildProperties)
	if err != nil {
		return i18n.WrapError(err)
	}

	return nil
}

func link(ctx *types.Context, objectFiles paths.PathList, coreDotARelPath *paths.Path, coreArchiveFilePath *paths.Path, buildProperties properties.Map) error {
	optRelax := addRelaxTrickIfATMEGA2560(buildProperties)

	quotedObjectFiles := utils.Map(objectFiles.AsStrings(), wrapWithDoubleQuotes)
	objectFileList := strings.Join(quotedObjectFiles, constants.SPACE)

	properties := buildProperties.Clone()
	properties[constants.BUILD_PROPERTIES_COMPILER_C_ELF_FLAGS] = properties[constants.BUILD_PROPERTIES_COMPILER_C_ELF_FLAGS] + optRelax
	properties[constants.BUILD_PROPERTIES_COMPILER_WARNING_FLAGS] = properties[constants.BUILD_PROPERTIES_COMPILER_WARNING_FLAGS+"."+ctx.WarningsLevel]
	properties[constants.BUILD_PROPERTIES_ARCHIVE_FILE] = coreDotARelPath.String()
	properties[constants.BUILD_PROPERTIES_ARCHIVE_FILE_PATH] = coreArchiveFilePath.String()
	properties[constants.BUILD_PROPERTIES_OBJECT_FILES] = objectFileList

	_, _, err := builder_utils.ExecRecipe(ctx, properties, constants.RECIPE_C_COMBINE_PATTERN, false /* stdout */, utils.ShowIfVerbose /* stderr */, utils.Show)
	return err
}

func wrapWithDoubleQuotes(value string) string {
	return "\"" + value + "\""
}

func addRelaxTrickIfATMEGA2560(buildProperties properties.Map) string {
	if buildProperties[constants.BUILD_PROPERTIES_BUILD_MCU] == "atmega2560" {
		return ",--relax"
	}
	return constants.EMPTY_STRING
}
