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

package sketch

import (
	"fmt"

	paths "github.com/arduino/go-paths-helper"
)

// WarnDeprecatedFiles warns the user that a type of sketch files are deprecated
func WarnDeprecatedFiles(sketchPath *paths.Path) string {
	if sketchPath.IsNotDir() {
		sketchPath = sketchPath.Parent()
	}

	files, err := sketchPath.ReadDirRecursive()
	if err != nil {
		return ""
	}
	files.FilterSuffix(".pde")

	// .pde files are still supported but deprecated, this warning urges the user to rename them
	if len(files) > 0 {
		msg := tr("Sketches with .pde extension are deprecated, please rename the following files to .ino:")
		for _, f := range files {
			msg += fmt.Sprintf("\n - %s", f)
		}
		return msg
	}
	return ""
}
