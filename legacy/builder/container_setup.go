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
	"fmt"

	bldr "github.com/arduino/arduino-cli/arduino/builder"
	sk "github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/arduino-cli/legacy/builder/builder_utils"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/go-paths-helper"
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
		&AddBuildBoardPropertyIfMissing{},
		&LibrariesLoader{},
	}

	for _, command := range commands {
		PrintRingNameIfDebug(ctx, command)
		err := command.Run(ctx)
		if err != nil {
			return errors.WithStack(err)
		}
		ctx.Progress.CompleteStep()
		builder_utils.PrintProgressIfProgressEnabledAndMachineLogger(ctx)
	}

	if ctx.SketchLocation != nil {
		// get abs path to sketch
		sketchLocation, err := ctx.SketchLocation.Abs()
		if err != nil {
			return errors.WithStack(err)
		}

		// load sketch
		sketch, err := bldr.SketchLoad(sketchLocation.String(), ctx.BuildPath.String())
		if e, ok := err.(*sk.InvalidSketchFoldernameError); ctx.IgnoreSketchFolderNameErrors && ok {
			// ignore error
			sketch = e.Sketch
		} else if err != nil {
			return errors.WithStack(err)
		}
		if sketch.MainFile == nil {
			return fmt.Errorf("main file missing from sketch")
		}
		ctx.SketchLocation = paths.New(sketch.MainFile.Path)
		ctx.Sketch = types.SketchToLegacy(sketch)
	}
	ctx.Progress.CompleteStep()
	builder_utils.PrintProgressIfProgressEnabledAndMachineLogger(ctx)

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
		builder_utils.PrintProgressIfProgressEnabledAndMachineLogger(ctx)
	}

	return nil
}
