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

package utils

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestURLParse(t *testing.T) {
	type test struct {
		URL          string
		ExpectedHost string
		ExpectedPath string
		Skip         bool
	}
	onWindows := runtime.GOOS == "windows"
	tests := []test{
		{"https://example.com", "example.com", "", false},
		{"https://example.com/some/path", "example.com", "/some/path", false},
		{"file:///home/user/some/path", "", "/home/user/some/path", onWindows},
		{"file:///C:/Users/me/some/path", "", "C:/Users/me/some/path", !onWindows},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("URLParseTest%02d", i), func(t *testing.T) {
			if test.Skip {
				t.Skip("Skipped")
			}
			res, err := URLParse(test.URL)
			require.NoError(t, err)
			require.Equal(t, test.ExpectedPath, res.Path)
		})
	}
}
