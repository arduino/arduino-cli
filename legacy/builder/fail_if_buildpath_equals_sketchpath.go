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
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/i18n"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/pkg/errors"
)

type FailIfBuildPathEqualsSketchPath struct{}

func (s *FailIfBuildPathEqualsSketchPath) Run(ctx *types.Context) error {
	if ctx.BuildPath == nil || ctx.SketchLocation == nil {
		return nil
	}

	buildPath, err := ctx.BuildPath.Abs()
	if err != nil {
		return errors.WithStack(err)
	}

	sketchPath, err := ctx.SketchLocation.Abs()
	if err != nil {
		return errors.WithStack(err)
	}
	sketchPath = sketchPath.Parent()

	if buildPath.EqualsTo(sketchPath) {
		return i18n.ErrorfWithLogger(ctx.GetLogger(), constants.MSG_SKETCH_CANT_BE_IN_BUILDPATH)
	}

	return nil
}
