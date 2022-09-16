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

	"github.com/arduino/arduino-cli/legacy/builder/types"
)

type PrintUsedLibrariesIfVerbose struct{}

func (s *PrintUsedLibrariesIfVerbose) Run(ctx *types.Context) error {
	if !ctx.Verbose || len(ctx.ImportedLibraries) == 0 {
		return nil
	}

	for _, library := range ctx.ImportedLibraries {
		legacy := ""
		if library.IsLegacy {
			legacy = tr("(legacy)")
		}
		if library.Version.String() == "" {
			ctx.Info(
				tr("Using library %[1]s in folder: %[2]s %[3]s",
					library.CanonicalName,
					library.InstallDir,
					legacy))
		} else {
			ctx.Info(
				tr("Using library %[1]s at version %[2]s in folder: %[3]s %[4]s",
					library.CanonicalName,
					library.Version,
					library.InstallDir,
					legacy))
		}
	}

	time.Sleep(100 * time.Millisecond)
	return nil
}
