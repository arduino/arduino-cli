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

import (
	"strings"
	"syscall"
	"unsafe"

	"github.com/sirupsen/logrus"
)

func getLocaleIdentifier() string {
	defer func() {
		if r := recover(); r != nil {
			logrus.WithField("error", r).Errorf("Failed to get windows user locale")
		}
	}()

	dll := syscall.MustLoadDLL("kernel32")
	defer dll.Release()
	proc := dll.MustFindProc("GetUserDefaultLocaleName")

	localeNameMaxLen := 85
	buffer := make([]uint16, localeNameMaxLen)
	len, _, err := proc.Call(uintptr(unsafe.Pointer(&buffer[0])), uintptr(localeNameMaxLen))

	if len == 0 {
		panic(err)
	}

	locale := syscall.UTF16ToString(buffer)

	return strings.ReplaceAll(locale, "-", "_")
}
