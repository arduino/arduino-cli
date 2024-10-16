// This file is part of arduino-cli.
//
// Copyright 2024 ARDUINO SA (http://www.arduino.cc/)
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

package utils

import (
	"golang.org/x/sys/windows"
)

func convertAnsiBytesToString(data []byte) (string, error) {
	dataSize := int32(len(data))
	size, err := windows.MultiByteToWideChar(windows.GetACP(), 0, &data[0], dataSize, nil, 0)
	if err != nil {
		return "", err
	}
	utf16 := make([]uint16, size)
	if _, err := windows.MultiByteToWideChar(windows.GetACP(), 0, &data[0], dataSize, &utf16[0], size); err != nil {
		return "", err
	}
	return windows.UTF16ToString(utf16), nil
}
