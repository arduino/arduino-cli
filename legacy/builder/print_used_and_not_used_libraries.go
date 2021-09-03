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
	"time"

	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
)

type PrintUsedAndNotUsedLibraries struct {
	// Was there an error while compiling the sketch?
	SketchError bool
}

func (s *PrintUsedAndNotUsedLibraries) Run(ctx *types.Context) error {
	var logLevel string
	// Print this message as warning when the sketch didn't compile,
	// as info when we're verbose and not all otherwise
	if s.SketchError {
		logLevel = constants.LOG_LEVEL_WARN
	} else if ctx.Verbose {
		logLevel = constants.LOG_LEVEL_INFO
	} else {
		return nil
	}

	logger := ctx.GetLogger()
	libraryResolutionResults := ctx.LibrariesResolutionResults

	for header, libResResult := range libraryResolutionResults {
		if len(libResResult.NotUsedLibraries) == 0 {
			continue
		}
		logger.Fprintln(os.Stdout, logLevel, tr("Multiple libraries were found for \"{0}\""), header)
		logger.Fprintln(os.Stdout, logLevel, " "+tr("Used: {0}"), libResResult.Library.InstallDir)
		for _, notUsedLibrary := range libResResult.NotUsedLibraries {
			logger.Fprintln(os.Stdout, logLevel, " "+tr("Not used: {0}"), notUsedLibrary.InstallDir)
		}
	}

	time.Sleep(100 * time.Millisecond)

	return nil
}
