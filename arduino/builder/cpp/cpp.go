// This file is part of arduino-cli.
//
// Copyright 2023 ARDUINO SA (http://www.arduino.cc/)
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
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/arduino/go-paths-helper"
)

// QuoteString returns the given string as a quoted string for use with the C
// preprocessor. This adds double quotes around it and escapes any
// double quotes and backslashes in the string.
func QuoteString(str string) string {
	str = strings.ReplaceAll(str, "\\", "\\\\")
	str = strings.ReplaceAll(str, "\"", "\\\"")
	return "\"" + str + "\""
}

// ParseLineMarker parses the given line as a gcc line marker and returns the contained
// filename.
func ParseLineMarker(line string) *paths.Path {
	// A line marker contains the line number and filename and looks like:
	// # 123 /path/to/file.cpp
	// It can be followed by zero or more flag number that indicate the
	// preprocessor state and can be ignored.
	// For exact details on this format, see:
	// https://github.com/gcc-mirror/gcc/blob/edd716b6b1caa1a5cb320a8cd7f626f30198e098/gcc/c-family/c-ppoutput.c#L413-L415

	split := strings.SplitN(line, " ", 3)
	if len(split) < 3 || len(split[0]) == 0 || split[0][0] != '#' {
		return nil
	}

	_, err := strconv.Atoi(split[1])
	if err != nil {
		return nil
	}

	// If we get here, we found a # followed by a line number, so
	// assume this is a line marker and see if the rest of the line
	// starts with a string containing the filename
	str, rest, ok := ParseString(split[2])

	if ok && (rest == "" || rest[0] == ' ') {
		return paths.New(str)
	}
	return nil
}

// ParseString parse a string as emitted by the preprocessor. This
// is a string contained in double quotes, with any backslashes or
// quotes escaped with a backslash. If a valid string was present at the
// start of the given line, returns the unquoted string contents, the
// remainder of the line (everything after the closing "), and true.
// Otherwise, returns the empty string, the entire line and false.
func ParseString(line string) (string, string, bool) {
	// For details about how these strings are output by gcc, see:
	// https://github.com/gcc-mirror/gcc/blob/a588355ab948cf551bc9d2b89f18e5ae5140f52c/libcpp/macro.c#L491-L511
	// Note that the documentation suggests all non-printable
	// characters are also escaped, but the implementation does not
	// actually do this. See https://gcc.gnu.org/bugzilla/show_bug.cgi?id=51259
	if len(line) < 1 || line[0] != '"' {
		return "", line, false
	}

	i := 1
	res := ""
	for {
		if i >= len(line) {
			return "", line, false
		}

		c, width := utf8.DecodeRuneInString(line[i:])

		switch c {
		case '\\':
			// Backslash, next character is used unmodified
			i += width
			if i >= len(line) {
				return "", line, false
			}
			res += string(line[i])
		case '"':
			// Quote, end of string
			return res, line[i+width:], true
		default:
			res += string(c)
		}

		i += width
	}
}

// WrapWithHyphenI fixdoc
func WrapWithHyphenI(value string) string {
	return "\"-I" + value + "\""
}
