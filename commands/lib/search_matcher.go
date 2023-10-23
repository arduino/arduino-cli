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

package lib

import (
	"strings"

	"github.com/arduino/arduino-cli/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/arduino/utils"
)

// matcherTokensFromQueryString parses the query string into tokens of interest
// for the qualifier-value pattern matching.
func matcherTokensFromQueryString(query string) []string {
	escaped := false
	quoted := false
	tokens := []string{}
	sb := &strings.Builder{}

	for _, r := range query {
		// Short circuit the loop on backslash so that all other paths can clear
		// the escaped flag.
		if !escaped && r == '\\' {
			escaped = true
			continue
		}

		if r == '"' {
			if !escaped {
				quoted = !quoted
			} else {
				sb.WriteRune(r)
			}
		} else if !quoted && r == ' ' {
			tokens = append(tokens, strings.ToLower(sb.String()))
			sb.Reset()
		} else {
			sb.WriteRune(r)
		}
		escaped = false
	}
	if sb.Len() > 0 {
		tokens = append(tokens, strings.ToLower(sb.String()))
	}

	return tokens
}

// defaulLibraryMatchExtractor returns a string describing the library that
// is used for the simple search.
func defaultLibraryMatchExtractor(lib *librariesindex.Library) string {
	res := lib.Name + " " +
		lib.Latest.Paragraph + " " +
		lib.Latest.Sentence + " " +
		lib.Latest.Author + " "
	for _, include := range lib.Latest.ProvidesIncludes {
		res += include + " "
	}
	return res
}

var qualifiers map[string]func(lib *librariesindex.Library) string = map[string]func(lib *librariesindex.Library) string{
	"name":          func(lib *librariesindex.Library) string { return lib.Name },
	"architectures": func(lib *librariesindex.Library) string { return strings.Join(lib.Latest.Architectures, " ") },
	"author":        func(lib *librariesindex.Library) string { return lib.Latest.Author },
	"category":      func(lib *librariesindex.Library) string { return lib.Latest.Category },
	"dependencies": func(lib *librariesindex.Library) string {
		names := make([]string, len(lib.Latest.Dependencies))
		for i, dep := range lib.Latest.Dependencies {
			names[i] = dep.GetName()
		}
		return strings.Join(names, " ")
	},
	"license":    func(lib *librariesindex.Library) string { return lib.Latest.License },
	"maintainer": func(lib *librariesindex.Library) string { return lib.Latest.Maintainer },
	"paragraph":  func(lib *librariesindex.Library) string { return lib.Latest.Paragraph },
	"provides":   func(lib *librariesindex.Library) string { return strings.Join(lib.Latest.ProvidesIncludes, " ") },
	"sentence":   func(lib *librariesindex.Library) string { return lib.Latest.Sentence },
	"types":      func(lib *librariesindex.Library) string { return strings.Join(lib.Latest.Types, " ") },
	"version":    func(lib *librariesindex.Library) string { return lib.Latest.Version.String() },
	"website":    func(lib *librariesindex.Library) string { return lib.Latest.Website },
}

// MatcherFromQueryString returns a closure that takes a library as a
// parameter and returns true if the library matches the query.
func MatcherFromQueryString(query string) func(*librariesindex.Library) bool {
	// A qv-query is one using <qualifier>[:=]<value> syntax.
	qvQuery := strings.Contains(query, ":") || strings.Contains(query, "=")

	if !qvQuery {
		queryTerms := utils.SearchTermsFromQueryString(query)
		return func(lib *librariesindex.Library) bool {
			return utils.Match(defaultLibraryMatchExtractor(lib), queryTerms)
		}
	}

	queryTerms := matcherTokensFromQueryString(query)

	return func(lib *librariesindex.Library) bool {
		matched := true
		for _, term := range queryTerms {
			if sepIdx := strings.IndexAny(term, ":="); sepIdx != -1 {
				qualifier, separator, target := term[:sepIdx], term[sepIdx], term[sepIdx+1:]
				if extractor, ok := qualifiers[qualifier]; ok {
					switch separator {
					case ':':
						matched = (matched && utils.Match(extractor(lib), []string{target}))
						continue
					case '=':
						matched = (matched && strings.ToLower(extractor(lib)) == target)
						continue
					}
				}
			}
			// We perform the usual match in the following cases:
			// 1. Unknown qualifier names revert to basic search terms.
			// 2. Terms that do not use qv-syntax.
			matched = (matched && utils.Match(defaultLibraryMatchExtractor(lib), []string{term}))
		}
		return matched
	}
}
