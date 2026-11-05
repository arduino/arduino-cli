// This file is part of arduino-cli.
//
// Copyright 2026 ARDUINO SA (http://www.arduino.cc/)
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

package cpp_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/arduino/arduino-cli/internal/arduino/builder/cpp"
	"github.com/stretchr/testify/require"
)

// Regression test for https://github.com/arduino/arduino-cli/issues/1239:
// compiler arguments that reference non-ASCII paths must be converted to a
// form old Windows GCC toolchains can open, without affecting anything else.
func TestToShortPathPassthrough(t *testing.T) {
	// Empty and ASCII paths are never converted, on any platform.
	require.Equal(t, "", cpp.ToShortPath(""))
	require.Equal(t, `C:\Users\test\libraries\Servo`, cpp.ToShortPath(`C:\Users\test\libraries\Servo`))
	require.Equal(t, "/tmp/arduino/build", cpp.ToShortPath("/tmp/arduino/build"))

	// Non-ASCII paths without an existing convertible ancestor fall back
	// to the original on every platform.
	nonExistent := filepath.Join("does", "not", "мікроконтролер", "exist")
	require.Equal(t, nonExistent, cpp.ToShortPath(nonExistent))
}

func TestShortenCompilerArgs(t *testing.T) {
	// Flags are left untouched.
	flags := []string{"avr-g++", "-w", "-x", "c++", "-E", "-CC", "-MMD", "-DFOO=bar"}
	require.Equal(t, flags, cpp.ShortenCompilerArgs(flags))

	// Joined -I with an ASCII path is unchanged.
	includes := []string{`-IC:\Users\test\libraries\Servo\src`}
	require.Equal(t, includes, cpp.ShortenCompilerArgs(includes))

	// Non-existent -I target falls back to the original.
	missing := []string{"-I" + filepath.Join(t.TempDir(), "мікроконтролер")}
	require.Equal(t, missing, cpp.ShortenCompilerArgs(missing))

	// Bare existing ASCII path args are unchanged on every platform.
	src := filepath.Join(t.TempDir(), "sketch.ino.cpp")
	require.NoError(t, os.WriteFile(src, []byte("x"), 0o644))
	require.Equal(t, []string{src}, cpp.ShortenCompilerArgs([]string{src}))

	// Existing non-ASCII directory: no-op off Windows; short ASCII form
	// (or unchanged fallback when 8.3 names are unavailable) on Windows.
	nonASCII := filepath.Join(t.TempDir(), "мікроконтролер", "libraries", "Servo")
	require.NoError(t, os.MkdirAll(nonASCII, 0o755))
	in := []string{"-I" + nonASCII, "-w"}
	orig := append([]string(nil), in...)
	got := cpp.ShortenCompilerArgs(in)
	require.Equal(t, orig, in, "input slice must not be mutated")
	require.Len(t, got, 2)
	require.Equal(t, "-w", got[1])
	if runtime.GOOS != "windows" {
		require.Equal(t, "-I"+nonASCII, got[0])
	} else {
		require.True(t,
			got[0] == "-I"+nonASCII || isASCII(got[0]),
			"expected short ASCII form or unchanged fallback, got %q", got[0])
	}
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return false
		}
	}
	return true
}
