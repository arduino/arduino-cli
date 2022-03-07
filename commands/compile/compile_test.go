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

package compile

import (
	"testing"

	properties "github.com/arduino/go-properties-orderedmap"
	"github.com/stretchr/testify/require"
)

func TestReplaceSecurityKeys(t *testing.T) {
	propsWithDefaultKeys := properties.NewFromHashmap(map[string]string{
		"tools.toolname.keys.path":    "/default-keys-path",
		"tools.toolname.sign.name":    "default-signing-key.pem",
		"tools.toolname.encrypt.name": "default-encrypt-key.pem",
	})
	newKeysPath := "/new-keys-path"
	newSignKeyName := "new-signing-key.pem"
	newEncryptKeyName := "new-encrypt-key.pem"
	goldProps := properties.NewFromHashmap(map[string]string{
		"tools.toolname.keys.path":    newKeysPath,
		"tools.toolname.sign.name":    newSignKeyName,
		"tools.toolname.encrypt.name": newEncryptKeyName,
	})

	ReplaceSecurityKeys(propsWithDefaultKeys, newKeysPath, newSignKeyName, newEncryptKeyName)
	require.True(t, goldProps.Equals(propsWithDefaultKeys))
}

func TestReplaceSecurityKeysEmpty(t *testing.T) {
	propsWithNoKeys := properties.NewFromHashmap(map[string]string{})
	goldProps := properties.NewFromHashmap(map[string]string{})
	newKeysPath := "/new-keys-path"
	newSignKeyName := "new-signing-key.pem"
	newEncryptKeyName := "new-encrypt-key.pem"

	// No error should be returned since the properties map is empty
	ReplaceSecurityKeys(propsWithNoKeys, newKeysPath, newSignKeyName, newEncryptKeyName)
	require.True(t, goldProps.Equals(propsWithNoKeys))
}

func TestReplaceSecurityKeysNothingToReplace(t *testing.T) {
	propsWithDifferentKeys := properties.NewFromHashmap(map[string]string{
		"tools.openocd.path":        "{runtime.tools.openocd.path}",
		"tools.openocd.cmd":         "bin/openocd",
		"tools.openocd.cmd.windows": "bin/openocd.exe",
	})
	goldProps := propsWithDifferentKeys.Clone()
	newKeysPath := "/new-keys-path"
	newSignKeyName := "new-signing-key.pem"
	newEncryptKeyName := "new-encrypt-key.pem"

	// No error should be returned since there are no keys in the properties map
	ReplaceSecurityKeys(propsWithDifferentKeys, newKeysPath, newSignKeyName, newEncryptKeyName)
	require.True(t, goldProps.Equals(propsWithDifferentKeys))
}
