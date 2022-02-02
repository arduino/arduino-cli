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
	"strconv"
	"strings"

	"github.com/arduino/go-paths-helper"

	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
)

type FilterSketchSource struct {
	Source            *string
	RemoveLineMarkers bool
}

func (s *FilterSketchSource) Run(ctx *types.Context) error {
	fileNames := paths.NewPathList()
	fileNames.Add(ctx.Sketch.MainFile)
	fileNames.AddAll(ctx.Sketch.OtherSketchFiles)

	inSketch := false
	filtered := ""

	scanner := bufio.NewScanner(strings.NewReader(*s.Source))
	for scanner.Scan() {
		line := scanner.Text()
		if filename := parseLineMarker(line); filename != nil {
			inSketch = fileNames.Contains(filename)
			if inSketch && s.RemoveLineMarkers {
				continue
			}
		}

		if inSketch {
			filtered += line + "\n"
		}
	}

	*s.Source = filtered
	return nil
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
	str, rest, ok := utils.ParseCppString(split[2])

	if ok && (rest == "" || rest[0] == ' ') {
		return paths.New(str)
	}
	return nil
}
