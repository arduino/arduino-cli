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

package po

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	r := strings.NewReader(`
msgid "\"a\"\nline"
msgstr "value\nline"
	`)
	catalog := ParseReader(r)

	var buf bytes.Buffer
	catalog.Write(&buf)

	require.Equal(t, `msgid "\"a\"\n"
"line"
msgstr "value\n"
"line"

`, buf.String())
}
