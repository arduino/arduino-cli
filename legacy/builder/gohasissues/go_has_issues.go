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

package gohasissues

import (
	"io/fs"
	"os"
	"path/filepath"
)

func ReadDir(dirname string) ([]os.FileInfo, error) {
	entries, err := os.ReadDir(dirname)
	if err != nil {
		return nil, err
	}

	infos := make([]fs.FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}

		info, err = resolveSymlink(dirname, info)
		if err != nil {
			// unresolvable symlinks should be skipped silently
			continue
		}
		infos = append(infos, info)
	}

	return infos, nil
}

func resolveSymlink(parentFolder string, info os.FileInfo) (os.FileInfo, error) {
	if !isSymlink(info) {
		return info, nil
	}
	return os.Stat(filepath.Join(parentFolder, info.Name()))
}

func isSymlink(info os.FileInfo) bool {
	return info.Mode()&os.ModeSymlink != 0
}
