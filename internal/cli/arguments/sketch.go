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

package arguments

import (
	"fmt"

	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
)

// InitSketchPath returns an instance of paths.Path pointing to sketchPath.
// If sketchPath is an empty string returns the current working directory.
// In both cases it warns the user if he's using deprecated files
func InitSketchPath(path string) (sketchPath *paths.Path) {
	if path != "" {
		sketchPath = paths.New(path)
	} else {
		wd, err := paths.Getwd()
		if err != nil {
			feedback.Fatal(tr("Couldn't get current working directory: %v", err), feedback.ErrGeneric)
		}
		logrus.Infof("Reading sketch from dir: %s", wd)
		sketchPath = wd
	}
	WarnDeprecatedFiles(sketchPath)
	return sketchPath
}

// NewSketch is a helper function useful to create a sketch instance
func NewSketch(sketchPath *paths.Path) (*sketch.Sketch, error) {
	return sketch.New(sketchPath)
}

// MustNewSketch is a helper function useful to create a sketch instance, exits if the
// initialization fails
func MustNewSketch(sketchPath *paths.Path) *sketch.Sketch {
	sketch, err := sketch.New(sketchPath)
	if err != nil {
		feedback.Fatal(tr("Error opening sketch: %v", err), feedback.ErrGeneric)
	}
	return sketch
}

// WarnDeprecatedFiles warns the user that a type of sketch files are deprecated
func WarnDeprecatedFiles(sketchPath *paths.Path) {
	// .pde files are still supported but deprecated, this warning urges the user to rename them
	if files := sketch.CheckForPdeFiles(sketchPath); len(files) > 0 {
		msg := tr("Sketches with .pde extension are deprecated, please rename the following files to .ino:")
		for _, f := range files {
			msg += fmt.Sprintf("\n - %s", f)
		}
		feedback.Warning(msg)
	}
}
