/*
 * This file is part of PropertiesMap library.
 *
 * Copyright 2017-2018 Arduino AG (http://www.arduino.cc/)
 *
 * PropertiesMap library is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 */

package properties

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// SplitQuotedString splits a string by spaces and at the same time allows
// to use spaces in a single element of the split using quote characters.
//
// For example the call:
//   SplitQuotedString(`This 'is an' "Hello World!" example`, `'"`, false)
// returns the following array:
//   []string{"This", "is an", "Hello World!", "example"}
func SplitQuotedString(src string, quoteChars string, acceptEmptyArguments bool) ([]string, error) {
	// Make a map of valid quote runes
	isQuote := map[rune]bool{}
	for _, c := range quoteChars {
		isQuote[c] = true
	}

	result := []string{}

	var escapingChar rune
	escapedArg := ""

	for _, current := range strings.Split(src, " ") {
		if escapingChar == 0 {
			first, size := firstRune(current)
			if !isQuote[first] {
				if acceptEmptyArguments || len(strings.TrimSpace(current)) > 0 {
					result = append(result, current)
				}
				continue
			}

			escapingChar = first
			current = current[size:]
			escapedArg = ""
		}

		last, size := lastRune(current)
		if last != escapingChar {
			escapedArg += current + " "
			continue
		}

		escapedArg += current[:len(current)-size]
		if acceptEmptyArguments || len(strings.TrimSpace(escapedArg)) > 0 {
			result = append(result, escapedArg)
		}
		escapingChar = 0
	}

	if escapingChar != 0 {
		return nil, fmt.Errorf("invalid quoting, no closing `%c` char found", escapingChar)
	}

	return result, nil
}

func firstRune(s string) (rune, int) {
	if len(s) == 0 || !utf8.ValidString(s) {
		return 0, 0
	}
	return utf8.DecodeRuneInString(s)
}

func lastRune(s string) (rune, int) {
	if len(s) == 0 || !utf8.ValidString(s) {
		return 0, 0
	}
	return utf8.DecodeLastRuneInString(s)
}
