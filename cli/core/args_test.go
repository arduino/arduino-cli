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

package core

import (
	"testing"

	"github.com/stretchr/testify/require"
	semver "go.bug.st/relaxed-semver"
)

func TestParsePlatformReferenceArgs(t *testing.T) {
	valid := func(arg, pack, arch, ver string) {
		version, _ := semver.Parse(ver) // use nil in case of error

		ref, err := parsePlatformReferenceArg(arg)
		require.NoError(t, err)
		require.Equal(t, pack, ref.Package)
		require.Equal(t, arch, ref.Architecture)
		require.Equal(t, version.String(), ref.Version)
	}
	invalid := func(arg string) {
		_, err := parsePlatformReferenceArg(arg)
		require.Error(t, err)
	}
	valid("arduino:avr", "arduino", "avr", "-")
	valid("arduino:avr@1.6.20", "arduino", "avr", "1.6.20")
	valid("arduino:avr@", "arduino", "avr", "")
	invalid("avr")
	invalid("arduino:avr:avr")
	invalid("arduino@1.6.20:avr")
	invalid("avr@1.6.20")
	invalid("arduino:avr:avr@1.6.20")
}
