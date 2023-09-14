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

package builder

import (
	"github.com/arduino/arduino-cli/arduino/builder/preprocessor"
	"github.com/arduino/go-paths-helper"
)

// preprocessSketch fixdoc
func (b *Builder) preprocessSketch(includes paths.PathList) error {
	// In the future we might change the preprocessor
	normalOutput, verboseOutput, err := preprocessor.PreprocessSketchWithCtags(
		b.sketch, b.buildPath, includes, b.lineOffset,
		b.buildProperties, b.onlyUpdateCompilationDatabase,
	)
	if b.logger.Verbose() {
		b.logger.WriteStdout(verboseOutput)
	}
	b.logger.WriteStdout(normalOutput)

	return err
}
