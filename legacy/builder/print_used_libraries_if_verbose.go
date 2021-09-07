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
	"time"

	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
)

type PrintUsedLibrariesIfVerbose struct{}

func (s *PrintUsedLibrariesIfVerbose) Run(ctx *types.Context) error {
	verbose := ctx.Verbose
	logger := ctx.GetLogger()

	if !verbose || len(ctx.ImportedLibraries) == 0 {
		return nil
	}

	for _, library := range ctx.ImportedLibraries {
		legacy := ""
		if library.IsLegacy {
			legacy = tr("(legacy)")
		}
		if library.Version.String() == "" {
			logger.Println(constants.LOG_LEVEL_INFO,
				tr("Using library {0} in folder: {1} {2}"),
				library.Name,
				library.InstallDir,
				legacy)
		} else {
			logger.Println(constants.LOG_LEVEL_INFO,
				tr("Using library {0} at version {1} in folder: {2} {3}"),
				library.Name,
				library.Version,
				library.InstallDir,
				legacy)
		}
	}

	time.Sleep(100 * time.Millisecond)

	return nil
}
