/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017 BCMI LABS SA (http://www.arduino.cc/)
 */

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
