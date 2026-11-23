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
	"slices"
	"strings"

	"github.com/arduino/arduino-cli/internal/arduino/builder/internal/runner"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
)

type detectorCache struct {
	curr    int
	entries []*detectorCacheEntry
	// sketchDirectives are the preprocessor directives the merged sketch held when the entries were recorded.
	sketchDirectives []string
}

type detectorCacheFile struct {
	SketchDirectives []string              `json:"sketch_directives"`
	Entries          []*detectorCacheEntry `json:"entries"`
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
	var res strings.Builder
	for _, entry := range c.entries {
		fmt.Fprintln(&res, entry)
	}
	return res.String()
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
	var file detectorCacheFile
	if err := json.Unmarshal(data, &file); err != nil {
		// Caches written before the sketch directives were recorded hold the bare entries
		if json.Unmarshal(data, &file.Entries) != nil {
			return err
		}
	}
	c.curr = 0
	c.entries = file.Entries
	c.sketchDirectives = file.SketchDirectives
	return nil
}

// SketchDirectivesUnchanged reports whether the merged sketch holds the same preprocessor
// directives as when the cache was recorded.
func (c *detectorCache) SketchDirectivesUnchanged(directives []string) bool {
	return c.sketchDirectives != nil && slices.Equal(c.sketchDirectives, directives)
}

// SetSketchDirectives records the preprocessor directives of the merged sketch for the next Save.
func (c *detectorCache) SetSketchDirectives(directives []string) {
	c.sketchDirectives = directives
}

// LastPass returns the task of the last recorded pass of the given unit, with the include paths
// added before it and, among them, the ones its earlier passes found. The task is nil when the
// unit has no recorded pass or its last one failed.
func (c *detectorCache) LastPass(unit sourceFile) (includes, found paths.PathList, pass *runner.Task) {
	var added paths.PathList
	first := -1
	for i, entry := range c.entries {
		if entry.AddedIncludePath != nil {
			added = append(added, entry.AddedIncludePath)
		}
		if entry.CompileTask == nil || !entry.Compile.Equals(unit) {
			continue
		}
		if first < 0 {
			first = len(added)
		}
		pass, includes, found = nil, nil, nil
		if i+1 < len(c.entries) {
			if missing := c.entries[i+1].MissingIncludeH; missing != nil && *missing == "" {
				pass, includes, found = entry.CompileTask, added.Clone(), slices.Clone(added[first:])
			}
		}
	}
	return includes, found, pass
}

// ExpectCompile adds a compile entry to the cache and checks if it matches the next expected entry.
func (c *detectorCache) ExpectCompile(sourceFile sourceFile, compileTask *runner.Task) {
	c.expect(&detectorCacheEntry{
		Compile:     &sourceFile,
		CompileTask: compileTask,
	})
}

// ExpectAddedIncludePath adds an added include path entry to the cache and checks if it matches the next expected entry.
func (c *detectorCache) ExpectAddedIncludePath(includePath *paths.Path) {
	c.expect(&detectorCacheEntry{
		AddedIncludePath: includePath,
	})
}

// ExpectMissingIncludeH adds a missing include file entry to the cache and checks if it matches the next expected entry.
func (c *detectorCache) ExpectMissingIncludeH(missingIncludeH string) {
	c.expect(&detectorCacheEntry{
		MissingIncludeH: &missingIncludeH,
	})
}

// expect adds an entry to the cache and checks if it matches the next expected entry.
func (c *detectorCache) expect(entry *detectorCacheEntry) {
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

	data, err := json.MarshalIndent(detectorCacheFile{SketchDirectives: c.sketchDirectives, Entries: c.entries}, "", "  ")
	if err != nil {
		return err
	}
	return cacheFile.WriteFile(data)
}
