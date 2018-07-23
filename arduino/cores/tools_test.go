/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2018 ARDUINO AG (http://www.arduino.cc/)
 */

package cores

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlavorCompatibility(t *testing.T) {
	type os struct {
		Os   string
		Arch string
	}
	windowsi386 := &os{"windows", "i386"}
	windowsx8664 := &os{"windows", "amd64"}
	linuxi386 := &os{"linux", "i386"}
	linuxamd64 := &os{"linux", "amd64"}
	linuxarm := &os{"linux", "arm"}
	linuxarmbe := &os{"linux", "armbe"}
	darwini386 := &os{"darwin", "i386"}
	darwinamd646 := &os{"darwin", "amd64"}
	freebsdi386 := &os{"freebsd", "i386"}
	freebsdamd64 := &os{"freebsd", "amd64"}
	oses := []*os{
		windowsi386,
		windowsx8664,
		linuxi386,
		linuxamd64,
		linuxarm,
		linuxarmbe,
		darwini386,
		darwinamd646,
		freebsdi386,
		freebsdamd64,
	}

	type test struct {
		Flavour   *Flavour
		Positives []*os
	}
	tests := []*test{
		&test{&Flavour{OS: "i686-mingw32"}, []*os{windowsi386, windowsx8664}},
		&test{&Flavour{OS: "i386-apple-darwin11"}, []*os{darwini386, darwinamd646}},
		&test{&Flavour{OS: "x86_64-apple-darwin"}, []*os{darwinamd646}},

		// Raspberry PI, BBB or other ARM based host
		// PI: "arm-linux-gnueabihf"
		// Raspbian on PI2: "arm-linux-gnueabihf"
		// Ubuntu Mate on PI2: "arm-linux-gnueabihf"
		// Debian 7.9 on BBB: "arm-linux-gnueabihf"
		// Raspbian on PI Zero: "arm-linux-gnueabihf"
		&test{&Flavour{OS: "arm-linux-gnueabihf"}, []*os{linuxarm, linuxarmbe}},
		// Arch-linux on PI2: "armv7l-unknown-linux-gnueabihf"
		&test{&Flavour{OS: "armv7l-unknown-linux-gnueabihf"}, []*os{linuxarm, linuxarmbe}},

		&test{&Flavour{OS: "i686-linux-gnu"}, []*os{linuxi386}},
		&test{&Flavour{OS: "i686-pc-linux-gnu"}, []*os{linuxi386}},
		&test{&Flavour{OS: "x86_64-linux-gnu"}, []*os{linuxamd64}},
		&test{&Flavour{OS: "x86_64-pc-linux-gnu"}, []*os{linuxamd64}},
	}

	check := func(test *test, os *os) {
		for _, positiveOs := range test.Positives {
			if positiveOs == os {
				require.True(t, test.Flavour.isCompatibleWith(os.Os, os.Arch), "'%s' tag compatible with '%s,%s' pair", test.Flavour.OS, os.Os, os.Arch)
				return
			}
		}
		require.False(t, test.Flavour.isCompatibleWith(os.Os, os.Arch), "'%s' tag compatible with '%s,%s' pair", test.Flavour.OS, os.Os, os.Arch)
	}

	for _, test := range tests {
		for _, os := range oses {
			check(test, os)
		}
	}
}
