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

package librariesresolver

import (
	"testing"

	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/stretchr/testify/require"
)

var l1 = &libraries.Library{Name: "Calculus Lib", Location: libraries.Sketchbook}
var l2 = &libraries.Library{Name: "Calculus Lib-master", Location: libraries.Sketchbook}
var l3 = &libraries.Library{Name: "Calculus Lib Improved", Location: libraries.Sketchbook}
var l4 = &libraries.Library{Name: "Another Calculus Lib", Location: libraries.Sketchbook}
var l5 = &libraries.Library{Name: "Yet Another Calculus Lib Improved", Location: libraries.Sketchbook}
var l6 = &libraries.Library{Name: "Calculus Unified Lib", Location: libraries.Sketchbook}
var l7 = &libraries.Library{Name: "AnotherLib", Location: libraries.Sketchbook}

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
		return resolver.ResolveFor(header, "avr").Name
	}
	require.Equal(t, "Calculus Lib", resolve("calculus_lib.h", l1, l2, l3, l4, l5, l6, l7))
	require.Equal(t, "Calculus Lib-master", resolve("calculus_lib.h", l2, l3, l4, l5, l6, l7))
	require.Equal(t, "Calculus Lib Improved", resolve("calculus_lib.h", l3, l4, l5, l6, l7))
	require.Equal(t, "Another Calculus Lib", resolve("calculus_lib.h", l4, l5, l6, l7))
	require.Equal(t, "Yet Another Calculus Lib Improved", resolve("calculus_lib.h", l5, l6, l7))
	require.Equal(t, "Calculus Unified Lib", resolve("calculus_lib.h", l6, l7))
	require.Equal(t, "Calculus Unified Lib", resolve("calculus_lib.h", l7, l6))
}
