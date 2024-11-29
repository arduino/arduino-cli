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

package po

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMerge(t *testing.T) {
	r := strings.NewReader(`
msgid "a"
msgstr ""

msgid "b"
msgstr "value-b"
	`)
	catalogA := ParseReader(r)

	r = strings.NewReader(`
msgid "a"
msgstr "value-a"

msgid "b"
msgstr "value-b"

msgid "c"
msgstr "value-c"
	`)

	catalogB := ParseReader(r)

	mergedCatalog := Merge(catalogA, catalogB)

	var buf bytes.Buffer
	mergedCatalog.Write(&buf)

	require.Equal(t, `msgid "a"
msgstr "value-a"

msgid "b"
msgstr "value-b"

`, buf.String())

}
