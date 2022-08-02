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
	sk "github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/pkg/errors"
)

type ContainerSetupHardwareToolsLibsSketchAndProps struct{}

func (s *ContainerSetupHardwareToolsLibsSketchAndProps) Run(ctx *types.Context) error {
	// total number of steps in this container: 14
	ctx.Progress.AddSubSteps(14)
	defer ctx.Progress.RemoveSubSteps()

	commands := []types.Command{
		&AddAdditionalEntriesToContext{},
		&FailIfBuildPathEqualsSketchPath{},
		&HardwareLoader{},
		&PlatformKeysRewriteLoader{},
		&RewriteHardwareKeys{},
		&TargetBoardResolver{},
		&ToolsLoader{},
		&LibrariesLoader{},
	}

	for _, command := range commands {
		PrintRingNameIfDebug(ctx, command)
		err := command.Run(ctx)
		if err != nil {
			return errors.WithStack(err)
		}
		ctx.Progress.CompleteStep()
		ctx.PushProgress()
	}

	if ctx.SketchLocation != nil {
		// get abs path to sketch
		sketchLocation, err := ctx.SketchLocation.Abs()
		if err != nil {
			return errors.WithStack(err)
		}

		// load sketch
		sketch, err := sk.New(sketchLocation)
		if err != nil {
			return errors.WithStack(err)
		}
		sketch.BuildPath = ctx.BuildPath
		ctx.SketchLocation = sketch.MainFile
		ctx.Sketch = sketch
	}
	ctx.Progress.CompleteStep()
	ctx.PushProgress()

	commands = []types.Command{
		&SetupBuildProperties{},
		&LoadVIDPIDSpecificProperties{},
		&SetCustomBuildProperties{},
		&AddMissingBuildPropertiesFromParentPlatformTxtFiles{},
	}

	for _, command := range commands {
		PrintRingNameIfDebug(ctx, command)
		err := command.Run(ctx)
		if err != nil {
			return errors.WithStack(err)
		}
		ctx.Progress.CompleteStep()
		ctx.PushProgress()
	}

	return nil
}
