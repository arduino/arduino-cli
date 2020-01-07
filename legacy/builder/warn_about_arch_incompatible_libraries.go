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
	"os"
	"strings"

	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
)

type WarnAboutArchIncompatibleLibraries struct{}

func (s *WarnAboutArchIncompatibleLibraries) Run(ctx *types.Context) error {
	if ctx.DebugLevel < 0 {
		return nil
	}

	targetPlatform := ctx.TargetPlatform
	buildProperties := ctx.BuildProperties
	logger := ctx.GetLogger()

	archs := []string{targetPlatform.Platform.Architecture}
	if overrides, ok := buildProperties.GetOk(constants.BUILD_PROPERTIES_ARCH_OVERRIDE_CHECK); ok {
		archs = append(archs, strings.Split(overrides, ",")...)
	}

	for _, importedLibrary := range ctx.ImportedLibraries {
		if !importedLibrary.SupportsAnyArchitectureIn(archs...) {
			logger.Fprintln(os.Stdout, constants.LOG_LEVEL_WARN, constants.MSG_LIBRARY_INCOMPATIBLE_ARCH,
				importedLibrary.Name,
				strings.Join(importedLibrary.Architectures, ", "),
				strings.Join(archs, ", "))
		}
	}

	return nil
}
