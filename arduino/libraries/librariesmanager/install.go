/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package librariesmanager

import (
	"errors"
	"fmt"

	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/arduino/utils"
	paths "github.com/arduino/go-paths-helper"
)

var (
	// ErrAlreadyInstalled is returned when a library is already installed and task
	// cannot proceed.
	ErrAlreadyInstalled = errors.New("library already installed")
)

// InstallPrerequisiteCheck performs prequisite checks to install a library. It returns the
// install path, where the library should be installed and the possible library that is already
// installed on the same folder and it's going to be replaced by the new one.
func (lm *LibrariesManager) InstallPrerequisiteCheck(indexLibrary *librariesindex.Release) (*paths.Path, *libraries.Library, error) {
	saneName := utils.SanitizeName(indexLibrary.Library.Name)

	var replaced *libraries.Library
	if installedLibs, have := lm.Libraries[saneName]; have {
		for _, installedLib := range installedLibs.Alternatives {
			if installedLib.Location != libraries.Sketchbook {
				continue
			}
			if installedLib.Version.Equal(indexLibrary.Version) {
				return installedLib.InstallDir, nil, ErrAlreadyInstalled
			}
			replaced = installedLib
		}
	}

	libsDir := lm.getSketchbookLibrariesDir()
	if libsDir == nil {
		return nil, nil, fmt.Errorf("sketchbook directory not set")
	}

	libPath := libsDir.Join(saneName)
	if replaced != nil && replaced.InstallDir.EquivalentTo(libPath) {

	} else if libPath.IsDir() {
		return nil, nil, fmt.Errorf("destination dir %s already exists, cannot install", libPath)
	}
	return libPath, replaced, nil
}

// Install installs a library on the specified path.
func (lm *LibrariesManager) Install(indexLibrary *librariesindex.Release, libPath *paths.Path) error {
	libsDir := lm.getSketchbookLibrariesDir()
	if libsDir == nil {
		return fmt.Errorf("sketchbook directory not set")
	}
	return indexLibrary.Resource.Install(lm.DownloadsDir, libsDir, libPath)
}

// Uninstall removes a Library
func (lm *LibrariesManager) Uninstall(lib *libraries.Library) error {
	if err := lib.InstallDir.RemoveAll(); err != nil {
		return fmt.Errorf("removing lib directory: %s", err)
	}

	lm.Libraries[lib.Name].Remove(lib)
	return nil
}
