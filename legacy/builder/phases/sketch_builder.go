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

package phases

import (
	"github.com/arduino/arduino-cli/legacy/builder/builder_utils"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
	"github.com/pkg/errors"
)

type SketchBuilder struct{}

func (s *SketchBuilder) Run(ctx *types.Context) error {
	sketchBuildPath := ctx.SketchBuildPath
	buildProperties := ctx.BuildProperties
	ctx.SetGlobalIncludeOption()
	includes := append(utils.Map(ctx.IncludeFolders.AsStrings(), utils.WrapWithHyphenI), ctx.GlobalIncludeOption)

	if err := sketchBuildPath.MkdirAll(); err != nil {
		return errors.WithStack(err)
	}

	objectFiles, err := builder_utils.CompileFiles(ctx, sketchBuildPath, false, sketchBuildPath, buildProperties, includes)
	if err != nil {
		return errors.WithStack(err)
	}

	// The "src/" subdirectory of a sketch is compiled recursively
	sketchSrcPath := sketchBuildPath.Join("src")
	if sketchSrcPath.IsDir() {
		srcObjectFiles, err := builder_utils.CompileFiles(ctx, sketchSrcPath, true, sketchSrcPath, buildProperties, includes)
		if err != nil {
			return errors.WithStack(err)
		}
		objectFiles.AddAll(srcObjectFiles)
	}

	ctx.SketchObjectFiles = objectFiles

	return nil
}
