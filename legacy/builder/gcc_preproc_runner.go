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

	f "github.com/arduino/arduino-cli/internal/algorithms"
	"github.com/arduino/arduino-cli/legacy/builder/builder_utils"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
	"github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
)

func GCCPreprocRunner(ctx *types.Context, sourceFilePath *paths.Path, targetFilePath *paths.Path, includes paths.PathList) ([]byte, error) {
	cmd, err := prepareGCCPreprocRecipeProperties(ctx, sourceFilePath, targetFilePath, includes)
	if err != nil {
		return nil, err
	}
	_, stderr, err := utils.ExecCommand(ctx, cmd, utils.ShowIfVerbose /* stdout */, utils.Capture /* stderr */)
	return stderr, err
}

func prepareGCCPreprocRecipeProperties(ctx *types.Context, sourceFilePath *paths.Path, targetFilePath *paths.Path, includes paths.PathList) (*exec.Cmd, error) {
	buildProperties := properties.NewMap()
	buildProperties.Set("preproc.macros.flags", "-w -x c++ -E -CC")
	buildProperties.Merge(ctx.BuildProperties)
	buildProperties.Set("build.library_discovery_phase", "1")
	buildProperties.SetPath("source_file", sourceFilePath)
	buildProperties.SetPath("preprocessed_file_path", targetFilePath)

	includesStrings := f.Map(includes.AsStrings(), utils.WrapWithHyphenI)
	buildProperties.Set("includes", strings.Join(includesStrings, " "))

	if buildProperties.Get("recipe.preproc.macros") == "" {
		// autogenerate preprocess macros recipe from compile recipe
		preprocPattern := buildProperties.Get("recipe.cpp.o.pattern")
		// add {preproc.macros.flags} to {compiler.cpp.flags}
		preprocPattern = strings.Replace(preprocPattern, "{compiler.cpp.flags}", "{compiler.cpp.flags} {preproc.macros.flags}", 1)
		// replace "{object_file}" with "{preprocessed_file_path}"
		preprocPattern = strings.Replace(preprocPattern, "{object_file}", "{preprocessed_file_path}", 1)

		buildProperties.Set("recipe.preproc.macros", preprocPattern)
	}

	cmd, err := builder_utils.PrepareCommandForRecipe(buildProperties, "recipe.preproc.macros", true, ctx.PackageManager.GetEnvVarsForSpawnedProcess())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Remove -MMD argument if present. Leaving it will make gcc try
	// to create a /dev/null.d dependency file, which won't work.
	cmd.Args = f.Filter(cmd.Args, f.NotEquals("-MMD"))

	return cmd, nil
}
