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

package phases

import (
	"io"

	"github.com/arduino/arduino-cli/arduino/builder/compilation"
	"github.com/arduino/arduino-cli/arduino/builder/cpp"
	"github.com/arduino/arduino-cli/arduino/builder/progress"
	"github.com/arduino/arduino-cli/arduino/builder/utils"
	f "github.com/arduino/arduino-cli/internal/algorithms"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
)

func SketchBuilder(
	sketchBuildPath *paths.Path,
	buildProperties *properties.Map,
	includesFolders paths.PathList,
	onlyUpdateCompilationDatabase, verbose bool,
	compilationDatabase *compilation.Database,
	jobs int,
	warningsLevel string,
	stdoutWriter, stderrWriter io.Writer,
	verboseInfoFn func(msg string),
	verboseStdoutFn, verboseStderrFn func(data []byte),
	progress *progress.Struct, progressCB rpc.TaskProgressCB,
) (paths.PathList, error) {
	includes := f.Map(includesFolders.AsStrings(), cpp.WrapWithHyphenI)

	if err := sketchBuildPath.MkdirAll(); err != nil {
		return nil, errors.WithStack(err)
	}

	sketchObjectFiles, err := utils.CompileFiles(
		sketchBuildPath, sketchBuildPath, buildProperties, includes,
		onlyUpdateCompilationDatabase,
		compilationDatabase,
		jobs,
		verbose,
		warningsLevel,
		stdoutWriter, stderrWriter,
		verboseInfoFn,
		verboseStdoutFn,
		verboseStderrFn,
		progress, progressCB,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// The "src/" subdirectory of a sketch is compiled recursively
	sketchSrcPath := sketchBuildPath.Join("src")
	if sketchSrcPath.IsDir() {
		srcObjectFiles, err := utils.CompileFilesRecursive(
			sketchSrcPath, sketchSrcPath, buildProperties, includes,
			onlyUpdateCompilationDatabase,
			compilationDatabase,
			jobs,
			verbose,
			warningsLevel,
			stdoutWriter, stderrWriter,
			verboseInfoFn,
			verboseStdoutFn,
			verboseStderrFn,
			progress, progressCB,
		)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		sketchObjectFiles.AddAll(srcObjectFiles)
	}

	return sketchObjectFiles, nil
}
