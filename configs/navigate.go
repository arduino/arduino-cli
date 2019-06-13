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
	paths "github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
)

// Navigate FIXMEDOC
func (c *Configuration) Navigate(pwd *paths.Path) {
	parents := pwd.Clean().Parents()

	// From the root to the current folder, search for arduino-cli.yaml files
	for i := range parents {
		path := parents[len(parents)-i-1].Join("arduino-cli.yaml")
		logrus.Info("Checking for config in: " + path.String())
		if err := c.LoadFromYAML(path); err != nil {
			logrus.WithError(err).Infof("error loading")
		}
	}
}
