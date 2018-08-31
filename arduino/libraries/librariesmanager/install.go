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
	"fmt"

	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/arduino/utils"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/go-paths-helper"
)

// Install installs a library and returns the installed path.
func (lm *LibrariesManager) Install(indexLibrary *librariesindex.Release) (*paths.Path, error) {
	var replaced *libraries.Library
	if installedLibs, have := lm.Libraries[indexLibrary.Library.Name]; have {
		for _, installedLib := range installedLibs.Alternatives {
			if installedLib.Location != libraries.Sketchbook {
				continue
			}
			if installedLib.Version.Equal(indexLibrary.Version) {
				return installedLib.InstallDir, fmt.Errorf("%s is already installed", indexLibrary.String())
			}
			replaced = installedLib
		}
	}

	libsDir := lm.getSketchbookLibrariesDir()
	if libsDir == nil {
		return nil, fmt.Errorf("sketchbook directory not set")
	}

	libPath := libsDir.Join(utils.SanitizeName(indexLibrary.Library.Name))
	if replaced != nil && replaced.InstallDir.EquivalentTo(libPath) {
		formatter.Print(fmt.Sprintf("Replacing %s with %s", replaced, indexLibrary))
	} else if libPath.IsDir() {
		return nil, fmt.Errorf("destination dir %s already exists, cannot install", libPath)
	}
	return libPath, indexLibrary.Resource.Install(lm.DownloadsDir, libsDir, libPath)
}

// Uninstall removes a Library
func (lm *LibrariesManager) Uninstall(lib *libraries.Library) error {
	if err := lib.InstallDir.RemoveAll(); err != nil {
		return fmt.Errorf("removing lib directory: %s", err)
	}

	lm.Libraries[lib.Name].Remove(lib)
	return nil
}
