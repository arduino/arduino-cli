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
	"path/filepath"
	"strings"

	"github.com/arduino/arduino-builder/builder_utils"
	"github.com/arduino/arduino-builder/constants"
	"github.com/arduino/arduino-builder/i18n"
	"github.com/arduino/arduino-builder/types"
	"github.com/arduino/arduino-builder/utils"
	"github.com/arduino/go-properties-map"
)

type GCCPreprocRunner struct {
	SourceFilePath string
	TargetFileName string
	Includes       []string
}

func (s *GCCPreprocRunner) Run(ctx *types.Context) error {
	properties, targetFilePath, err := prepareGCCPreprocRecipeProperties(ctx, s.SourceFilePath, s.TargetFileName, s.Includes)
	if err != nil {
		return i18n.WrapError(err)
	}

	if properties[constants.RECIPE_PREPROC_MACROS] == constants.EMPTY_STRING {
		//generate PREPROC_MACROS from RECIPE_CPP_PATTERN
		properties[constants.RECIPE_PREPROC_MACROS] = GeneratePreprocPatternFromCompile(properties[constants.RECIPE_CPP_PATTERN])
	}

	verbose := ctx.Verbose
	logger := ctx.GetLogger()
	_, err = builder_utils.ExecRecipe(properties, constants.RECIPE_PREPROC_MACROS, true, verbose, false, logger)
	if err != nil {
		return i18n.WrapError(err)
	}

	ctx.FileToRead = targetFilePath

	return nil
}

type GCCPreprocRunnerForDiscoveringIncludes struct {
	SourceFilePath string
	TargetFilePath string
	Includes       []string
}

func (s *GCCPreprocRunnerForDiscoveringIncludes) Run(ctx *types.Context) error {
	properties, _, err := prepareGCCPreprocRecipeProperties(ctx, s.SourceFilePath, s.TargetFilePath, s.Includes)
	if err != nil {
		return i18n.WrapError(err)
	}

	verbose := ctx.Verbose
	logger := ctx.GetLogger()

	if properties[constants.RECIPE_PREPROC_MACROS] == constants.EMPTY_STRING {
		//generate PREPROC_MACROS from RECIPE_CPP_PATTERN
		properties[constants.RECIPE_PREPROC_MACROS] = GeneratePreprocPatternFromCompile(properties[constants.RECIPE_CPP_PATTERN])
	}

	stderr, err := builder_utils.ExecRecipeCollectStdErr(properties, constants.RECIPE_PREPROC_MACROS, true, verbose, logger)
	if err != nil {
		return i18n.WrapError(err)
	}

	ctx.SourceGccMinusE = string(stderr)

	return nil
}

func prepareGCCPreprocRecipeProperties(ctx *types.Context, sourceFilePath string, targetFilePath string, includes []string) (properties.Map, string, error) {
	if targetFilePath != utils.NULLFile() {
		preprocPath := ctx.PreprocPath
		err := utils.EnsureFolderExists(preprocPath)
		if err != nil {
			return nil, "", i18n.WrapError(err)
		}
		targetFilePath = filepath.Join(preprocPath, targetFilePath)
	}

	properties := ctx.BuildProperties.Clone()
	properties[constants.BUILD_PROPERTIES_SOURCE_FILE] = sourceFilePath
	properties[constants.BUILD_PROPERTIES_PREPROCESSED_FILE_PATH] = targetFilePath

	includes = utils.Map(includes, utils.WrapWithHyphenI)
	properties[constants.BUILD_PROPERTIES_INCLUDES] = strings.Join(includes, constants.SPACE)
	builder_utils.RemoveHyphenMDDFlagFromGCCCommandLine(properties)

	return properties, targetFilePath, nil
}

func GeneratePreprocPatternFromCompile(compilePattern string) string {
	// add {preproc.macros.flags}
	// replace "{object_file}" with "{preprocessed_file_path}"
	returnString := compilePattern
	returnString = strings.Replace(returnString, "{compiler.cpp.flags}", "{compiler.cpp.flags} {preproc.macros.flags}", 1)
	returnString = strings.Replace(returnString, "{object_file}", "{preprocessed_file_path}", 1)
	return returnString
}
