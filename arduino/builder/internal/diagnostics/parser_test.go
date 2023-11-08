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

package diagnostics

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	t.Run("Generic001", func(t *testing.T) { runParserTest(t, "test001.txt") })
	t.Run("Generic002", func(t *testing.T) { runParserTest(t, "test002.txt") })
	t.Run("Generic003", func(t *testing.T) { runParserTest(t, "test003.txt") })
	t.Run("Generic004", func(t *testing.T) { runParserTest(t, "test004.txt") })
}

func runParserTest(t *testing.T, testFile string) {
	testData, err := paths.New("testdata", "compiler_outputs", testFile).ReadFile()
	require.NoError(t, err)
	// The first line contains the compiler arguments
	idx := bytes.Index(testData, []byte("\n"))
	require.NotEqual(t, -1, idx)
	args := strings.Split(string(testData[0:idx]), " ")
	// The remainder of the file is the compiler output
	data := testData[idx:]

	// Run compiler detection and parse compiler output
	detectedCompiler := DetectCompilerFromCommandLine(args, true)
	require.NotNil(t, detectedCompiler)
	diags, err := ParseCompilerOutput(detectedCompiler, data)
	require.NoError(t, err)

	// Check if the parsed data match the expected output
	output, err := json.MarshalIndent(diags, "", "  ")
	require.NoError(t, err)
	golden, err := paths.New("testdata", "compiler_outputs", testFile+".json").ReadFile()
	require.NoError(t, err)
	require.Equal(t, string(golden), string(output))
}
