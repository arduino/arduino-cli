package checksums

import (
	"bytes"
	"crypto"
	"encoding/hex"
	"hash"
	"io"
	"os"
	"strings"

	"github.com/bcmi-labs/arduino-cli/common/releases"
)

func getHashAlgoAndComponent(checksum string) (hash.Hash, []byte) {
	components := strings.SplitN(checksum, ":", 2)
	hashAlgo := components[0]
	hashMid, err := hex.DecodeString(components[1])
	if err != nil {
		return nil, nil
	}

	hash := []byte(hashMid)
	switch hashAlgo {
	case "SHA-256":
		return crypto.SHA256.New(), hash
	case "SHA1":
		return crypto.SHA1.New(), hash
	case "MD5":
		return crypto.MD5.New(), hash
	default:
		return nil, nil
	}
}

// Match checks the checksum of a Release archive, in compliance with
// What Checksum is expected.
func Match(r releases.Release) bool {
	hash, content := getHashAlgoAndComponent(r.ExpectedChecksum())
	filePath, err := r.ArchivePath()
	if err != nil {
		return false
	}

	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()
	io.Copy(hash, file)
	return bytes.Compare(hash.Sum(nil), content) == 0
}
