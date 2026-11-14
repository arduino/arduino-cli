// This file is part of arduino-cli.
//
// Copyright 2026 ARDUINO SA (http://www.arduino.cc/)
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

package buildhistory

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func always(*Metadata) bool { return true }

// makeArtifacts creates a fake build output containing the given content.
func makeArtifacts(t *testing.T, content string) paths.PathList {
	dir := paths.New(t.TempDir())
	hex := dir.Join("sketch.ino.hex")
	elf := dir.Join("sketch.ino.elf")
	require.NoError(t, hex.WriteFile([]byte("hex:"+content)))
	require.NoError(t, elf.WriteFile([]byte("elf:"+content)))
	return paths.PathList{hex, elf}
}

func manifest(content string) Manifest {
	return Manifest{Files: []FileHash{{Path: "/lib/a.h", SHA256: content}}}
}

func validEntries(_ *testing.T, s *Store) []*Entry {
	return s.entries()
}

func setLastUsed(t *testing.T, e *Entry, tm time.Time) {
	require.NoError(t, os.Chtimes(e.Dir().Join(lastUsedFileName).String(), tm, tm))
}

func TestAddLookupRestore(t *testing.T) {
	s := New(paths.New(t.TempDir()), 10)

	require.Nil(t, s.Lookup("key1", always))

	added, err := s.Add(Metadata{Key: "key1", Manifest: manifest("v1")}, makeArtifacts(t, "v1"))
	require.NoError(t, err)
	require.True(t, added)

	// Wrong key or invalid manifest: miss
	require.Nil(t, s.Lookup("key2", always))
	require.Nil(t, s.Lookup("key1", func(*Metadata) bool { return false }))

	e := s.Lookup("key1", always)
	require.NotNil(t, e)
	require.Equal(t, manifest("v1").Files, e.Manifest.Files)
	require.Len(t, e.Artifacts, 2)

	dest := paths.New(t.TempDir())
	require.NoError(t, dest.Join("sketch.ino.hex").WriteFile([]byte("old content")))
	require.NoError(t, e.Restore(dest))
	hex, err := dest.Join("sketch.ino.hex").ReadFile()
	require.NoError(t, err)
	require.Equal(t, "hex:v1", string(hex))
	elf, err := dest.Join("sketch.ino.elf").ReadFile()
	require.NoError(t, err)
	require.Equal(t, "elf:v1", string(elf))

	// No temporary files left in the destination
	files, err := dest.ReadDir()
	require.NoError(t, err)
	require.Len(t, files, 2)
}

func TestLookupSelectsMatchingManifest(t *testing.T) {
	s := New(paths.New(t.TempDir()), 10)
	_, err := s.Add(Metadata{Key: "key", Manifest: manifest("lib-v1")}, makeArtifacts(t, "with-lib-v1"))
	require.NoError(t, err)
	_, err = s.Add(Metadata{Key: "key", Manifest: manifest("lib-v2")}, makeArtifacts(t, "with-lib-v2"))
	require.NoError(t, err)

	for _, v := range []string{"lib-v1", "lib-v2"} {
		e := s.Lookup("key", func(md *Metadata) bool { return md.Manifest.Files[0].SHA256 == v })
		require.NotNil(t, e)
		dest := paths.New(t.TempDir())
		require.NoError(t, e.Restore(dest))
		hex, err := dest.Join("sketch.ino.hex").ReadFile()
		require.NoError(t, err)
		require.Equal(t, "hex:with-"+v, string(hex))
	}
}

func TestDuplicatesAreNotStored(t *testing.T) {
	s := New(paths.New(t.TempDir()), 10)
	added, err := s.Add(Metadata{Key: "key", Manifest: manifest("v1")}, makeArtifacts(t, "v1"))
	require.NoError(t, err)
	require.True(t, added)

	added, err = s.Add(Metadata{Key: "key", Manifest: manifest("v1")}, makeArtifacts(t, "v1"))
	require.NoError(t, err)
	require.False(t, added)
	require.Len(t, validEntries(t, s), 1)

	// Manifest order does not matter
	m := Manifest{Files: []FileHash{{Path: "/a", SHA256: "1"}, {Path: "/b", SHA256: "2"}}}
	mReversed := Manifest{Files: []FileHash{{Path: "/b", SHA256: "2"}, {Path: "/a", SHA256: "1"}}}
	added, err = s.Add(Metadata{Key: "key", Manifest: m}, makeArtifacts(t, "v2"))
	require.NoError(t, err)
	require.True(t, added)
	added, err = s.Add(Metadata{Key: "key", Manifest: mReversed}, makeArtifacts(t, "v2"))
	require.NoError(t, err)
	require.False(t, added)
	require.Len(t, validEntries(t, s), 2)
}

func TestEvictionIsLeastRecentlyUsed(t *testing.T) {
	s := New(paths.New(t.TempDir()), 2)
	now := time.Now()

	_, err := s.Add(Metadata{Key: "k1"}, makeArtifacts(t, "1"))
	require.NoError(t, err)
	setLastUsed(t, s.Lookup("k1", always), now.Add(-3*time.Hour))
	_, err = s.Add(Metadata{Key: "k2"}, makeArtifacts(t, "2"))
	require.NoError(t, err)
	setLastUsed(t, s.Lookup("k2", always), now.Add(-2*time.Hour))

	// Use k1: k2 becomes the least recently used
	s.Touch(s.Lookup("k1", always))

	_, err = s.Add(Metadata{Key: "k3"}, makeArtifacts(t, "3"))
	require.NoError(t, err)

	require.Len(t, validEntries(t, s), 2)
	require.NotNil(t, s.Lookup("k1", always))
	require.Nil(t, s.Lookup("k2", always))
	require.NotNil(t, s.Lookup("k3", always))

	// No trash left behind
	dirs, err := s.Dir().ReadDir()
	require.NoError(t, err)
	require.Len(t, dirs, 2)
}

func TestEvictionWhenMaxEntriesIsLowered(t *testing.T) {
	dir := paths.New(t.TempDir())
	s := New(dir, 10)
	for i := range 5 {
		_, err := s.Add(Metadata{Key: fmt.Sprint(i)}, makeArtifacts(t, fmt.Sprint(i)))
		require.NoError(t, err)
		setLastUsed(t, s.Lookup(fmt.Sprint(i), always), time.Now().Add(time.Duration(i-10)*time.Minute))
	}
	require.Len(t, validEntries(t, s), 5)

	s = New(dir, 2)
	_, err := s.Add(Metadata{Key: "new"}, makeArtifacts(t, "new"))
	require.NoError(t, err)
	require.Len(t, validEntries(t, s), 2)
	require.NotNil(t, s.Lookup("new", always))
	require.NotNil(t, s.Lookup("4", always))
}

func TestZeroMaxEntriesStoresNothing(t *testing.T) {
	s := New(paths.New(t.TempDir()), 0)
	added, err := s.Add(Metadata{Key: "k"}, makeArtifacts(t, "1"))
	require.NoError(t, err)
	require.False(t, added)
	dirs, err := s.Dir().ReadDir()
	require.NoError(t, err)
	require.Empty(t, dirs)
	require.Nil(t, s.Lookup("k", always))
}

func TestIncompleteEntriesAreIgnored(t *testing.T) {
	s := New(paths.New(t.TempDir()), 10)
	_, err := s.Add(Metadata{Key: "key"}, makeArtifacts(t, "good"))
	require.NoError(t, err)
	good := s.Lookup("key", always)
	require.NotNil(t, good)

	// Interrupted staging directory with a complete metadata file
	staging := s.Dir().Join(stagingPrefix + "abc-123")
	require.NoError(t, good.Dir().CopyDirTo(staging))
	// Entry without metadata (should never happen thanks to the atomic rename)
	require.NoError(t, s.Dir().Join("nometadata", artifactsDirName).MkdirAll())
	// Truncated metadata
	require.NoError(t, s.Dir().Join("truncated").MkdirAll())
	require.NoError(t, s.Dir().Join("truncated", metadataFileName).WriteFile([]byte(`{"format_version": 1, "key": "key"`)))
	// Entry from a future format
	require.NoError(t, s.Dir().Join("future").MkdirAll())
	require.NoError(t, s.Dir().Join("future", metadataFileName).WriteFile([]byte(`{"format_version": 999, "key": "key"}`)))

	entries := validEntries(t, s)
	require.Len(t, entries, 1)
	require.Equal(t, good.Dir().String(), entries[0].Dir().String())

	// Stale staging directories are removed on the next Add
	old := time.Now().Add(-2 * staleTempAge)
	require.NoError(t, os.Chtimes(staging.String(), old, old))
	recent := s.Dir().Join(stagingPrefix + "in-progress")
	require.NoError(t, recent.MkdirAll())
	_, err = s.Add(Metadata{Key: "other"}, makeArtifacts(t, "other"))
	require.NoError(t, err)
	require.False(t, staging.Exist())
	require.True(t, recent.Exist(), "a staging directory of a running process must not be removed")
}

func TestCorruptedArtifactsAreNotRestored(t *testing.T) {
	s := New(paths.New(t.TempDir()), 10)
	_, err := s.Add(Metadata{Key: "key"}, makeArtifacts(t, "v1"))
	require.NoError(t, err)
	e := s.Lookup("key", always)
	require.NotNil(t, e)

	dest := paths.New(t.TempDir())
	require.NoError(t, dest.Join("sketch.ino.hex").WriteFile([]byte("current")))

	// Same size, different content
	require.NoError(t, e.Dir().Join(artifactsDirName, "sketch.ino.hex").WriteFile([]byte("hex:XX")))
	require.Error(t, e.Restore(dest))
	// Missing artifact
	require.NoError(t, e.Dir().Join(artifactsDirName, "sketch.ino.hex").Remove())
	require.Error(t, e.Restore(dest))

	// The destination is untouched and contains no temporary files
	files, err := dest.ReadDir()
	require.NoError(t, err)
	require.Len(t, files, 1)
	data, err := dest.Join("sketch.ino.hex").ReadFile()
	require.NoError(t, err)
	require.Equal(t, "current", string(data))
}

func TestMaliciousArtifactNamesAreRejected(t *testing.T) {
	s := New(paths.New(t.TempDir()), 10)
	_, err := s.Add(Metadata{Key: "key"}, makeArtifacts(t, "v1"))
	require.NoError(t, err)
	e := s.Lookup("key", always)
	require.NotNil(t, e)
	e.Artifacts[0].Name = "../escape.hex"
	require.Error(t, e.Restore(paths.New(t.TempDir())))
}

func TestConcurrentAccess(t *testing.T) {
	dir := paths.New(t.TempDir())
	const maxEntries = 3
	const workers = 8
	const rounds = 10

	artifacts := map[string]paths.PathList{}
	for i := range 5 {
		k := fmt.Sprint(i)
		artifacts[k] = makeArtifacts(t, strings.Repeat(k, 100000))
	}

	var wg sync.WaitGroup
	errs := make(chan error, workers*rounds*2)
	for w := range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Each worker uses its own Store, as different processes would
			s := New(dir, maxEntries)
			for r := range rounds {
				k := fmt.Sprint((w + r) % 5)
				if _, err := s.Add(Metadata{Key: k}, artifacts[k]); err != nil {
					errs <- err
				}
				// Restoring may fail if the entry is concurrently evicted, but
				// must never produce the wrong content.
				if e := s.Lookup(k, always); e != nil {
					dest := paths.New(t.TempDir())
					if err := e.Restore(dest); err == nil {
						data, err := dest.Join("sketch.ino.hex").ReadFile()
						if err != nil || string(data) != "hex:"+strings.Repeat(k, 100000) {
							errs <- fmt.Errorf("wrong content restored for %s", k)
						}
					}
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}

	s := New(dir, maxEntries)
	entries := validEntries(t, s)
	require.NotEmpty(t, entries)
	require.LessOrEqual(t, len(entries), maxEntries)
	for _, e := range entries {
		require.NoError(t, e.Restore(paths.New(t.TempDir())))
	}
	dirs, err := dir.ReadDir()
	require.NoError(t, err)
	for _, d := range dirs {
		require.False(t, strings.HasPrefix(d.Base(), stagingPrefix), "leftover staging dir %s", d)
		require.False(t, strings.HasPrefix(d.Base(), trashPrefix), "leftover trash dir %s", d)
	}
}
