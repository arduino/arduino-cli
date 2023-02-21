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
	properties "github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
)

type SetupBuildProperties struct{}

func (s *SetupBuildProperties) Run(ctx *types.Context) error {
	targetPlatform := ctx.TargetPlatform
	actualPlatform := ctx.ActualPlatform

	buildProperties := properties.NewMap()
	buildProperties.Merge(actualPlatform.Properties)
	buildProperties.Merge(targetPlatform.Properties)
	buildProperties.Merge(ctx.TargetBoardBuildProperties)

	if ctx.BuildPath != nil {
		buildProperties.SetPath("build.path", ctx.BuildPath)
	}
	if ctx.Sketch != nil {
		buildProperties.Set("build.project_name", ctx.Sketch.MainFile.Base())
	}

	if ctx.OptimizeForDebug {
		if buildProperties.ContainsKey("compiler.optimization_flags.debug") {
			buildProperties.Set("compiler.optimization_flags", buildProperties.Get("compiler.optimization_flags.debug"))
		}
	} else {
		if buildProperties.ContainsKey("compiler.optimization_flags.release") {
			buildProperties.Set("compiler.optimization_flags", buildProperties.Get("compiler.optimization_flags.release"))
		}
	}
	ctx.OptimizationFlags = buildProperties.Get("compiler.optimization_flags")

	buildProperties.SetPath("build.source.path", ctx.Sketch.FullPath)

	buildProperties.Merge(ctx.PackageManager.GetCustomGlobalProperties())

	keychainProp := buildProperties.ContainsKey("build.keys.keychain")
	signProp := buildProperties.ContainsKey("build.keys.sign_key")
	encryptProp := buildProperties.ContainsKey("build.keys.encrypt_key")
	// we verify that all the properties for the secure boot keys are defined or none of them is defined.
	if (keychainProp || signProp || encryptProp) && !(keychainProp && signProp && encryptProp) {
		return errors.Errorf("%s platform does not specify correctly default sign and encryption keys", targetPlatform.Platform)
	}

	ctx.BuildProperties = buildProperties

	return nil
}
