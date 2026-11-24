// This file is part of arduino-cli.
//
// Copyright 2026 ARDUINO SA (http://www.arduino.cc/)
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

package resources

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/arduino/go-paths-helper"
)

// ValidateFileName checks that name is a plain file name (without any directory
// component) that can be safely joined to a directory without escaping it.
// File names coming from untrusted sources (package indexes, URLs, etc.) must
// be validated with this function before being used to build a filesystem path.
func ValidateFileName(name string) error {
	if name == "" || name == "." || name == ".." ||
		strings.ContainsAny(name, "/\\\x00") ||
		filepath.VolumeName(name) != "" ||
		filepath.Base(name) != name {
		return errors.New(i18n.Tr("invalid file name: %q", name))
	}
	return nil
}

// safeJoin joins the file name to dir, ensuring that the result is a direct
// child of dir.
func safeJoin(dir *paths.Path, name string) (*paths.Path, error) {
	if err := ValidateFileName(name); err != nil {
		return nil, err
	}
	res := dir.Join(name)
	if inside, err := res.IsInsideDir(dir); err != nil || !inside {
		return nil, errors.New(i18n.Tr("invalid file name: %q", name))
	}
	return res, nil
}
