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
	"bufio"
	"strconv"
	"strings"

	"github.com/arduino/arduino-builder/types"
	"github.com/arduino/arduino-builder/utils"
)

type FilterSketchSource struct {
	Source *string
}

func (s *FilterSketchSource) Run(ctx *types.Context) error {
	fileNames := []string{ctx.Sketch.MainFile.Name}
	for _, file := range ctx.Sketch.OtherSketchFiles {
		fileNames = append(fileNames, file.Name)
	}

	inSketch := false
	filtered := ""

	scanner := bufio.NewScanner(strings.NewReader(*s.Source))
	for scanner.Scan() {
		line := scanner.Text()
		filename := parseLineMarker(line)
		if filename != "" {
			inSketch = utils.SliceContains(fileNames, filename)
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
func parseLineMarker(line string) string {
	// A line marker contains the line number and filename and looks like:
	// # 123 /path/to/file.cpp
	// It can be followed by zero or more flag number that indicate the
	// preprocessor state and can be ignored.
	// For exact details on this format, see:
	// https://github.com/gcc-mirror/gcc/blob/edd716b6b1caa1a5cb320a8cd7f626f30198e098/gcc/c-family/c-ppoutput.c#L413-L415

	split := strings.SplitN(line, " ", 3)
	if len(split) < 3 || split[0] != "#" {
		return ""
	}

	_, err := strconv.Atoi(split[1])
	if err != nil {
		return ""
	}

	// If we get here, we found a # followed by a line number, so
	// assume this is a line marker and see if the rest of the line
	// starts with a string containing the filename
	str, rest, ok := utils.ParseCppString(split[2])

	if ok && (rest == "" || rest[0] == ' ') {
		return str
	}
	return ""
}
