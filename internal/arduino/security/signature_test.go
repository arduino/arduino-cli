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

var (
	PackageIndexPath      = paths.New("testdata/package_index.json")
	PackageSignaturePath  = paths.New("testdata/package_index.json.sig")
	ModuleFWIndexPath     = paths.New("testdata/module_firmware_index.json")
	ModuleFWSignaturePath = paths.New("testdata/module_firmware_index.json.sig")
	ModuleFWIndexKey      = paths.New("testdata/module_firmware_index_public.gpg.key")
	InvalidIndexPath      = paths.New("testdata/invalid_file.json")
)

func TestVerifyArduinoDetachedSignature(t *testing.T) {
	res, signer, err := VerifyArduinoDetachedSignature(PackageIndexPath, PackageSignaturePath)
	require.NoError(t, err)
	require.NotNil(t, signer)
	require.True(t, res)
	require.Equal(t, uint64(0x7baf404c2dfab4ae), signer.PrimaryKey.KeyId)

	res, signer, err = VerifyArduinoDetachedSignature(InvalidIndexPath, PackageSignaturePath)
	require.False(t, res)
	require.Nil(t, signer)
	require.Error(t, err)
}

func TestVerifyDetachedSignature(t *testing.T) {
	res, signer, err := VerifyDetachedSignature(ModuleFWIndexPath, ModuleFWSignaturePath, ModuleFWIndexKey)
	require.NoError(t, err)
	require.NotNil(t, signer)
	require.True(t, res)
	require.Equal(t, uint64(0x82f2d7c7c5a22a73), signer.PrimaryKey.KeyId)

	res, signer, err = VerifyDetachedSignature(InvalidIndexPath, PackageSignaturePath, ModuleFWIndexKey)
	require.False(t, res)
	require.Nil(t, signer)
	require.Error(t, err)
}

func TestVerifySignature(t *testing.T) {
	arduinoKeyringFile, err := keys.Open("keys/arduino_public.gpg.key")
	if err != nil {
		panic("could not find bundled signature keys")
	}
	res, signer, err := VerifySignature(PackageIndexPath, PackageSignaturePath, arduinoKeyringFile)
	require.NoError(t, err)
	require.NotNil(t, signer)
	require.True(t, res)
	require.Equal(t, uint64(0x7baf404c2dfab4ae), signer.PrimaryKey.KeyId)

	res, signer, err = VerifySignature(InvalidIndexPath, PackageSignaturePath, arduinoKeyringFile)
	require.False(t, res)
	require.Nil(t, signer)
	require.Error(t, err)
}
