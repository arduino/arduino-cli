/*
 * This file is part of Arduino Builder.
 *
 * Arduino Builder is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2015 Arduino LLC (http://www.arduino.cc/)
 */

package builder

import (
	"os"
	"time"

	"github.com/arduino/arduino-builder/constants"
	"github.com/arduino/arduino-builder/types"
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
		logger.Fprintln(os.Stdout, logLevel, constants.MSG_LIBRARIES_MULTIPLE_LIBS_FOUND_FOR, header)
		logger.Fprintln(os.Stdout, logLevel, constants.MSG_LIBRARIES_USED, libResResult.Library.InstallDir)
		for _, notUsedLibrary := range libResResult.NotUsedLibraries {
			logger.Fprintln(os.Stdout, logLevel, constants.MSG_LIBRARIES_NOT_USED, notUsedLibrary.InstallDir)
		}
	}

	time.Sleep(100 * time.Millisecond)

	return nil
}
