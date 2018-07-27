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

	"github.com/bcmi-labs/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/bcmi-labs/arduino-cli/arduino/libraries/librariesresolver"

	"github.com/arduino/arduino-builder/i18n"
	"github.com/arduino/arduino-builder/types"
	"github.com/bcmi-labs/arduino-cli/arduino/libraries"
)

type LibrariesLoader struct{}

func (s *LibrariesLoader) Run(ctx *types.Context) error {
	lm := librariesmanager.NewLibraryManager(nil, nil)
	ctx.LibrariesManager = lm

	builtInLibrariesFolders := ctx.BuiltInLibrariesFolders
	if err := builtInLibrariesFolders.ToAbs(); err != nil {
		return i18n.WrapError(err)
	}
	for _, folder := range builtInLibrariesFolders {
		lm.AddLibrariesDir(folder, libraries.IDEBuiltIn)
	}

	debugLevel := ctx.DebugLevel
	logger := ctx.GetLogger()

	actualPlatform := ctx.ActualPlatform
	platform := ctx.TargetPlatform
	if actualPlatform != platform {
		lm.AddPlatformReleaseLibrariesDir(actualPlatform, libraries.ReferencedPlatformBuiltIn)
	}
	lm.AddPlatformReleaseLibrariesDir(platform, libraries.PlatformBuiltIn)

	librariesFolders := ctx.OtherLibrariesFolders
	if err := librariesFolders.ToAbs(); err != nil {
		return i18n.WrapError(err)
	}
	for _, folder := range librariesFolders {
		lm.AddLibrariesDir(folder, libraries.Sketchbook)
	}

	if err := lm.RescanLibraries(); err != nil {
		return i18n.WrapError(err)
	}

	if debugLevel > 0 {
		for _, lib := range lm.Libraries {
			for _, libAlt := range lib.Alternatives {
				warnings, err := libAlt.Lint()
				if err != nil {
					return i18n.WrapError(err)
				}
				for _, warning := range warnings {
					logger.Fprintln(os.Stdout, "warn", warning)
				}
			}
		}
	}

	resolver := librariesresolver.NewCppResolver()
	if err := resolver.ScanFromLibrariesManager(lm); err != nil {
		return i18n.WrapError(err)
	}
	ctx.LibrariesResolver = resolver

	return nil
}
