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
	"path/filepath"
	"strings"

	paths "github.com/arduino/go-paths-helper"
)

func (c *Configuration) Navigate(root, pwd string) {
	relativePath, err := filepath.Rel(root, pwd)
	if err != nil {
		return
	}

	// From the root to the current folder, search for arduino-cli.yaml files
	parts := strings.Split(relativePath, string(filepath.Separator))
	for i := range parts {
		path := paths.New(root)
		path = path.Join(parts[:i+1]...)
		path = path.Join("arduino-cli.yaml")
		_ = c.LoadFromYAML(path)
	}
}
