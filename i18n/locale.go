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

package i18n

//go:generate ./embed-i18n.sh

import (
	"sync"

	rice "github.com/GeertJohan/go.rice"
	"github.com/leonelquinteros/gotext"
)

var (
	loadOnce sync.Once
	po       *gotext.Po
)

func init() {
	po = new(gotext.Po)
}

// SetLocale sets the locate used for i18n
func SetLocale(locale string) {
	box := rice.MustFindBox("./data")
	poFile, err := box.Bytes(locale + ".po")

	if err != nil {
		poFile = box.MustBytes("en.po")
	}

	po = new(gotext.Po)
	po.Parse(poFile)
}
