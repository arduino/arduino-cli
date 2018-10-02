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
	"github.com/arduino/arduino-builder/ctags"
	"github.com/arduino/arduino-builder/i18n"
	"github.com/arduino/arduino-builder/types"
	"github.com/arduino/arduino-builder/utils"
)

type CTagsRunner struct{}

func (s *CTagsRunner) Run(ctx *types.Context) error {
	buildProperties := ctx.BuildProperties
	ctagsTargetFilePath := ctx.CTagsTargetFile
	logger := ctx.GetLogger()

	properties := buildProperties.Clone()
	properties.Merge(buildProperties.SubTree(constants.BUILD_PROPERTIES_TOOLS_KEY).SubTree(constants.CTAGS))
	properties.SetPath(constants.BUILD_PROPERTIES_SOURCE_FILE, ctagsTargetFilePath)

	pattern := properties.Get(constants.BUILD_PROPERTIES_PATTERN)
	if pattern == constants.EMPTY_STRING {
		return i18n.ErrorfWithLogger(logger, constants.MSG_PATTERN_MISSING, constants.CTAGS)
	}

	commandLine := properties.ExpandPropsInString(pattern)
	command, err := utils.PrepareCommand(commandLine, logger, "")
	if err != nil {
		return i18n.WrapError(err)
	}

	sourceBytes, _, err := utils.ExecCommand(ctx, command, utils.Capture /* stdout */, utils.Ignore /* stderr */)
	if err != nil {
		return i18n.WrapError(err)
	}

	ctx.CTagsOutput = string(sourceBytes)

	parser := &ctags.CTagsParser{}

	ctx.CTagsOfPreprocessedSource = parser.Parse(ctx.CTagsOutput, ctx.Sketch.MainFile.Name)
	parser.FixCLinkageTagsDeclarations(ctx.CTagsOfPreprocessedSource)

	protos, line := parser.GeneratePrototypes()
	if line != -1 {
		ctx.PrototypesLineWhereToInsert = line
	}
	ctx.Prototypes = protos

	return nil
}
