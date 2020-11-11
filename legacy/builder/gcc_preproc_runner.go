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
	"os/exec"
	"strings"

	"github.com/arduino/arduino-cli/legacy/builder/builder_utils"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
	"github.com/arduino/go-paths-helper"
	"github.com/pkg/errors"
)

func GCCPreprocRunner(ctx *types.Context, sourceFilePath *paths.Path, targetFilePath *paths.Path, includes paths.PathList) error {
	cmd, err := prepareGCCPreprocRecipeProperties(ctx, sourceFilePath, targetFilePath, includes)
	if err != nil {
		return errors.WithStack(err)
	}

	_, _, err = utils.ExecCommand(ctx, cmd /* stdout */, utils.ShowIfVerbose /* stderr */, utils.Show)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func GCCPreprocRunnerForDiscoveringIncludes(ctx *types.Context, sourceFilePath *paths.Path, targetFilePath *paths.Path, includes paths.PathList) ([]byte, error) {
	cmd, err := prepareGCCPreprocRecipeProperties(ctx, sourceFilePath, targetFilePath, includes)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	_, stderr, err := utils.ExecCommand(ctx, cmd /* stdout */, utils.ShowIfVerbose /* stderr */, utils.Capture)
	if err != nil {
		return stderr, errors.WithStack(err)
	}

	return stderr, nil
}

func prepareGCCPreprocRecipeProperties(ctx *types.Context, sourceFilePath *paths.Path, targetFilePath *paths.Path, includes paths.PathList) (*exec.Cmd, error) {
	properties := ctx.BuildProperties.Clone()
	properties.Set("build.library_discovery_phase", "1")
	properties.SetPath(constants.BUILD_PROPERTIES_SOURCE_FILE, sourceFilePath)
	properties.SetPath(constants.BUILD_PROPERTIES_PREPROCESSED_FILE_PATH, targetFilePath)

	includesStrings := utils.Map(includes.AsStrings(), utils.WrapWithHyphenI)
	properties.Set(constants.BUILD_PROPERTIES_INCLUDES, strings.Join(includesStrings, constants.SPACE))

	if properties.Get(constants.RECIPE_PREPROC_MACROS) == "" {
		//generate PREPROC_MACROS from RECIPE_CPP_PATTERN
		properties.Set(constants.RECIPE_PREPROC_MACROS, GeneratePreprocPatternFromCompile(properties.Get(constants.RECIPE_CPP_PATTERN)))
	}

	cmd, err := builder_utils.PrepareCommandForRecipe(properties, constants.RECIPE_PREPROC_MACROS, true)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Remove -MMD argument if present. Leaving it will make gcc try
	// to create a /dev/null.d dependency file, which won't work.
	cmd.Args = utils.Filter(cmd.Args, func(a string) bool { return a != "-MMD" })

	return cmd, nil
}

func GeneratePreprocPatternFromCompile(compilePattern string) string {
	// add {preproc.macros.flags}
	// replace "{object_file}" with "{preprocessed_file_path}"
	returnString := compilePattern
	returnString = strings.Replace(returnString, "{compiler.cpp.flags}", "{compiler.cpp.flags} {preproc.macros.flags}", 1)
	returnString = strings.Replace(returnString, "{object_file}", "{preprocessed_file_path}", 1)
	return returnString
}
