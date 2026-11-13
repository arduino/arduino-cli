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
	"path/filepath"

	"golang.org/x/sys/windows"
)

// ToShortPath converts path to its 8.3 short form, which is plain ASCII and
// can therefore be opened by old Windows GCC toolchains that only use the
// ANSI APIs (see ShortenCompilerArgs). ASCII paths are returned unchanged,
// as are paths for which no usable short form exists (for example when 8.3
// name creation is disabled on the volume).
func ToShortPath(path string) string {
	if path == "" || isASCII(path) {
		return path
	}
	if short, ok := getShortPathName(path); ok {
		return short
	}
	if !filepath.IsAbs(path) {
		return path
	}
	// The path may not exist yet (e.g. compiler outputs under a non-ASCII
	// directory): convert the closest existing ancestor and re-append the
	// remainder.
	var suffix []string
	curr := path
	for range 32 {
		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		suffix = append(suffix, filepath.Base(curr))
		curr = parent
		if _, err := os.Stat(curr); err != nil {
			continue
		}
		short, ok := getShortPathName(curr)
		if !ok {
			return path
		}
		for i := len(suffix) - 1; i >= 0; i-- {
			short = filepath.Join(short, suffix[i])
		}
		return short
	}
	return path
}

// getShortPathName returns the 8.3 short form of an existing path.
// ok is false when no usable short path exists (8.3 names disabled,
// path not found, or the result still contains non-ASCII characters).
func getShortPathName(path string) (short string, ok bool) {
	longp, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return "", false
	}
	n, err := windows.GetShortPathName(longp, nil, 0)
	if err != nil || n == 0 {
		return "", false
	}
	buf := make([]uint16, n)
	n, err = windows.GetShortPathName(longp, &buf[0], n)
	if err != nil || n == 0 {
		return "", false
	}
	short = windows.UTF16ToString(buf)
	if short == "" || !isASCII(short) {
		// With 8.3 name creation disabled the API returns the long
		// path unchanged, which does not help.
		return "", false
	}
	return short, true
}
