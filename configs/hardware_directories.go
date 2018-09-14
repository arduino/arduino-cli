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

package configs

import (
	"github.com/arduino/go-paths-helper"
)

// HardwareDirectories returns all paths that may contains hardware packages.
func (config *Configuration) HardwareDirectories() (paths.PathList, error) {
	res := paths.PathList{}

	if config.IsBundledInDesktopIDE() {
		bundledHardwareDir := config.ArduinoIDEDirectory.Join("hardware")
		if bundledHardwareDir.IsDir() {
			res.Add(bundledHardwareDir)
		}
	}

	if dir := config.PackagesDir(); dir.IsDir() {
		res.Add(dir)
	}

	if config.SketchbookDir != nil {
		if dir := config.SketchbookDir.Join("hardware"); dir.IsDir() {
			res.Add(dir)
		}
	}

	return res, nil
}

// BundleToolsDirectories returns all paths that may contains bundled-tools.
func (config *Configuration) BundleToolsDirectories() (paths.PathList, error) {
	res := paths.PathList{}

	if config.IsBundledInDesktopIDE() {
		bundledToolsDir := config.ArduinoIDEDirectory.Join("hardware", "tools")
		if bundledToolsDir.IsDir() {
			res = append(res, bundledToolsDir)
		}
	}

	return res, nil
}
