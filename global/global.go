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

package global

import (
	"os"
	"path/filepath"
)

// appName is the command line name of the Arduino CLI executable on the user system (users may change it)
var appName = filepath.Base(os.Args[0])

var (
	application = "arduino-cli"
	version     = "0.3.6-alpha.preview"
	commit      = "missing"
	cvsRef      = "missing"
	buildDate   = "missing"
	repository  = "missing"
)

func GetAppName() string {
	return appName
}

func GetApplication() string {
	return application
}

func GetVersion() string {
	return version
}

func GetCommit() string {
	return commit
}

func GetCvsRef() string {
	return cvsRef
}

func GetBuildDate() string {
	return buildDate
}

func GetRepository() string {
	return repository
}

type Info struct {
	Application string `json:"application"`
	Version     string `json:"version"`
	Commit      string `json:"commit"`
	CvsRef      string `json:"cvsRef"`
	BuildDate   string `json:"buildDate"`
	Repository  string `json:"repository"`
}
