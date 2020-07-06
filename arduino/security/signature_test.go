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

package security

import (
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestSignatureVerification(t *testing.T) {
	res, signer, err := VerifyArduinoDetachedSignature(paths.New("testdata/package_index.json"), paths.New("testdata/package_index.json.sig"))
	require.NoError(t, err)
	require.NotNil(t, signer)
	require.True(t, res)
	require.Equal(t, uint64(0x7baf404c2dfab4ae), signer.PrimaryKey.KeyId)

	res, signer, err = VerifyArduinoDetachedSignature(paths.New("testdata/invalid_file.json"), paths.New("testdata/package_index.json.sig"))
	require.False(t, res)
	require.Nil(t, signer)
	require.Error(t, err)
}
