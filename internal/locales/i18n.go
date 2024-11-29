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

package locales

// Init initializes the i18n module, setting the locale according to this order of preference:
// 1. Locale specified via the function call
// 2. OS Locale
// 3. en (default)
func Init(configLocale string) {
	locales := supportedLocales()
	if configLocale != "" {
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
