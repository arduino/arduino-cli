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

package feedback

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOutputSelection(t *testing.T) {
	reset()

	myErr := new(bytes.Buffer)
	myOut := new(bytes.Buffer)
	SetOut(myOut)
	SetErr(myErr)
	SetFormat(Text)

	// Could not change output stream after format has been set
	require.Panics(t, func() { SetOut(nil) })
	require.Panics(t, func() { SetErr(nil) })

	// Coule not change output format twice
	require.Panics(t, func() { SetFormat(JSON) })

	Print("Hello")
	require.Equal(t, myOut.String(), "Hello\n")
}

func TestJSONOutputStream(t *testing.T) {
	reset()

	require.Panics(t, func() { OutputStreams() })

	SetFormat(JSON)
	stdout, stderr, res := OutputStreams()
	fmt.Fprint(stdout, "Hello")
	fmt.Fprint(stderr, "Hello ERR")

	d, err := json.Marshal(res())
	require.NoError(t, err)
	require.JSONEq(t, `{"stdout":"Hello","stderr":"Hello ERR"}`, string(d))

	stdout.Write([]byte{0xc2, 'A'}) // Invaid UTF-8

	d, err = json.Marshal(res())
	require.NoError(t, err)
	require.JSONEq(t, string(d), `{"stdout":"Hello\ufffdA","stderr":"Hello ERR"}`)
}

func TestJsonOutputOnCustomStreams(t *testing.T) {
	reset()

	myErr := new(bytes.Buffer)
	myOut := new(bytes.Buffer)
	SetOut(myOut)
	SetErr(myErr)
	SetFormat(JSON)

	// Could not change output stream after format has been set
	require.Panics(t, func() { SetOut(nil) })
	require.Panics(t, func() { SetErr(nil) })
	// Could not change output format twice
	require.Panics(t, func() { SetFormat(JSON) })

	Print("Hello") // Output interactive data

	require.Equal(t, "", myOut.String())
	require.Equal(t, "", myErr.String())
	require.Equal(t, "Hello\n", bufferOut.String())

	PrintResult(&testResult{Success: true})

	require.JSONEq(t, myOut.String(), `{ "success": true }`)
	require.Equal(t, myErr.String(), "")
	myOut.Reset()

	_, _, res := OutputStreams()
	PrintResult(&testResult{Success: false, Output: res()})

	require.JSONEq(t, `
{
  "success": false,
  "output": {
    "stdout": "Hello\n",
    "stderr": ""
  }
}`, myOut.String())
	require.Equal(t, myErr.String(), "")
}

type testResult struct {
	Success bool                 `json:"success"`
	Output  *OutputStreamsResult `json:"output,omitempty"`
}

func (r *testResult) Data() interface{} {
	return r
}

func (r *testResult) String() string {
	if r.Success {
		return "Success"
	}
	return "Failure"
}
