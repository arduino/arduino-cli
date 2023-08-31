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
	"bytes"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesresolver"
	"github.com/arduino/go-paths-helper"
	"github.com/pkg/errors"
)

func LibrariesLoader(
	useCachedLibrariesResolution bool,
	librariesManager *librariesmanager.LibrariesManager,
	builtInLibrariesDirs *paths.Path, libraryDirs, otherLibrariesDirs paths.PathList,
	actualPlatform, targetPlatform *cores.PlatformRelease,
) (*librariesmanager.LibrariesManager, *librariesresolver.Cpp, []byte, error) {
	verboseOut := &bytes.Buffer{}
	lm := librariesManager
	if useCachedLibrariesResolution {
		// Since we are using the cached libraries resolution
		// the library manager is not needed.
		lm = librariesmanager.NewLibraryManager(nil, nil)
	}
	if librariesManager == nil {
		lm = librariesmanager.NewLibraryManager(nil, nil)

		builtInLibrariesFolders := builtInLibrariesDirs
		if builtInLibrariesFolders != nil {
			if err := builtInLibrariesFolders.ToAbs(); err != nil {
				return nil, nil, nil, errors.WithStack(err)
			}
			lm.AddLibrariesDir(builtInLibrariesFolders, libraries.IDEBuiltIn)
		}

		if actualPlatform != targetPlatform {
			lm.AddPlatformReleaseLibrariesDir(actualPlatform, libraries.ReferencedPlatformBuiltIn)
		}
		lm.AddPlatformReleaseLibrariesDir(targetPlatform, libraries.PlatformBuiltIn)

		librariesFolders := otherLibrariesDirs
		if err := librariesFolders.ToAbs(); err != nil {
			return nil, nil, nil, errors.WithStack(err)
		}
		for _, folder := range librariesFolders {
			lm.AddLibrariesDir(folder, libraries.User)
		}

		for _, status := range lm.RescanLibraries() {
			// With the refactoring of the initialization step of the CLI we changed how
			// errors are returned when loading platforms and libraries, that meant returning a list of
			// errors instead of a single one to enhance the experience for the user.
			// I have no intention right now to start a refactoring of the legacy package too, so
			// here's this shitty solution for now.
			// When we're gonna refactor the legacy package this will be gone.
			verboseOut.Write([]byte(status.Message()))
		}

		for _, dir := range libraryDirs {
			// Libraries specified this way have top priority
			if err := lm.LoadLibraryFromDir(dir, libraries.Unmanaged); err != nil {
				return nil, nil, nil, errors.WithStack(err)
			}
		}
	}

	resolver := librariesresolver.NewCppResolver()
	if err := resolver.ScanIDEBuiltinLibraries(lm); err != nil {
		return nil, nil, nil, errors.WithStack(err)
	}
	if err := resolver.ScanUserAndUnmanagedLibraries(lm); err != nil {
		return nil, nil, nil, errors.WithStack(err)
	}
	if err := resolver.ScanPlatformLibraries(lm, targetPlatform); err != nil {
		return nil, nil, nil, errors.WithStack(err)
	}
	if actualPlatform != targetPlatform {
		if err := resolver.ScanPlatformLibraries(lm, actualPlatform); err != nil {
			return nil, nil, nil, errors.WithStack(err)
		}
	}
	return lm, resolver, verboseOut.Bytes(), nil
}
