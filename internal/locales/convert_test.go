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

package locales

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func format(format string, a ...interface{}) string {
	format = FromJavaToGoSyntax(format)
	message := fmt.Sprintf(format, a...)
	return message
}

func TestI18NSyntax(t *testing.T) {
	require.Equal(t, "Do you want to remove %[1]v?\nIf you do so you won't be able to use %[1]v any more.", FromJavaToGoSyntax("Do you want to remove {0}?\nIf you do so you won't be able to use {0} any more."))
	require.Equal(t, "A file named \"%[1]v\" already exists in \"%[2]v\"", FromJavaToGoSyntax("A file named \"{0}\" already exists in \"{1}\""))
	require.Equal(t, "Board %[1]v:%[2]v:%[3]v doesn't define a 'build.board' preference. Auto-set to: %[4]v", FromJavaToGoSyntax("Board {0}:{1}:{2} doesn''t define a ''build.board'' preference. Auto-set to: {3}"))

	require.Equal(t, "22 11\n", fmt.Sprintf("%[2]d %[1]d\n", 11, 22))
	require.Equal(t, "d c b a", format("{3} {2} {1} {0}", "a", "b", "c", "d"))

	require.Equal(t, "e d b a", format("{4} {3} {1} {0}", "a", "b", "c", "d", "e"))

	require.Equal(t, "a b", format("{0} {1}", "a", "b", "c", "d", "e"))

	require.Equal(t, "%!v(BADINDEX) c b a", format("{3} {2} {1} {0}", "a", "b", "c"))
	require.Equal(t, "%!v(BADINDEX) %!v(BADINDEX) %!v(BADINDEX) %!v(BADINDEX)", format("{3} {2} {1} {0}"))

	require.Equal(t, "I'm %[1]v%% sure", FromJavaToGoSyntax("I'm {0}%% sure"))
	require.Equal(t, "I'm 100% sure", format("I'm {0}%% sure", 100))

	require.Equal(t, "Either in [1] or in [2]", format("Either in [{0}] or in [{1}]", 1, 2))

	require.Equal(t, "Using library a at version b in folder: c ", format("Using library {0} at version {1} in folder: {2} {3}", "a", "b", "c", ""))

	require.Equal(t, "Missing 'a' from library in b", format("Missing '{0}' from library in {1}", "a", "b"))
}
