package releases

import (
	"bytes"
	"crypto"
	"encoding/hex"
	"errors"
	"hash"
	"io"
	"os"
	"strings"
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

// ChecksumMatches checks the checksum of a Release archive, in compliance with
// What Checksum is expected.
func checksumMatches(r Release) bool {
	hash, content := getHashAlgoAndComponent(r.ExpectedChecksum())
	filePath, err := ArchivePath(r)
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

// CheckLocalArchive check for integrity of the local archive.
func checkLocalArchive(release Release) error {
	archivePath, err := ArchivePath(release)
	if err != nil {
		return err
	}
	stats, err := os.Stat(archivePath)
	if os.IsNotExist(err) {
		return errors.New("Archive does not exist")
	}
	if err != nil {
		return err
	}
	if stats.Size() > release.ArchiveSize() {
		return errors.New("Archive size does not match with specification of this release, assuming corruption")
	}
	if !checksumMatches(release) {
		return errors.New("Checksum does not match, assuming corruption")
	}
	return nil
}
