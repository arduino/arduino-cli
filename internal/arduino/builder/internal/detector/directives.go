// This file is part of arduino-cli.
//
// Copyright 2024 ARDUINO SA (http://www.arduino.cc/)
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

package detector

import (
	"bytes"
	"regexp"
	"slices"
	"strings"

	"github.com/arduino/go-paths-helper"
)

// sketchDirectives returns the preprocessor directives of the merged sketch, in order.
func sketchDirectives(mergedSketch *paths.Path) ([]string, error) {
	source, err := mergedSketch.ReadFile()
	if err != nil {
		return nil, err
	}
	return directiveLines(source), nil
}

var lineContinuation = regexp.MustCompile(`\\[ \t\f\v]*\n`)

// directiveLines returns the preprocessor directives of a C++ source, in order. The source is
// read as the compiler reads it: continued lines are joined, comments are left out and the
// lines that start with a # are the directives. They cannot tell a directive that a condition
// turns off: they only select the edits worth the preprocessor pass that confirms a replay of
// the detection.
func directiveLines(source []byte) []string {
	source = bytes.ReplaceAll(source, []byte("\r\n"), []byte("\n"))
	source = bytes.ReplaceAll(source, []byte("\r"), []byte("\n"))
	source = lineContinuation.ReplaceAll(source, nil)
	directives := []string{}
	var line []byte
	endLine := func() {
		text := bytes.TrimSpace(line)
		if bytes.HasPrefix(text, []byte("#")) || bytes.HasPrefix(text, []byte("%:")) {
			// Bytes that are not text are recorded as the cache file stores them
			directives = append(directives, string([]rune(string(text))))
		}
		line = line[:0]
	}
	for pos := 0; pos < len(source); {
		rest := source[pos:]
		length := 1
		switch {
		case rest[0] == '\n':
			endLine()
			pos++
			continue
		case bytes.HasPrefix(rest, []byte("//")):
			pos += lengthBefore(rest, "\n")
			continue
		case bytes.HasPrefix(rest, []byte("/*")):
			line = append(line, ' ')
			pos += 2 + lengthThrough(rest[2:], "*/")
			continue
		case rest[0] == '"' && isRawStringPrefix(lastWord(line)):
			length = rawStringLength(rest)
		case rest[0] == '"', rest[0] == '\'' && !isDigitSeparator(line, rest):
			length = quotedLength(rest)
		}
		line = append(line, rest[:length]...)
		pos += length
	}
	endLine()
	return directives
}

// followCode reports whether a directive names a macro whose value changes with the code
// around it: an edit of the code alone can then change what the directives do.
func followCode(directives []string) bool {
	return slices.ContainsFunc(directives, func(directive string) bool {
		return strings.Contains(directive, "__LINE__") || strings.Contains(directive, "__COUNTER__")
	})
}

func lengthBefore(text []byte, mark string) int {
	if i := bytes.Index(text, []byte(mark)); i >= 0 {
		return i
	}
	return len(text)
}

func lengthThrough(text []byte, mark string) int {
	if i := bytes.Index(text, []byte(mark)); i >= 0 {
		return i + len(mark)
	}
	return len(text)
}

// lastWord returns the letters, digits and underscores the line ends with.
func lastWord(line []byte) []byte {
	start := len(line)
	for start > 0 && isWordChar(line[start-1]) {
		start--
	}
	return line[start:]
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }
func isWordChar(c byte) bool {
	return c == '_' || isDigit(c) || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

// isDigitSeparator tells the quote inside a number, as in 1'000 or 0xFF'A5'00, from the one
// that opens a character.
func isDigitSeparator(line, rest []byte) bool {
	start := len(line)
	for start > 0 && (line[start-1] == '\'' || isWordChar(line[start-1])) {
		start--
	}
	return start < len(line) && isDigit(line[start]) && len(rest) > 1 && isWordChar(rest[1])
}

func isRawStringPrefix(word []byte) bool {
	return slices.Contains([]string{"R", "LR", "uR", "UR", "u8R"}, string(word))
}

// quotedLength returns the length of the string or character the text starts with, which ends
// with the line when its closing quote is missing.
func quotedLength(text []byte) int {
	for i := 1; i < len(text); i++ {
		switch text[i] {
		case '\\':
			i++
		case text[0]:
			return i + 1
		case '\n':
			return i
		}
	}
	return len(text)
}

// rawStringLength returns the length of the raw string the text starts with, as in "tag(...)tag".
func rawStringLength(text []byte) int {
	open := lengthBefore(text, "(")
	closing := ")" + string(text[1:open]) + `"`
	return open + lengthThrough(text[open:], closing)
}
