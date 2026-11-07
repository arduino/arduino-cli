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

package utils

import (
	"github.com/arduino/go-paths-helper"
)

// CopyFileAtomically copies src to a synced temp file in the dst folder, then renames it to dst.
// After a crash, dst has either the old or the new content.
func CopyFileAtomically(src, dst *paths.Path) error {
	tmp, err := paths.MkTempFile(dst.Parent(), dst.Base()+".")
	if err != nil {
		return err
	}
	tmp.Close()
	tmpPath := paths.New(tmp.Name())
	if err := src.CopyTo(tmpPath); err != nil {
		_ = tmpPath.Remove()
		return err
	}
	if err := tmpPath.Rename(dst); err != nil {
		_ = tmpPath.Remove()
		return err
	}
	return nil
}
