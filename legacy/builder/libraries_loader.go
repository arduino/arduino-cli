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
	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesresolver"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/pkg/errors"
)

type LibrariesLoader struct{}

func (s *LibrariesLoader) Run(ctx *types.Context) error {
	if ctx.UseCachedLibrariesResolution {
		// Since we are using the cached libraries resolution
		// the library manager is not needed.
		lm := librariesmanager.NewLibraryManager(nil, nil)
		ctx.LibrariesManager = lm
	} else if ctx.LibrariesManager == nil {
		lm := librariesmanager.NewLibraryManager(nil, nil)
		ctx.LibrariesManager = lm

		builtInLibrariesFolders := ctx.BuiltInLibrariesDirs
		if builtInLibrariesFolders != nil {
			if err := builtInLibrariesFolders.ToAbs(); err != nil {
				return errors.WithStack(err)
			}
			lm.AddLibrariesDir(builtInLibrariesFolders, libraries.IDEBuiltIn)
		}

		if ctx.ActualPlatform != ctx.TargetPlatform {
			lm.AddPlatformReleaseLibrariesDir(ctx.ActualPlatform, libraries.ReferencedPlatformBuiltIn)
		}
		lm.AddPlatformReleaseLibrariesDir(ctx.TargetPlatform, libraries.PlatformBuiltIn)

		librariesFolders := ctx.OtherLibrariesDirs
		if err := librariesFolders.ToAbs(); err != nil {
			return errors.WithStack(err)
		}
		for _, folder := range librariesFolders {
			lm.AddLibrariesDir(folder, libraries.User)
		}

		if errs := lm.RescanLibraries(); len(errs) > 0 {
			// With the refactoring of the initialization step of the CLI we changed how
			// errors are returned when loading platforms and libraries, that meant returning a list of
			// errors instead of a single one to enhance the experience for the user.
			// I have no intention right now to start a refactoring of the legacy package too, so
			// here's this shitty solution for now.
			// When we're gonna refactor the legacy package this will be gone.
			return errors.WithStack(errs[0].Err())
		}

		for _, dir := range ctx.LibraryDirs {
			// Libraries specified this way have top priority
			if err := lm.LoadLibraryFromDir(dir, libraries.Unmanaged); err != nil {
				return err
			}
		}
	}

	resolver := librariesresolver.NewCppResolver()
	if err := resolver.ScanIDEBuiltinLibraries(ctx.LibrariesManager); err != nil {
		return errors.WithStack(err)
	}
	if err := resolver.ScanUserAndUnmanagedLibraries(ctx.LibrariesManager); err != nil {
		return errors.WithStack(err)
	}
	if err := resolver.ScanPlatformLibraries(ctx.LibrariesManager, ctx.TargetPlatform); err != nil {
		return errors.WithStack(err)
	}
	if ctx.ActualPlatform != ctx.TargetPlatform {
		if err := resolver.ScanPlatformLibraries(ctx.LibrariesManager, ctx.ActualPlatform); err != nil {
			return errors.WithStack(err)
		}
	}
	ctx.LibrariesResolver = resolver

	return nil
}
