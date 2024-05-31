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

package detector

import (
	"encoding/json"
	"time"

	"github.com/arduino/go-paths-helper"
)

type includeCache struct {
	// Are the cache contents valid so far?
	valid bool
	// Index into entries of the next entry to be processed. Unused
	// when the cache is invalid.
	next    int
	entries []*includeCacheEntry
}

// Next Return the next cache entry. Should only be called when the cache is
// valid and a next entry is available (the latter can be checked with
// ExpectFile). Does not advance the cache.
func (cache *includeCache) Next() *includeCacheEntry {
	return cache.entries[cache.next]
}

// ExpectFile check that the next cache entry is about the given file. If it is
// not, or no entry is available, the cache is invalidated. Does not
// advance the cache.
func (cache *includeCache) ExpectFile(sourcefile *paths.Path) {
	if cache.valid && (cache.next >= len(cache.entries) || !cache.Next().Sourcefile.EqualsTo(sourcefile)) {
		cache.valid = false
		cache.entries = cache.entries[:cache.next]
	}
}

// ExpectEntry check that the next entry matches the given values. If so, advance
// the cache. If not, the cache is invalidated. If the cache is
// invalidated, or was already invalid, an entry with the given values
// is appended.
func (cache *includeCache) ExpectEntry(sourcefile *paths.Path, include string, librarypath *paths.Path) {
	entry := &includeCacheEntry{Sourcefile: sourcefile, Include: include, Includepath: librarypath}
	if cache.valid {
		if cache.next < len(cache.entries) && cache.Next().Equals(entry) {
			cache.next++
		} else {
			cache.valid = false
			cache.entries = cache.entries[:cache.next]
		}
	}

	if !cache.valid {
		cache.entries = append(cache.entries, entry)
	}
}

// ExpectEnd check that the cache is completely consumed. If not, the cache is
// invalidated.
func (cache *includeCache) ExpectEnd() {
	if cache.valid && cache.next < len(cache.entries) {
		cache.valid = false
		cache.entries = cache.entries[:cache.next]
	}
}

// Read the cache from the given file
func readCache(path *paths.Path) *includeCache {
	bytes, err := path.ReadFile()
	if err != nil {
		// Return an empty, invalid cache
		return &includeCache{}
	}
	result := &includeCache{}
	err = json.Unmarshal(bytes, &result.entries)
	if err != nil {
		// Return an empty, invalid cache
		return &includeCache{}
	}
	result.valid = true
	return result
}

// Write the given cache to the given file if it is invalidated. If the
// cache is still valid, just update the timestamps of the file.
func (cache *includeCache) write(path *paths.Path) error {
	// If the cache was still valid all the way, just touch its file
	// (in case any source file changed without influencing the
	// includes). If it was invalidated, overwrite the cache with
	// the new contents.
	if cache.valid {
		path.Chtimes(time.Now(), time.Now())
	} else {
		bytes, err := json.MarshalIndent(cache.entries, "", "  ")
		if err != nil {
			return err
		}
		err = path.WriteFile(bytes)
		if err != nil {
			return err
		}
	}
	return nil
}
