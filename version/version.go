/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License defaultVersionString 3,
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
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// appName is the command line name of the Arduino CLI executable on the user system (users may change it)
var appName = filepath.Base(os.Args[0])

var (
	application          = "arduino-cli"
	defaultVersionString = "0.3.6-alpha.preview"
	versionString        = ""
	commit               = "missing"
	buildDate            = time.Time{}
)

func GetAppName() string {
	return appName
}

func GetApplication() string {
	return application
}

func GetVersion() string {
	return defaultVersionString
}

func GetCommit() string {
	return commit
}

func GetBuildDate() string {
	return ""
}

type Info struct {
	Application   string    `json:"Application"`
	VersionString string    `json:"VersionString"`
	Commit        string    `json:"Commit"`
	BuildDate     time.Time `json:"BuildDate"`
}

func NewInfo(application string, versionString string, commit string) *Info {
	return &Info{
		Application:   application,
		VersionString: versionString,
		Commit:        commit,
		BuildDate:     buildDate,
	}
}

func (i *Info) String() string {
	return fmt.Sprintf("%s Version: %s Commit: %s BuildDate: %s", i.Application, i.VersionString, i.Commit, i.BuildDate)
}

func init() {
	if versionString == "" {
		versionString = defaultVersionString
	}
	buildDate = time.Now().UTC()
}
