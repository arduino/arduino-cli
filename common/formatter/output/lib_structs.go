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

package output

import "fmt"

// VersionResult represents the output of the version commands.
type VersionResult struct {
	CommandName string `json:"command,required"`
	Version     string `json:"version,required"`
}

func (vr VersionResult) String() string {
	return fmt.Sprintf("%s version %s", vr.CommandName, vr.Version)
}
