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

package utils

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// removeDiatrics removes accents and similar diatrics from unicode characters.
// An empty string is returned in case of errors.
// This might not be the best solution but it works well enough for our usecase,
// in the future we might want to use the golang.org/x/text/secure/precis package
// when its API will be finalized.
// From https://stackoverflow.com/a/26722698
func removeDiatrics(s string) (string, error) {
	transformer := transform.Chain(
		norm.NFD,
		runes.Remove(runes.In(unicode.Mn)),
		norm.NFC,
	)
	s, _, err := transform.String(transformer, s)
	if err != nil {
		return "", err
	}
	return s, nil
}

// SearchTermsFromQueryString returns the terms inside the query string.
// All non alphanumeric characters (expect ':') are considered separators.
// All search terms are converted to lowercase.
func SearchTermsFromQueryString(query string) []string {
	// Split on anything but 0-9, a-z or :
	return strings.FieldsFunc(strings.ToLower(query), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != ':'
	})
}

// Match returns true if all substrings are contained in str.
// Both str and substrings are transforms to lower case and have their
// accents and other unicode diatrics removed.
// If strings transformation fails an error is returned.
func Match(str string, substrings []string) bool {
	clean := func(s string) string {
		s = strings.ToLower(s)
		if s2, err := removeDiatrics(s); err == nil {
			return s2
		}
		return s
	}

	str = clean(str)
	for _, sub := range substrings {
		if !strings.Contains(str, clean(sub)) {
			return false
		}
	}
	return true
}

// MatchAny checks if query matches at least one of the
// string in arrayToMatch using the utils.Match function.
func MatchAny(query string, arrayToMatch []string) bool {
	queryArgs := strings.Split(strings.TrimSpace(query), " ")
	if len(queryArgs) == 0 {
		return true
	}
	for _, t := range arrayToMatch {
		if Match(t, queryArgs) {
			return true
		}
	}
	return false
}
