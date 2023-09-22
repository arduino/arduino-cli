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

package monitor

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTimeStampWriter(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := newTimeStampWriter(buf)

	writer.Write([]byte("foo"))
	// The first received bytes get a timestamp prepended
	require.Regexp(t, `^\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\] foo$`, buf)

	buf.Reset()
	writer.Write([]byte("\nbar\n"))
	// A timestamp should be inserted before the first char of the next line
	require.Regexp(t, "^\n"+`\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\] bar` + "\n$", buf)
}
