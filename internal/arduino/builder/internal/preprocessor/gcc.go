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
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/arduino/arduino-cli/internal/arduino/builder/cpp"
	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"go.bug.st/f"
)

// GCC performs a run of the gcc preprocess (macro/includes expansion). The function outputs the result
// to targetFilePath. Returns the stdout/stderr of gcc if any.
func GCC(
	ctx context.Context,
	sourceFilePath, targetFilePath *paths.Path,
	includes paths.PathList, buildProperties *properties.Map,
) (Result, error) {
	gccBuildProperties := properties.NewMap()
	gccBuildProperties.Set("preproc.macros.flags", "-w -x c++ -E -CC")
	gccBuildProperties.Merge(buildProperties)
	gccBuildProperties.Set("build.library_discovery_phase", "1")
	gccBuildProperties.SetPath("source_file", sourceFilePath)
	gccBuildProperties.SetPath("preprocessed_file_path", targetFilePath)

	includesStrings := f.Map(includes.AsStrings(), cpp.WrapWithHyphenI)
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
		return Result{}, errors.New(i18n.Tr("%s pattern is missing", gccPreprocRecipeProperty))
	}

	commandLine := gccBuildProperties.ExpandPropsInString(pattern)
	commandLine = properties.DeleteUnexpandedPropsFromString(commandLine)
	args, err := properties.SplitQuotedString(commandLine, `"'`, false)
	if err != nil {
		return Result{}, err
	}

	// Remove -MMD argument if present. Leaving it will make gcc try
	// to create a /dev/null.d dependency file, which won't work.
	args = f.Filter(args, f.NotEquals("-MMD"))

	proc, err := paths.NewProcess(nil, args...)
	if err != nil {
		return Result{}, err
	}

	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	count := 0
	stderrLimited := writerFunc(func(p []byte) (int, error) {
		// Limit the size of the stderr buffer to 100KB
		n, err := stderr.Write(p)
		count += n
		if count > 100*1024 {
			cancel()
			fmt.Fprintln(stderr, i18n.Tr("Compiler error output has been truncated."))
		}
		return n, err
	})

	proc.RedirectStdoutTo(stdout)
	proc.RedirectStderrTo(stderrLimited)

	// Append gcc arguments to stdout before running the command
	fmt.Fprintln(stdout, strings.Join(args, " "))

	if err := proc.Start(); err != nil {
		return Result{}, err
	}

	// Wait for the process to finish
	err = proc.WaitWithinContext(ctx)

	return Result{args: proc.GetArgs(), stdout: stdout.Bytes(), stderr: stderr.Bytes()}, err
}

type writerFunc func(p []byte) (n int, err error)

func (f writerFunc) Write(p []byte) (n int, err error) {
	return f(p)
}
