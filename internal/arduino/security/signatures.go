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
	"bytes"
	"embed"
	"errors"
	"io"
	"os"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	pgperrors "github.com/ProtonMail/go-crypto/openpgp/errors"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
)

//go:embed keys/*
var keys embed.FS

// VerifyArduinoDetachedSignature checks that the detached GPG signature (in the
// signaturePath file) matches the given targetPath file and is an authentic
// signature from the bundled trusted keychain. If any of the above conditions
// fails this function returns false. The PGP entity in the trusted keychain that
// produced the signature is returned too. This function use the default and bundled
// arduino_public.gpg.key
func VerifyArduinoDetachedSignature(targetPath *paths.Path, signaturePath *paths.Path) (bool, *openpgp.Entity, error) {
	arduinoKeyringFile, err := keys.Open("keys/arduino_public.gpg.key")
	if err != nil {
		panic("could not find bundled signature keys")
	}
	defer arduinoKeyringFile.Close()
	return VerifySignature(targetPath, signaturePath, arduinoKeyringFile)
}

// VerifyDetachedSignature checks that the detached GPG signature (in the
// signaturePath file) matches the given targetPath file and is an authentic
// signature from the bundled trusted keychain. The keyPath is the path of the public key used.
// This function allows to specify the path of the key to use.
// If any of the above conditions fails this function returns false.
// The PGP entity in the trusted keychain that produced the signature is returned too.
func VerifyDetachedSignature(targetPath *paths.Path, signaturePath *paths.Path, keyPath *paths.Path) (bool, *openpgp.Entity, error) {
	arduinoKeyringFile, err := os.Open(keyPath.String())
	if err != nil {
		panic("could not open signature keys")
	}
	defer arduinoKeyringFile.Close()
	return VerifySignature(targetPath, signaturePath, arduinoKeyringFile)
}

// VerifySignature checks that the detached GPG signature (in the
// signaturePath file) matches the given targetPath file and is an authentic
// signature. This function allows to pass an io.Reader to read the custom key.
//
// If any of the above conditions fails this function returns false.
//
// The PGP entity in the trusted keychain that produced the signature is returned too.
func VerifySignature(targetPath *paths.Path, signaturePath *paths.Path, arduinoKeyringFile io.Reader) (bool, *openpgp.Entity, error) {
	keyRing, err := openpgp.ReadKeyRing(arduinoKeyringFile)
	if err != nil {
		return false, nil, errors.New(i18n.Tr("retrieving Arduino public keys: %s", err))
	}
	target, err := targetPath.ReadFile()
	if err != nil {
		return false, nil, errors.New(i18n.Tr("opening target file: %s", err))
	}
	signature, err := signaturePath.ReadFile()
	if err != nil {
		return false, nil, errors.New(i18n.Tr("opening signature file: %s", err))
	}
	signer, err := openpgp.CheckDetachedSignature(keyRing, bytes.NewBuffer(target), bytes.NewBuffer(signature), nil)

	// Some users reported spurious "expired signature" errors. After some investigation
	// we found that all of them had a wrong system date set on their machine, with
	// a date set in the past.
	// Even if the error says that the signature is "expired", it's actually a
	// signature that is not yet valid (it will be in the future).
	// Since we could not trust the system clock, we recheck the signature with a date set
	// in the future, so we may avoid to display a difficult to understand error to the user.
	year2100 := time.Date(2100, 0, 0, 0, 0, 0, 0, time.UTC)
	if errors.Is(err, pgperrors.ErrSignatureExpired) && time.Now().Before(year2100) {
		logrus.Warn("Ignoring expired signature")
		signer, err = openpgp.CheckDetachedSignature(keyRing, bytes.NewBuffer(target), bytes.NewBuffer(signature), &packet.Config{
			Time: func() time.Time { return year2100 },
		})
	}

	return (signer != nil && err == nil), signer, err
}
