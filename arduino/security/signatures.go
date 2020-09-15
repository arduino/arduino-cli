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
	"fmt"

	"github.com/arduino/go-paths-helper"
	rice "github.com/cmaglie/go.rice"
	"golang.org/x/crypto/openpgp"
)

// VerifyArduinoDetachedSignature checks that the detached GPG signature (in the
// signaturePath file) matches the given targetPath file and is an authentic
// signature from the bundled trusted keychain. If any of the above conditions
// fails this function returns false. The PGP entity in the trusted keychain that
// produced the signature is returned too.
func VerifyArduinoDetachedSignature(targetPath *paths.Path, signaturePath *paths.Path) (bool, *openpgp.Entity, error) {
	keysBox, err := rice.FindBox("keys")
	if err != nil {
		panic("could not find bundled signature keys")
	}
	arduinoKeyringFile, err := keysBox.Open("arduino_public.gpg.key")
	if err != nil {
		panic("could not find bundled signature keys")
	}
	keyRing, err := openpgp.ReadKeyRing(arduinoKeyringFile)
	if err != nil {
		return false, nil, fmt.Errorf("retrieving Arduino public keys: %s", err)
	}

	target, err := targetPath.Open()
	if err != nil {
		return false, nil, fmt.Errorf("opening target file: %s", err)
	}
	defer target.Close()
	signature, err := signaturePath.Open()
	if err != nil {
		return false, nil, fmt.Errorf("opening signature file: %s", err)
	}
	defer signature.Close()
	signer, err := openpgp.CheckDetachedSignature(keyRing, target, signature)
	return (signer != nil && err == nil), signer, err
}
