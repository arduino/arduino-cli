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
	"time"

	"github.com/arduino/arduino-cli/legacy/builder/types"
)

type PrintUsedAndNotUsedLibraries struct {
	// Was there an error while compiling the sketch?
	SketchError bool
}

func (s *PrintUsedAndNotUsedLibraries) Run(ctx *types.Context) error {
	// Print this message:
	// - as warning, when the sketch didn't compile
	// - as info, when verbose is on
	// - otherwise, output nothing
	if !s.SketchError && !ctx.Verbose {
		return nil
	}

	res := ""
	for header, libResResult := range ctx.LibrariesResolutionResults {
		if len(libResResult.NotUsedLibraries) == 0 {
			continue
		}
		res += fmt.Sprintln(tr(`Multiple libraries were found for "%[1]s"`, header))
		res += fmt.Sprintln("  " + tr("Used: %[1]s", libResResult.Library.InstallDir))
		for _, notUsedLibrary := range libResResult.NotUsedLibraries {
			res += fmt.Sprintln("  " + tr("Not used: %[1]s", notUsedLibrary.InstallDir))
		}
	}

	if s.SketchError {
		ctx.Warn(res)
	} else {
		ctx.Info(res)
	}
	time.Sleep(100 * time.Millisecond)
	return nil
}
