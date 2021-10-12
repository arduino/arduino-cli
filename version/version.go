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

package version

import (
	"github.com/arduino/arduino-cli/i18n"
)

var (
	defaultVersionString = "0.0.0-git"
	versionString        = ""
	commit               = ""
	status               = "alpha"
	date                 = ""
	tr                   = i18n.Tr
)

// Info FIXMEDOC
type Info struct {
	Application   string `json:"Application"`
	VersionString string `json:"VersionString"`
	LatestVersion string `json:"LatestVersion,omitempty"`
	Commit        string `json:"Commit"`
	Status        string `json:"Status"`
	Date          string `json:"Date"`
}

// NewInfo FIXMEDOC
func NewInfo(application string) *Info {
	return &Info{
		Application:   application,
		VersionString: versionString,
		Commit:        commit,
		Status:        status,
		Date:          date,
	}
}

func (i *Info) String() string {
	return tr("%[1]s %[2]s Version: %[3]s Commit: %[4]s Date: %[5]s", i.Application, i.Status, i.VersionString, i.Commit, i.Date)
}

//nolint:gochecknoinits
func init() {
	if versionString == "" {
		versionString = defaultVersionString
	}
}
