/*
 * This file is part of arduino-cli.
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
 *
 * Copyright 2017 BCMI LABS SA (http://www.arduino.cc/)
 */

package libraries

import (
	"archive/zip"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"strings"

	"github.com/bcmi-labs/arduino-cli/common"
	"github.com/bcmi-labs/arduino-cli/common/releases"
)

// Uninstall a library means remove its directory.
var Uninstall = os.RemoveAll

// InstallLib installs a library.
func InstallLib(name string, release releases.Release) error {
	if release == nil {
		return errors.New("Not existing version of the library")
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
	libFolder, err := common.GetDefaultLibFolder()
	if err != nil {
		return err
	}

	stagingFolder, err := common.GetDownloadCacheFolder()
	if err != nil {
		return err
	}

	cacheFilePath := filepath.Join(stagingFolder, release.ArchiveName())

	zipArchive, err := zip.OpenReader(cacheFilePath)
	if err != nil {
		return err
	}
	defer zipArchive.Close()

	err = common.Unzip(zipArchive, libFolder)
	if err != nil {
		return err
	}

	return nil
}

func removeRelease(libName string, r *Release) error {
	libFolder, err := common.GetDefaultLibFolder()
	if err != nil {
		return err
	}

	libName = strings.Replace(libName, " ", "_", -1)

	path := filepath.Join(libFolder, fmt.Sprintf("%s-%s", libName, r.Version))
	return Uninstall(path)
}
