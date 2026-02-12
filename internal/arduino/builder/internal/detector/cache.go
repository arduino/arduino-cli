// This file is part of arduino-cli.
//
// Copyright 2024 ARDUINO SA (http://www.arduino.cc/)
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
	"fmt"

	"github.com/arduino/arduino-cli/internal/arduino/builder/internal/runner"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
)

type detectorCache struct {
	curr    int
	entries []*detectorCacheEntry
}

type detectorCacheEntry struct {
	AddedIncludePath *paths.Path  `json:"added_include_path,omitempty"`
	Compile          *sourceFile  `json:"compile,omitempty"`
	CompileTask      *runner.Task `json:"compile_task,omitempty"`
	MissingIncludeH  *string      `json:"missing_include_h,omitempty"`
}

func (e *detectorCacheEntry) String() string {
	if e.AddedIncludePath != nil {
		return "Added include path: " + e.AddedIncludePath.String()
	}
	if e.CompileTask != nil {
		return "Compiling: " + e.Compile.String() + " / " + e.CompileTask.String()
	}
	if e.MissingIncludeH != nil {
		if *e.MissingIncludeH == "" {
			return "No missing include files detected"
		}
		return "Missing include file: " + *e.MissingIncludeH
	}
	return "No operation"
}

func (e *detectorCacheEntry) LogMsg() string {
	if e.CompileTask == nil {
		return e.String()
	}
	return "Compiling: " + e.Compile.SourcePath.String()
}

func (e *detectorCacheEntry) Equals(entry *detectorCacheEntry) bool {
	return e.String() == entry.String()
}

func newDetectorCache() *detectorCache {
	return &detectorCache{}
}

func (c *detectorCache) String() string {
	res := ""
	for _, entry := range c.entries {
		res += fmt.Sprintln(entry)
	}
	return res
}

// Load reads a saved cache from the given file.
// If the file do not exists, it does nothing.
func (c *detectorCache) Load(cacheFile *paths.Path) error {
	if exist, err := cacheFile.ExistCheck(); err != nil {
		return err
	} else if !exist {
		return nil
	}
	data, err := cacheFile.ReadFile()
	if err != nil {
		return err
	}
	var entries []*detectorCacheEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return err
	}
	c.curr = 0
	c.entries = entries
	return nil
}

// ExpectCompile adds a compile entry to the cache and checks if it matches the next expected entry.
func (c *detectorCache) ExpectCompile(sourceFile sourceFile, compileTask *runner.Task) {
	c.Expect(&detectorCacheEntry{
		Compile:     &sourceFile,
		CompileTask: compileTask,
	})
}

// Expect adds an entry to the cache and checks if it matches the next expected entry.
func (c *detectorCache) Expect(entry *detectorCacheEntry) {
	if c.curr < len(c.entries) {
		if c.entries[c.curr].Equals(entry) {
			// Cache hit, move to the next entry
			c.curr++
			logrus.Tracef("[LD] CACHE: HIT %s", entry.LogMsg())
			return
		}
		// Cache mismatch, invalidate and cut the remainder of the cache
		logrus.Tracef("[LD] CACHE: INVALIDATE %s", entry.LogMsg())
		logrus.Tracef("[LD]              (was %s)", c.entries[c.curr])
		c.entries = c.entries[:c.curr]
	} else {
		logrus.Tracef("[LD] CACHE: MISSING %s", entry.LogMsg())
	}
	c.curr++
	c.entries = append(c.entries, entry)
}

// Peek returns the next cache entry to be expected or nil if the cache is fully consumed.
func (c *detectorCache) Peek() *detectorCacheEntry {
	if c.curr < len(c.entries) {
		return c.entries[c.curr]
	}
	return nil
}

// EntriesAhead returns the entries that are ahead of the current cache position.
func (c *detectorCache) EntriesAhead() []*detectorCacheEntry {
	if c.curr < len(c.entries) {
		return c.entries[c.curr:]
	}
	return nil
}

// Save writes the current cache to the given file.
func (c *detectorCache) Save(cacheFile *paths.Path) error {
	// Cut off the cache if it is not fully consumed
	c.entries = c.entries[:c.curr]

	data, err := json.MarshalIndent(c.entries, "", "  ")
	if err != nil {
		return err
	}
	return cacheFile.WriteFile(data)
}
