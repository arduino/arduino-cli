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

package cpp

import (
	"os"
	"strings"
)

// ShortenCompilerArgs returns a copy of args in which paths handed to the
// compiler are converted to their 8.3 short form on Windows.
//
// Old GCC releases shipped with some Arduino toolchains (notably AVR gcc
// 7.3.0) use the Windows ANSI APIs and fail to open files whose path
// contains characters outside the active code page, producing a spurious
// "No such file or directory" error even though the file exists.
// See: https://github.com/arduino/arduino-cli/issues/1239
// Upstream: https://gcc.gnu.org/bugzilla/show_bug.cgi?id=108865 (fixed in GCC 13,
// requested for AVR in https://github.com/arduino/toolchain-avr/issues/95)
//
// Only existing on-disk paths and -I include directories are converted;
// flags are left untouched. On non-Windows platforms, and when a short
// path is unavailable (e.g. 8.3 names disabled on the volume), the
// arguments are returned unchanged.
func ShortenCompilerArgs(args []string) []string {
	res := make([]string, len(args))
	copy(res, args)
	for i, arg := range args {
		if p, ok := strings.CutPrefix(arg, "-I"); ok {
			if p != "" {
				res[i] = "-I" + ToShortPath(p)
			} else if i+1 < len(args) {
				// Split "-I <path>" form
				res[i+1] = ToShortPath(args[i+1])
			}
			continue
		}
		if _, err := os.Stat(arg); err == nil {
			res[i] = ToShortPath(arg)
		}
	}
	return res
}

// isASCII reports whether s contains only ASCII characters.
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return false
		}
	}
	return true
}
