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

package i18n

import (
	"os"
	"path/filepath"
	"strings"

	rice "github.com/cmaglie/go.rice"
	"github.com/leonelquinteros/gotext"
)

var (
	po  *gotext.Po
	box *rice.Box
)

func init() {
	po = new(gotext.Po)
}

func initRiceBox() {
	box = rice.MustFindBox("./data")
}

func supportedLocales() []string {
	var locales []string
	box.Walk("", func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".po" {
			locales = append(locales, strings.TrimSuffix(path, ".po"))
		}
		return nil
	})
	return locales
}

func findMatchingLanguage(language string, supportedLocales []string) string {
	var matchingLocales []string
	for _, supportedLocale := range supportedLocales {
		if strings.HasPrefix(supportedLocale, language) {
			matchingLocales = append(matchingLocales, supportedLocale)
		}
	}

	if len(matchingLocales) == 1 {
		return matchingLocales[0]
	}

	return ""
}

func findMatchingLocale(locale string, supportedLocales []string) string {
	for _, suportedLocale := range supportedLocales {
		if locale == suportedLocale {
			return suportedLocale
		}
	}

	parts := strings.Split(locale, "_")

	return findMatchingLanguage(parts[0], supportedLocales)
}

func setLocale(locale string) {
	poFile := box.MustBytes(locale + ".po")
	po = new(gotext.Po)
	po.Parse(poFile)
}
