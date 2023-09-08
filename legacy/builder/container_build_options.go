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

type ContainerBuildOptions struct{}

func (s *ContainerBuildOptions) Run(ctx *types.Context) error {
	buildOptionsJSON, err := CreateBuildOptionsMap(
		ctx.HardwareDirs, ctx.BuiltInToolsDirs, ctx.OtherLibrariesDirs,
		ctx.BuiltInLibrariesDirs, ctx.Sketch, ctx.CustomBuildProperties,
		ctx.FQBN.String(), ctx.BuildProperties.Get("compiler.optimization_flags"),
	)
	if err != nil {
		return errors.WithStack(err)
	}
	ctx.BuildOptionsJson = buildOptionsJSON

	buildOptionsJsonPrevious, err := LoadPreviousBuildOptionsMap(ctx.BuildPath)
	if err != nil {
		return errors.WithStack(err)
	}
	ctx.BuildOptionsJsonPrevious = buildOptionsJsonPrevious

	commands := []types.Command{&WipeoutBuildPathIfBuildOptionsChanged{}}
	for _, command := range commands {
		PrintRingNameIfDebug(ctx, command)
		err := command.Run(ctx)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return StoreBuildOptionsMap(ctx.BuildPath, ctx.BuildOptionsJson)
}
