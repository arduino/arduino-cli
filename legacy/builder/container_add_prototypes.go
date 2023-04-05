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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	bldr "github.com/arduino/arduino-cli/arduino/builder"
	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/go-paths-helper"
	"github.com/pkg/errors"
)

func PreprocessSketchWithCtags(ctx *types.Context) error {
	// Generate the full pathname for the preproc output file
	if err := ctx.PreprocPath.MkdirAll(); err != nil {
		return errors.WithStack(err)
	}
	targetFilePath := ctx.PreprocPath.Join("sketch_merged.cpp")

	// Run preprocessor
	sourceFile := ctx.SketchBuildPath.Join(ctx.Sketch.MainFile.Base() + ".cpp")
	stderr, err := GCCPreprocRunner(ctx, sourceFile, targetFilePath, ctx.IncludeFolders)
	ctx.WriteStderr(stderr)
	if err != nil {
		if !ctx.OnlyUpdateCompilationDatabase {
			return errors.WithStack(err)
		}

		// Do not bail out if we are generating the compile commands database
		ctx.Info(
			fmt.Sprintf("%s: %s",
				tr("An error occurred adding prototypes"),
				tr("the compilation database may be incomplete or inaccurate")))
		if err := sourceFile.CopyTo(targetFilePath); err != nil {
			return errors.WithStack(err)
		}
	}

	if src, err := targetFilePath.ReadFile(); err != nil {
		return err
	} else {
		ctx.SketchSourceAfterCppPreprocessing = filterSketchSource(ctx.Sketch, bytes.NewReader(src), false)
	}

	if err := targetFilePath.WriteFile([]byte(ctx.SketchSourceAfterCppPreprocessing)); err != nil {
		return err
	}

	ctagsStdout, ctagsStderr, err := RunCTags(targetFilePath, ctx.BuildProperties)
	if ctx.Verbose {
		ctx.WriteStderr(ctagsStderr)
	}
	if err != nil {
		return err
	}
	ctx.SketchSourceAfterArduinoPreprocessing = PrototypesAdder(ctx.Sketch, ctx.SketchSourceMerged, ctagsStdout, ctx.LineOffset)

	if err := bldr.SketchSaveItemCpp(ctx.Sketch.MainFile, []byte(ctx.SketchSourceAfterArduinoPreprocessing), ctx.SketchBuildPath); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func filterSketchSource(sketch *sketch.Sketch, source io.Reader, removeLineMarkers bool) string {
	fileNames := paths.NewPathList()
	fileNames.Add(sketch.MainFile)
	fileNames.AddAll(sketch.OtherSketchFiles)

	inSketch := false
	filtered := ""

	scanner := bufio.NewScanner(source)
	for scanner.Scan() {
		line := scanner.Text()
		if filename := parseLineMarker(line); filename != nil {
			inSketch = fileNames.Contains(filename)
			if inSketch && removeLineMarkers {
				continue
			}
		}

		if inSketch {
			filtered += line + "\n"
		}
	}

	return filtered
}

// Parses the given line as a gcc line marker and returns the contained
// filename.
func parseLineMarker(line string) *paths.Path {
	// A line marker contains the line number and filename and looks like:
	// # 123 /path/to/file.cpp
	// It can be followed by zero or more flag number that indicate the
	// preprocessor state and can be ignored.
	// For exact details on this format, see:
	// https://github.com/gcc-mirror/gcc/blob/edd716b6b1caa1a5cb320a8cd7f626f30198e098/gcc/c-family/c-ppoutput.c#L413-L415

	split := strings.SplitN(line, " ", 3)
	if len(split) < 3 || len(split[0]) == 0 || split[0][0] != '#' {
		return nil
	}

	_, err := strconv.Atoi(split[1])
	if err != nil {
		return nil
	}

	// If we get here, we found a # followed by a line number, so
	// assume this is a line marker and see if the rest of the line
	// starts with a string containing the filename
	str, rest, ok := bldr.ParseCppString(split[2])

	if ok && (rest == "" || rest[0] == ' ') {
		return paths.New(str)
	}
	return nil
}
