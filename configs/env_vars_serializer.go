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
	"os"

	paths "github.com/arduino/go-paths-helper"
)

// LoadFromEnv read configurations from the environment variables
func (config *Configuration) LoadFromEnv() {
	if p, has := os.LookupEnv("PROXY_TYPE"); has {
		config.ProxyType = p
	}
	if dir, has := os.LookupEnv("ARDUINO_SKETCHBOOK_DIR"); has {
		config.SketchbookDir = paths.New(dir)
	}
	if dir, has := os.LookupEnv("ARDUINO_DATA_DIR"); has {
		config.DataDir = paths.New(dir)
	}
	if dir, has := os.LookupEnv("ARDUINO_DOWNLOADS_DIR"); has {
		config.ArduinoDownloadsDir = paths.New(dir)
	}
}
