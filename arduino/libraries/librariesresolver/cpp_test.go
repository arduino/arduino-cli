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

package librariesresolver

import (
	"testing"

	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/stretchr/testify/require"
)

var l1 = &libraries.Library{CanonicalName: "Calculus Lib", Location: libraries.User}
var l2 = &libraries.Library{CanonicalName: "Calculus Lib-master", Location: libraries.User}
var l3 = &libraries.Library{CanonicalName: "Calculus Lib Improved", Location: libraries.User}
var l4 = &libraries.Library{CanonicalName: "Another Calculus Lib", Location: libraries.User}
var l5 = &libraries.Library{CanonicalName: "Yet Another Calculus Lib Improved", Location: libraries.User}
var l6 = &libraries.Library{CanonicalName: "Calculus Unified Lib", Location: libraries.User}
var l7 = &libraries.Library{CanonicalName: "AnotherLib", Location: libraries.User}
var bundleServo = &libraries.Library{CanonicalName: "Servo", Location: libraries.IDEBuiltIn, Architectures: []string{"avr", "sam", "samd"}}

func runResolver(include string, arch string, libs ...*libraries.Library) *libraries.Library {
	libraryList := libraries.List{}
	libraryList.Add(libs...)
	resolver := NewCppResolver()
	resolver.headers[include] = libraryList
	return resolver.ResolveFor(include, arch)
}

func TestArchitecturePriority(t *testing.T) {
	userServo := &libraries.Library{
		CanonicalName: "Servo",
		Location:      libraries.User,
		Architectures: []string{"avr", "sam", "samd"}}
	userServoAllArch := &libraries.Library{
		CanonicalName: "Servo",
		Location:      libraries.User,
		Architectures: []string{"*"}}
	userServoNonavr := &libraries.Library{
		CanonicalName: "Servo",
		Location:      libraries.User,
		Architectures: []string{"sam", "samd"}}
	userAnotherServo := &libraries.Library{
		CanonicalName: "AnotherServo",
		Location:      libraries.User,
		Architectures: []string{"avr", "sam", "samd", "esp32"}}

	res := runResolver("Servo.h", "avr", bundleServo, userServo)
	require.NotNil(t, res)
	require.Equal(t, userServo, res, "selected library")

	res = runResolver("Servo.h", "avr", bundleServo, userServoNonavr)
	require.NotNil(t, res)
	require.Equal(t, bundleServo, res, "selected library")

	res = runResolver("Servo.h", "avr", bundleServo, userAnotherServo)
	require.NotNil(t, res)
	require.Equal(t, bundleServo, res, "selected library")

	res = runResolver("Servo.h", "esp32", bundleServo, userAnotherServo)
	require.NotNil(t, res)
	require.Equal(t, userAnotherServo, res, "selected library")

	res = runResolver("Servo.h", "esp32", userServoAllArch, userAnotherServo)
	require.NotNil(t, res)
	require.Equal(t, userServoAllArch, res, "selected library")

	userSDAllArch := &libraries.Library{
		CanonicalName: "SD",
		Location:      libraries.User,
		Architectures: []string{"*"}}
	builtinSDesp := &libraries.Library{
		CanonicalName: "SD",
		Location:      libraries.PlatformBuiltIn,
		Architectures: []string{"esp8266"}}
	res = runResolver("SD.h", "esp8266", userSDAllArch, builtinSDesp)
	require.Equal(t, builtinSDesp, res, "selected library")
}

func TestClosestMatchWithTotallyDifferentNames(t *testing.T) {
	libraryList := libraries.List{}
	libraryList.Add(l5)
	libraryList.Add(l6)
	libraryList.Add(l7)
	resolver := NewCppResolver()
	resolver.headers["XYZ.h"] = libraryList
	res := resolver.ResolveFor("XYZ.h", "xyz")
	require.NotNil(t, res)
	require.Equal(t, l7, res, "selected library")
}

func TestCppHeaderPriority(t *testing.T) {
	r1 := computePriority(l1, "calculus_lib.h", "avr")
	r2 := computePriority(l2, "calculus_lib.h", "avr")
	r3 := computePriority(l3, "calculus_lib.h", "avr")
	r4 := computePriority(l4, "calculus_lib.h", "avr")
	r5 := computePriority(l5, "calculus_lib.h", "avr")
	r6 := computePriority(l6, "calculus_lib.h", "avr")
	r7 := computePriority(l7, "calculus_lib.h", "avr")
	require.True(t, r1 > r2)
	require.True(t, r2 > r3)
	require.True(t, r3 > r4)
	require.True(t, r4 > r5)
	require.True(t, r5 > r6)
	require.True(t, r6 == r7)
}

func TestCppHeaderResolverWithNilResult(t *testing.T) {
	resolver := NewCppResolver()
	libraryList := libraries.List{}
	libraryList.Add(l1)
	resolver.headers["aaa.h"] = libraryList
	require.Nil(t, resolver.ResolveFor("bbb.h", "avr"))
}

func TestCppHeaderResolver(t *testing.T) {
	resolve := func(header string, libs ...*libraries.Library) string {
		resolver := NewCppResolver()
		librarylist := libraries.List{}
		for _, lib := range libs {
			librarylist.Add(lib)
		}
		resolver.headers[header] = librarylist
		return resolver.ResolveFor(header, "avr").CanonicalName
	}
	require.Equal(t, "Calculus Lib", resolve("calculus_lib.h", l1, l2, l3, l4, l5, l6, l7))
	require.Equal(t, "Calculus Lib-master", resolve("calculus_lib.h", l2, l3, l4, l5, l6, l7))
	require.Equal(t, "Calculus Lib Improved", resolve("calculus_lib.h", l3, l4, l5, l6, l7))
	require.Equal(t, "Another Calculus Lib", resolve("calculus_lib.h", l4, l5, l6, l7))
	require.Equal(t, "Yet Another Calculus Lib Improved", resolve("calculus_lib.h", l5, l6, l7))
	require.Equal(t, "Calculus Unified Lib", resolve("calculus_lib.h", l6, l7))
	require.Equal(t, "Calculus Unified Lib", resolve("calculus_lib.h", l7, l6))
}

func TestCppHeaderResolverWithLibrariesInStrangeDirectoryNames(t *testing.T) {
	resolver := NewCppResolver()
	librarylist := libraries.List{}
	librarylist.Add(&libraries.Library{CanonicalName: "onewire_2_3_4", Name: "OneWire", Architectures: []string{"*"}})
	librarylist.Add(&libraries.Library{CanonicalName: "onewireng_2_3_4", Name: "OneWireNg", Architectures: []string{"avr"}})
	resolver.headers["OneWire.h"] = librarylist
	require.Equal(t, "onewire_2_3_4", resolver.ResolveFor("OneWire.h", "avr").CanonicalName)

	librarylist2 := libraries.List{}
	librarylist2.Add(&libraries.Library{CanonicalName: "OneWire", Name: "OneWire", Architectures: []string{"*"}})
	librarylist2.Add(&libraries.Library{CanonicalName: "onewire_2_3_4", Name: "OneWire", Architectures: []string{"avr"}})
	resolver.headers["OneWire.h"] = librarylist2
	require.Equal(t, "OneWire", resolver.ResolveFor("OneWire.h", "avr").CanonicalName)
}
