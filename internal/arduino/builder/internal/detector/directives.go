// This file is part of arduino-cli.
//
// Copyright 2024 ARDUINO SA (http://www.arduino.cc/)
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

package detector

import (
	"bytes"

	"github.com/arduino/go-paths-helper"
)

// sketchDirectives returns the lines of the merged sketch that may be preprocessor directives:
// every line holding a #, with the lines a trailing backslash continues them on. Lines that
// only carry a # in code or comments come along. The lines cannot tell a directive that a
// comment or a condition turns off: they only select the edits worth the preprocessor pass
// that confirms a replay of the detection.
func sketchDirectives(mergedSketch *paths.Path) ([]string, error) {
	source, err := mergedSketch.ReadFile()
	if err != nil {
		return nil, err
	}
	return directiveLines(source), nil
}

func directiveLines(source []byte) []string {
	directives := []string{}
	continued := false
	for line := range bytes.Lines(source) {
		line = bytes.TrimRight(line, "\r\n")
		if !continued && !bytes.Contains(line, []byte("#")) {
			continue
		}
		directives = append(directives, string(line))
		continued = bytes.HasSuffix(line, []byte("\\"))
	}
	return directives
}
