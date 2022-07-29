// This file is part of arduino-cli.
//
// Copyright 2022 ARDUINO SA (http://www.arduino.cc/)
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

package integrationtest

import (
	"encoding/json"
	"testing"

	"github.com/itchyny/gojq"
	"github.com/stretchr/testify/require"
)

// JQQuery performs a test on a given json output. A jq-like query is performed
// on the given jsonData and the result is compared with the expected output.
// If the output doesn't match the test fails. If msgAndArgs are provided they
// will be used to explain the error.
func JQQuery(t *testing.T, jsonData []byte, jqQuery string, expected interface{}, msgAndArgs ...interface{}) {
	var data interface{}
	require.NoError(t, json.Unmarshal(jsonData, &data))
	q, err := gojq.Parse(jqQuery)
	require.NoError(t, err)
	i := q.Run(data)
	v, ok := i.Next()
	require.True(t, ok)
	require.IsType(t, expected, v)
	require.Equal(t, expected, v, msgAndArgs...)
}
