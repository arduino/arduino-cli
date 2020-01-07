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
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/ctags"
	"github.com/arduino/arduino-cli/legacy/builder/i18n"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
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
