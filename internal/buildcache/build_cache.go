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

package buildcache

import (
	"time"

	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
)

const (
	createDirErrCode = 1
	fileWriteErrCode = 2
)

type cacheError struct {
	Code       int
	wrappedErr error
}

func (e cacheError) Error() string {
	return e.wrappedErr.Error()
}

func (e cacheError) Unwrap() error {
	return e.wrappedErr
}

func (e cacheError) Is(target error) bool {
	te, ok := target.(cacheError)
	return ok && te.Code == e.Code
}

var (
	// CreateDirErr error occurred when creating the cache directory
	CreateDirErr = cacheError{Code: createDirErrCode}
	// FileWriteErr error occurred when writing the placeholder file
	FileWriteErr = cacheError{Code: fileWriteErrCode}
)

const lastUsedFileName = ".last-used"

// BuildCache represents a cache of built files (sketches and cores), it's designed
// to work on directories. Given a directory as "base" it handles direct subdirectories as
// keys
type BuildCache struct {
	baseDir *paths.Path
}

// GetOrCreate retrieves or creates the cache directory at the given path
// If the cache already exists the lifetime of the cache is extended.
func (bc *BuildCache) GetOrCreate(key string) (*paths.Path, error) {
	keyDir := bc.baseDir.Join(key)
	if err := keyDir.MkdirAll(); err != nil {
		return nil, cacheError{createDirErrCode, err}
	}

	if err := keyDir.Join(lastUsedFileName).WriteFile([]byte{}); err != nil {
		return nil, cacheError{fileWriteErrCode, err}
	}
	return keyDir, nil
}

// Purge removes all cache directories within baseDir that have expired
// To know how long ago a directory has been last used
// it checks into the .last-used file.
func (bc *BuildCache) Purge(ttl time.Duration) {
	files, err := bc.baseDir.ReadDir()
	if err != nil {
		return
	}
	for _, file := range files {
		if file.IsDir() {
			removeIfExpired(file, ttl)
		}
	}
}

// New instantiates a build cache
func New(baseDir *paths.Path) *BuildCache {
	return &BuildCache{baseDir}
}

func removeIfExpired(dir *paths.Path, ttl time.Duration) {
	fileInfo, err := dir.Join(lastUsedFileName).Stat()
	if err != nil {
		return
	}
	lifeExpectancy := ttl - time.Since(fileInfo.ModTime())
	if lifeExpectancy > 0 {
		return
	}
	logrus.Tracef(`Purging cache directory "%s". Expired by %s`, dir, lifeExpectancy.Abs())
	err = dir.RemoveAll()
	if err != nil {
		logrus.Tracef(`Error while pruning cache directory "%s": %s`, dir, err)
	}
}
