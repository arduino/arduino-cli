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

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocaleMatch(t *testing.T) {
	supportedLocales := []string{
		"en",
		"pt_BR",
		"it_IT",
		"es_CO",
		"es_ES",
	}

	require.Equal(t, "pt_BR", findMatchingLocale("pt", supportedLocales), "Language match")
	require.Equal(t, "pt_BR", findMatchingLocale("pt_BR", supportedLocales), "Exact match")
	require.Equal(t, "pt_BR", findMatchingLocale("pt_PT", supportedLocales), "Language match with country")
	require.Equal(t, "", findMatchingLocale("es", supportedLocales), "Multiple languages match")
	require.Equal(t, "", findMatchingLocale("zn_CH", supportedLocales), "Not supported")
}
