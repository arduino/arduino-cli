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
	"crypto/md5"
	"encoding/hex"
	"os"
	"strings"

	"github.com/arduino/go-paths-helper"
	"github.com/pkg/errors"
)

// GenBuildPath generates a suitable name for the build folder.
// The sketchPath, if not nil, is also used to furhter differentiate build paths.
func GenBuildPath(sketchPath *paths.Path) *paths.Path {
	path := ""
	if sketchPath != nil {
		path = sketchPath.String()
	}
	md5SumBytes := md5.Sum([]byte(path))
	md5Sum := strings.ToUpper(hex.EncodeToString(md5SumBytes[:]))
	return paths.TempDir().Join("arduino-sketch-" + md5Sum)
}

// EnsureBuildPathExists creates the build path if doesn't already exists.
func EnsureBuildPathExists(path string) error {
	if err := os.MkdirAll(path, os.FileMode(0755)); err != nil {
		return errors.Wrap(err, "unable to create build path")
	}
	return nil
}
