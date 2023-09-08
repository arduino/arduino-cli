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
	"strings"

	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
)

type WarnAboutArchIncompatibleLibraries struct{}

func (s *WarnAboutArchIncompatibleLibraries) Run(ctx *types.Context) error {
	targetPlatform := ctx.TargetPlatform
	buildProperties := ctx.BuildProperties

	archs := []string{targetPlatform.Platform.Architecture}
	if overrides, ok := buildProperties.GetOk(constants.BUILD_PROPERTIES_ARCH_OVERRIDE_CHECK); ok {
		archs = append(archs, strings.Split(overrides, ",")...)
	}

	for _, importedLibrary := range ctx.SketchLibrariesDetector.ImportedLibraries() {
		if !importedLibrary.SupportsAnyArchitectureIn(archs...) {
			ctx.Info(
				tr("WARNING: library %[1]s claims to run on %[2]s architecture(s) and may be incompatible with your current board which runs on %[3]s architecture(s).",
					importedLibrary.Name,
					strings.Join(importedLibrary.Architectures, ", "),
					strings.Join(archs, ", ")))
		}
	}

	return nil
}
