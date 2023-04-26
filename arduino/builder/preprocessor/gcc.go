// This file is part of arduino-cli.
//
// Copyright 2023 ARDUINO SA (http://www.arduino.cc/)
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

package preprocessor

import (
	"context"
	"fmt"
	"strings"

	"github.com/arduino/arduino-cli/executils"
	f "github.com/arduino/arduino-cli/internal/algorithms"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
)

func GCC(sourceFilePath *paths.Path, targetFilePath *paths.Path, includes paths.PathList, buildProperties *properties.Map) ([]byte, []byte, error) {
	gccBuildProperties := properties.NewMap()
	gccBuildProperties.Set("preproc.macros.flags", "-w -x c++ -E -CC")
	gccBuildProperties.Merge(buildProperties)
	gccBuildProperties.Set("build.library_discovery_phase", "1")
	gccBuildProperties.SetPath("source_file", sourceFilePath)
	gccBuildProperties.SetPath("preprocessed_file_path", targetFilePath)

	includesStrings := f.Map(includes.AsStrings(), utils.WrapWithHyphenI)
	gccBuildProperties.Set("includes", strings.Join(includesStrings, " "))

	const gccPreprocRecipeProperty = "recipe.preproc.macros"
	if gccBuildProperties.Get(gccPreprocRecipeProperty) == "" {
		// autogenerate preprocess macros recipe from compile recipe
		preprocPattern := gccBuildProperties.Get("recipe.cpp.o.pattern")
		// add {preproc.macros.flags} to {compiler.cpp.flags}
		preprocPattern = strings.Replace(preprocPattern, "{compiler.cpp.flags}", "{compiler.cpp.flags} {preproc.macros.flags}", 1)
		// replace "{object_file}" with "{preprocessed_file_path}"
		preprocPattern = strings.Replace(preprocPattern, "{object_file}", "{preprocessed_file_path}", 1)

		gccBuildProperties.Set(gccPreprocRecipeProperty, preprocPattern)
	}

	pattern := gccBuildProperties.Get(gccPreprocRecipeProperty)
	if pattern == "" {
		return nil, nil, errors.Errorf(tr("%s pattern is missing"), gccPreprocRecipeProperty)
	}

	commandLine := gccBuildProperties.ExpandPropsInString(pattern)
	args, err := properties.SplitQuotedString(commandLine, `"'`, false)
	if err != nil {
		return nil, nil, err
	}

	// Remove -MMD argument if present. Leaving it will make gcc try
	// to create a /dev/null.d dependency file, which won't work.
	args = f.Filter(args, f.NotEquals("-MMD"))

	proc, err := executils.NewProcess(nil, args...)
	if err != nil {
		return nil, nil, err
	}
	stdout, stderr, err := proc.RunAndCaptureOutput(context.Background())

	// Append gcc arguments to stdout
	stdout = append([]byte(fmt.Sprintln(strings.Join(args, " "))), stdout...)

	return stdout, stderr, err
}
