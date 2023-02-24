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
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/pkg/errors"
)

type FailIfBuildPathEqualsSketchPath struct{}

func (s *FailIfBuildPathEqualsSketchPath) Run(ctx *types.Context) error {
	if ctx.BuildPath == nil || ctx.Sketch == nil {
		return nil
	}
	buildPath := ctx.BuildPath.Canonical()
	sketchPath := ctx.Sketch.FullPath.Canonical()
	if buildPath.EqualsTo(sketchPath) {
		return errors.New(tr("Sketch cannot be located in build path. Please specify a different build path"))
	}

	return nil
}
