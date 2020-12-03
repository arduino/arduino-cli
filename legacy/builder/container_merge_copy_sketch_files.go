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
	bldr "github.com/arduino/arduino-cli/arduino/builder"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/pkg/errors"
)

type ContainerMergeCopySketchFiles struct{}

func (s *ContainerMergeCopySketchFiles) Run(ctx *types.Context) error {
	sk := types.SketchFromLegacy(ctx.Sketch)
	if sk == nil {
		return errors.New("unable to convert legacy sketch to the new type")
	}
	offset, source, err := bldr.SketchMergeSources(sk, ctx.SourceOverride)
	if err != nil {
		return err
	}
	ctx.LineOffset = offset
	ctx.Source = source

	if err := bldr.SketchSaveItemCpp(ctx.Sketch.MainFile.Name.String(), []byte(ctx.Source), ctx.SketchBuildPath.String()); err != nil {
		return errors.WithStack(err)
	}

	if err := bldr.SketchCopyAdditionalFiles(sk, ctx.SketchBuildPath.String(), ctx.SourceOverride); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
