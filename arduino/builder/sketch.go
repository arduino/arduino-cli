// This file is part of arduino-cli.
//
// Copyright 2019 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to modify or
// otherwise use the software for commercial activities involving the Arduino
// software without disclosing the source code of your own applications. To purchase
// a commercial license, send an email to license@arduino.cc.

package builder

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// SaveSketch saves a preprocessed .cpp sketch file on disk
func SaveSketch(sketchName string, source string, buildPath string) error {

	if err := os.MkdirAll(buildPath, os.FileMode(0755)); err != nil {
		return errors.Wrap(err, "unable to create a folder to save the sketch")
	}

	destFile := filepath.Join(buildPath, sketchName+".cpp")

	if err := ioutil.WriteFile(destFile, []byte(source), os.FileMode(0644)); err != nil {
		return errors.Wrap(err, "unable to save the sketch on disk")
	}

	return nil
}
