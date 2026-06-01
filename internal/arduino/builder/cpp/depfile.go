// This file is part of arduino-cli.
//
// Copyright 2025 ARDUINO SA (http://www.arduino.cc/)
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
	"errors"
	"runtime"
	"strings"
	"unicode"

	"github.com/arduino/go-paths-helper"
	"go.bug.st/f"
)

// Dependencies represents the dependencies of a source file.
type Dependencies struct {
	ObjectFile   string
	Dependencies []string
}

// ReadDepFile reads a dependency file and returns the dependencies.
// It may return nil if the dependency file is empty.
func ReadDepFile(depFilePath *paths.Path) (*Dependencies, error) {
	depFileData, err := depFilePath.ReadFile()
	if err != nil {
		return nil, err
	}

	if runtime.GOOS == "windows" {
		// This is required because on Windows we don't know which encoding is used
		// by gcc to write the dep file (it could be UTF-8 or any of the Windows
		// ANSI mappings).
		if decoded, err := convertAnsiBytesToString(depFileData); err == nil {
			if res, err := readDepFile(decoded); err == nil && res != nil {
				return res, nil
			}
		}
		// Fallback to UTF-8...
	}

	return readDepFile(string(depFileData))
}

func readDepFile(depFile string) (*Dependencies, error) {
	rows, err := unescapeAndSplit(depFile)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return &Dependencies{}, nil
	}

	if !strings.HasSuffix(rows[0], ":") {
		return nil, errors.New("no colon in first item of depfile")
	}
	res := &Dependencies{
		ObjectFile:   strings.TrimSuffix(rows[0], ":"),
		Dependencies: rows[1:],
	}
	return res, nil
}

func unescapeAndSplit(s string) ([]string, error) {
	var res []string
	backslash := false
	dollar := false
	current := strings.Builder{}
	for _, c := range s {
		if c == '\r' {
			// Ignore CR (Windows line ending style immediately followed by LF)
			continue
		}
		if backslash {
			switch c {
			case ' ':
				current.WriteRune(' ')
			case '#':
				current.WriteRune('#')
			case '\\':
				current.WriteRune('\\')
			case '\n':
				// ignore
			default:
				current.WriteRune('\\')
				current.WriteRune(c)
			}
			backslash = false
			continue
		}
		if dollar {
			if c != '$' {
				return nil, errors.New("invalid dollar sequence: $" + string(c))
			}
			current.WriteByte('$')
			dollar = false
			continue
		}

		if c == '\\' {
			backslash = true
			continue
		}
		if c == '$' {
			dollar = true
			continue
		}

		if unicode.IsSpace(c) {
			if current.Len() > 0 {
				res = append(res, current.String())
				current.Reset()
			}
			continue
		}
		current.WriteRune(c)
	}
	if dollar {
		return nil, errors.New("unclosed escape sequence at end of depfile")
	}
	if current.Len() > 0 {
		res = append(res, current.String())
	}
	res = f.Map(res, strings.TrimSpace)
	res = f.Filter(res, f.NotEquals(""))
	return res, nil
}
