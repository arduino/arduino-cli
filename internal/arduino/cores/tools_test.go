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

package cores

import (
	"testing"

	"github.com/arduino/arduino-cli/internal/arduino/resources"
	"github.com/stretchr/testify/require"
)

func TestFlavorCompatibility(t *testing.T) {
	type os struct {
		Os   string
		Arch string
	}
	windows32 := &os{"windows", "386"}
	windows64 := &os{"windows", "amd64"}
	linux32 := &os{"linux", "386"}
	linux64 := &os{"linux", "amd64"}
	linuxArm := &os{"linux", "arm"}
	linuxArmbe := &os{"linux", "armbe"}
	linuxArm64 := &os{"linux", "arm64"}
	darwin32 := &os{"darwin", "386"}
	darwin64 := &os{"darwin", "amd64"}
	darwinArm64 := &os{"darwin", "arm64"}
	freebsd32 := &os{"freebsd", "386"}
	freebsd64 := &os{"freebsd", "amd64"}
	oses := []*os{
		windows32,
		windows64,
		linux32,
		linux64,
		linuxArm,
		linuxArmbe,
		linuxArm64,
		darwin32,
		darwin64,
		darwinArm64,
		freebsd32,
		freebsd64,
	}

	type test struct {
		Flavour     *Flavor
		Compatibles []*os
		ExactMatch  []*os
	}
	tests := []*test{
		{&Flavor{OS: "i686-mingw32"}, []*os{windows32, windows64}, []*os{windows32}},
		{&Flavor{OS: "x86_64-mingw32"}, []*os{windows64}, []*os{windows64}},
		{&Flavor{OS: "i386-apple-darwin11"}, []*os{darwin32, darwin64, darwinArm64}, []*os{darwin32}},
		{&Flavor{OS: "x86_64-apple-darwin"}, []*os{darwin64, darwinArm64}, []*os{darwin64}},
		{&Flavor{OS: "arm64-apple-darwin"}, []*os{darwinArm64}, []*os{darwinArm64}},

		// Raspberry PI, BBB or other ARM based host
		// PI: "arm-linux-gnueabihf"
		// Raspbian on PI2: "arm-linux-gnueabihf"
		// Ubuntu Mate on PI2: "arm-linux-gnueabihf"
		// Debian 7.9 on BBB: "arm-linux-gnueabihf"
		// Raspbian on PI Zero: "arm-linux-gnueabihf"
		{&Flavor{OS: "arm-linux-gnueabihf"}, []*os{linuxArm, linuxArmbe}, []*os{linuxArm, linuxArmbe}},
		// Arch-linux on PI2: "armv7l-unknown-linux-gnueabihf"
		{&Flavor{OS: "armv7l-unknown-linux-gnueabihf"}, []*os{linuxArm, linuxArmbe}, []*os{linuxArm, linuxArmbe}},

		{&Flavor{OS: "i686-linux-gnu"}, []*os{linux32}, []*os{linux32}},
		{&Flavor{OS: "i686-pc-linux-gnu"}, []*os{linux32}, []*os{linux32}},
		{&Flavor{OS: "x86_64-linux-gnu"}, []*os{linux64}, []*os{linux64}},
		{&Flavor{OS: "x86_64-pc-linux-gnu"}, []*os{linux64}, []*os{linux64}},
		{&Flavor{OS: "aarch64-linux-gnu"}, []*os{linuxArm64}, []*os{linuxArm64}},
		{&Flavor{OS: "arm64-linux-gnu"}, []*os{linuxArm64}, []*os{linuxArm64}},
	}

	checkCompatible := func(test *test, os *os) {
		// if the os is in the "positive" set iCompatibleWith must return true...
		res, _ := test.Flavour.isCompatibleWith(os.Os, os.Arch)
		for _, compatibleOs := range test.Compatibles {
			if compatibleOs == os {
				require.True(t, res, "'%s' tag compatible with '%s,%s' pair", test.Flavour.OS, os.Os, os.Arch)
				return
			}
		}
		// ...otherwise false
		require.False(t, res, "'%s' tag compatible with '%s,%s' pair", test.Flavour.OS, os.Os, os.Arch)
	}
	checkExactMatch := func(test *test, os *os) {
		// if the os is in the "positive" set iExactMatchWith must return true...
		for _, positiveOs := range test.ExactMatch {
			if positiveOs == os {
				require.True(t, test.Flavour.isExactMatchWith(os.Os, os.Arch), "'%s' tag exact match with '%s,%s' pair", test.Flavour.OS, os.Os, os.Arch)
				return
			}
		}
		// ...otherwise false
		require.False(t, test.Flavour.isExactMatchWith(os.Os, os.Arch), "'%s' tag exact match with '%s,%s' pair", test.Flavour.OS, os.Os, os.Arch)
	}

	for _, test := range tests {
		for _, os := range oses {
			checkCompatible(test, os)
			checkExactMatch(test, os)
		}
	}
}

func TestFlavorPrioritySelection(t *testing.T) {
	res := (&ToolRelease{
		Flavors: []*Flavor{
			{OS: "i386-apple-darwin11", Resource: &resources.DownloadResource{ArchiveFileName: "1"}},
			{OS: "x86_64-apple-darwin", Resource: &resources.DownloadResource{ArchiveFileName: "2"}},
			{OS: "arm64-apple-darwin", Resource: &resources.DownloadResource{ArchiveFileName: "3"}},
		},
	}).GetFlavourCompatibleWith("darwin", "arm64")
	require.NotNil(t, res)
	require.Equal(t, "3", res.ArchiveFileName)

	res = (&ToolRelease{
		Flavors: []*Flavor{
			{OS: "i386-apple-darwin11", Resource: &resources.DownloadResource{ArchiveFileName: "1"}},
			{OS: "x86_64-apple-darwin", Resource: &resources.DownloadResource{ArchiveFileName: "2"}},
		},
	}).GetFlavourCompatibleWith("darwin", "arm64")
	require.NotNil(t, res)
	require.Equal(t, "2", res.ArchiveFileName)

	res = (&ToolRelease{
		Flavors: []*Flavor{
			{OS: "x86_64-apple-darwin", Resource: &resources.DownloadResource{ArchiveFileName: "2"}},
			{OS: "i386-apple-darwin11", Resource: &resources.DownloadResource{ArchiveFileName: "1"}},
		},
	}).GetFlavourCompatibleWith("darwin", "arm64")
	require.NotNil(t, res)
	require.Equal(t, "2", res.ArchiveFileName)

	res = (&ToolRelease{
		Flavors: []*Flavor{
			{OS: "i386-apple-darwin11", Resource: &resources.DownloadResource{ArchiveFileName: "1"}},
		},
	}).GetFlavourCompatibleWith("darwin", "arm64")
	require.NotNil(t, res)
	require.Equal(t, "1", res.ArchiveFileName)

	res = (&ToolRelease{
		Flavors: []*Flavor{
			{OS: "i686-mingw32", Resource: &resources.DownloadResource{ArchiveFileName: "1"}},
			{OS: "x86_64-mingw32", Resource: &resources.DownloadResource{ArchiveFileName: "2"}},
		},
	}).GetFlavourCompatibleWith("windows", "amd64")
	require.NotNil(t, res)
	require.Equal(t, "2", res.ArchiveFileName)

	res = (&ToolRelease{
		Flavors: []*Flavor{
			{OS: "x86_64-mingw32", Resource: &resources.DownloadResource{ArchiveFileName: "2"}},
			{OS: "i686-mingw32", Resource: &resources.DownloadResource{ArchiveFileName: "1"}},
		},
	}).GetFlavourCompatibleWith("windows", "amd64")
	require.NotNil(t, res)
	require.Equal(t, "2", res.ArchiveFileName)
}
