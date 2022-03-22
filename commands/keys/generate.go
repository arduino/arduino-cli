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

package keys

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
)

var tr = i18n.Tr

// Generate creates a new set of security keys via gRPC
func Generate(ctx context.Context, req *rpc.KeysGenerateRequest) (*rpc.KeysGenerateResponse, error) {
	// check if the keychain is passed as argument
	keysKeychainDir := paths.New(req.GetKeysKeychain())
	if keysKeychainDir == nil {
		keysKeychainDir = paths.New(".")
		keysKeychainDir, _ = keysKeychainDir.Abs()
	}
	pathExists, err := keysKeychainDir.ExistCheck()
	if !pathExists {
		if err = keysKeychainDir.MkdirAll(); err != nil {
			return nil, &arduino.PermissionDeniedError{Message: tr("Cannot create directory"), Cause: err}
		}
	}
	if err != nil {
		return nil, &arduino.PermissionDeniedError{Message: tr("Cannot verify if the directory %s exists", keysKeychainDir), Cause: err}
	}

	// check the number of keynames passed
	if len(req.GetKeyName()) == 0 {
		return nil, &arduino.InvalidArgumentError{Message: tr("Wrong number of key names specified, please use at least one")}
	}

	// check if the algorithm has been specified, set the default if not
	algorithmType := req.GetAlgorithmType()
	if algorithmType == "" {
		algorithmType = "ecdsa-p256"
	}

	// do the actual key generation
	for _, key := range req.GetKeyName() {
		// build the path where to save the security key
		privKeyPath := keysKeychainDir.Join("priv_" + key)
		pubKeyPath := keysKeychainDir.Join("pub_" + key)
		err = doKeyGen(privKeyPath, pubKeyPath, algorithmType)
		if err != nil {
			return nil, &arduino.FileCreationFailedError{Message: tr("Cannot create file"), Cause: err}
		}
	}
	return &rpc.KeysGenerateResponse{KeysKeychain: keysKeychainDir.String()}, nil
}

// adapted from https://git.furworks.de/Zephyr/mcuboot/src/commit/3869e760901a27adff47ccaea803a42f1b0169c0/imgtool/imgtool.go#L69
// doKeyGen will take the paths of the public and private keys to write and will generate keys according to keyType
func doKeyGen(privKeyPath, pubKeyPath *paths.Path, keyType string) (err error) {
	var priv509, pubAsn1 []byte
	var privPemType, pubPemType string
	switch keyType {
	case "ecdsa-p256":
		priv509, pubAsn1, err = genEcdsaP256()
		privPemType = "PRIVATE KEY"
		pubPemType = "PUBLIC KEY"
	// support for multiple algorithms can be added there
	default:
		err = errors.New(tr("Unsupported algorithm: %s", keyType))
	}

	if err != nil {
		return err
	}
	keysKeychainDir := privKeyPath.Parent()

	// create the private key file
	if privKeyPath.Exist() {
		return errors.New(tr("File already exists: %s", privKeyPath))
	}
	privKeyFile, err := privKeyPath.Create()
	if err != nil {
		return err
	}
	defer privKeyFile.Close()

	// create the public key file
	if pubKeyPath.Exist() {
		return errors.New(tr("File already exists: %s", pubKeyPath))
	}
	pubKeyFile, err := pubKeyPath.Create()
	if err != nil {
		return err
	}
	defer pubKeyFile.Close()

	// create the private header files
	privHeader := strings.TrimSuffix(privKeyPath.Base(), privKeyPath.Ext())
	privHeaderPath := keysKeychainDir.Join(privHeader + ".h")
	if privHeaderPath.Exist() {
		return errors.New(tr("File already exists: %s", privHeaderPath))
	}
	privHeaderFile, err := privHeaderPath.Create()
	if err != nil {
		return err
	}
	defer privHeaderFile.Close()

	// create the public header files
	pubHeader := strings.TrimSuffix(pubKeyPath.Base(), pubKeyPath.Ext())
	pubHeaderPath := keysKeychainDir.Join(pubHeader + ".h")
	if pubHeaderPath.Exist() {
		return errors.New(tr("File already exists: %s", pubHeaderPath))
	}
	pubHeaderFile, err := pubHeaderPath.Create()
	if err != nil {
		return err
	}
	defer pubHeaderFile.Close()

	privBlock := pem.Block{
		Type:  privPemType,
		Bytes: priv509,
	}
	err = pem.Encode(privKeyFile, &privBlock)
	if err != nil {
		return err
	}
	err = genCFile(privHeaderFile, priv509, "priv")
	if err != nil {
		return err
	}

	pubBlock := pem.Block{
		Type:  pubPemType,
		Bytes: pubAsn1,
	}
	err = pem.Encode(pubKeyFile, &pubBlock)
	if err != nil {
		return err
	}
	err = genCFile(pubHeaderFile, pubAsn1, "pub")
	if err != nil {
		return err
	}
	return nil
}

// genEcdsaP256 will generate private and public ecdsap256 keypair and return them
// it will encode the private PKCS #8, ASN.1 DER form, and for the public will use the PKIX, ASN.1 DER form
// adapted from https://git.furworks.de/Zephyr/mcuboot/src/commit/3869e760901a27adff47ccaea803a42f1b0169c0/imgtool/imgtool.go#L165
func genEcdsaP256() ([]byte, []byte, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	keyPriv, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, nil, err
	}

	pub := priv.Public()
	keyPub, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, nil, err
	}

	return keyPriv, keyPub, nil
}

// genCFile will take a file as input, the byte slice that represents the key and the type of the key.
// If will then write the file and report an error if one is found
func genCFile(file *os.File, bytes []byte, t string) (err error) {
	fileContent := fmt.Sprintf(`/* Autogenerated, do not edit */
const unsigned char rsa_%s_key[] = {
	%s
};
const unsigned int ec_%s_key_len = %d;
`, t, formatCData(bytes), t, len(bytes))
	_, err = file.WriteString(fileContent)
	return err
}

// formatCData will take the byte slice representing a key and format correctly as "C" data. It will return it as a string
// taken and adapted from https://git.furworks.de/Zephyr/mcuboot/src/commit/3869e760901a27adff47ccaea803a42f1b0169c0/imgtool/imgtool.go#L313
func formatCData(data []byte) string {
	buf := new(bytes.Buffer)
	indText := strings.Repeat("\t", 1)
	for i, b := range data {
		if i%8 == 0 {
			if i > 0 {
				fmt.Fprintf(buf, "\n%s", indText)
			}
		} else {
			fmt.Fprintf(buf, " ")
		}
		fmt.Fprintf(buf, "0x%02x,", b)
	}
	return buf.String()
}
