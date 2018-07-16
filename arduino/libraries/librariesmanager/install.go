/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO AG (http://www.arduino.cc/)
 *
 * arduino-cli is free software; you can redistribute it and/or modify
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
 */

package librariesmanager

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bcmi-labs/arduino-cli/arduino/libraries"
	"github.com/bcmi-labs/arduino-cli/arduino/libraries/librariesindex"
	"github.com/bcmi-labs/arduino-cli/arduino/utils"

	"github.com/bcmi-labs/arduino-cli/configs"
)

// Install installs a library and returns the installed path.
func Install(library *librariesindex.Release) (string, error) {
	if library == nil {
		return "", errors.New("Not existing version of the library")
	}

	/*
		installedRelease, err := library.InstalledRelease()
		if err != nil {
			return err
		}
		if installedRelease != nil {
			//if installedRelease.Version != library.Latest().Version {
			err := removeRelease(library.Name, installedRelease)
			if err != nil {
				return err
			}
			//} else {
			//	return nil // Already installed and latest version.
			//}
		}
	*/

	libsFolder, err := configs.LibrariesFolder.Get()
	if err != nil {
		return "", fmt.Errorf("getting libraries directory: %s", err)
	}

	libPath := filepath.Join(libsFolder, utils.SanitizeName(library.Library.Name))
	return libPath, library.Resource.Install(libsFolder, libPath)
}

func removeRelease(libName string, r *libraries.Library) error {
	libFolder, err := configs.LibrariesFolder.Get()
	if err != nil {
		return err
	}

	libName = utils.SanitizeName(libName)
	path := filepath.Join(libFolder, libName)
	return os.RemoveAll(path)
}
