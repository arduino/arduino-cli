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

import "github.com/arduino/arduino-cli/configuration"

// Init initializes the i18n module, setting the locale according to this order of preference:
// 1. Configuration set in arduino-cli.yaml
// 2. OS Locale
// 3. en (default)
func Init() {
	initRiceBox()
	locales := supportedLocales()

	if configLocale := configuration.Settings.GetString("locale"); configLocale != "" {
		if locale := findMatchingLocale(configLocale, locales); locale != "" {
			setLocale(locale)
			return
		}
	}

	if osLocale := getLocaleIdentifierFromOS(); osLocale != "" {
		if locale := findMatchingLocale(osLocale, locales); locale != "" {
			setLocale(locale)
			return
		}
	}

	setLocale("en")
}

// Tr returns msg translated to the selected locale
// the msg argument must be a literal string
func Tr(msg string, args ...interface{}) string {
	return po.Get(msg, args...)
}
