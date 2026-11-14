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

// Package buildhistory implements a persistent store of the artifacts of
// previous successful builds, so that an identical build can be restored
// instead of being compiled again.
//
// The store is a directory containing one sub-directory per entry. Every entry
// is fully written in a hidden staging directory and then published with an
// atomic rename, so that interrupted processes can never leave a partially
// written entry behind and concurrent processes never need a lock: a reader
// either sees a complete entry or no entry at all. Entries are removed by
// renaming them to a hidden trash directory before deleting them, for the same
// reason. Artifact checksums are verified again when an entry is restored.
//
// The package does not know anything about how a build is fingerprinted: the
// caller provides an opaque configuration Key, a Manifest of the input files
// the build depended on, and decides if a stored Manifest is still valid.
package buildhistory

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
)

// FormatVersion is the version of the on-disk format of the entries. Entries
// with a different format version are ignored by Lookup and eventually evicted.
const FormatVersion = 1

const (
	metadataFileName = "metadata.json"
	lastUsedFileName = ".last-used"
	artifactsDirName = "artifacts"
	stagingPrefix    = ".tmp-"
	trashPrefix      = ".trash-"
	// staleTempAge is the age after which a staging or trash directory is
	// considered abandoned by an interrupted process and is removed.
	staleTempAge = time.Hour
)

// Store is a build history store rooted in a directory.
type Store struct {
	dir        *paths.Path
	maxEntries int
}

// New returns a Store rooted in dir that retains at most maxEntries entries.
// The directory is created when the first entry is added.
func New(dir *paths.Path, maxEntries uint) *Store {
	return &Store{dir: dir, maxEntries: int(maxEntries)}
}

// Dir returns the root directory of the store.
func (s *Store) Dir() *paths.Path {
	return s.dir
}

// FileHash is the content checksum of an input file of a build.
type FileHash struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

// DirListing is a digest of the list of files contained in a directory used
// by a build (for example a library), used to detect added or removed files.
type DirListing struct {
	Path   string `json:"path"`
	Digest string `json:"digest"`
}

// Manifest lists the inputs of a build that are only known after the build
// has been completed (for example the headers included by the sources).
type Manifest struct {
	Files     []FileHash   `json:"files"`
	Libraries []DirListing `json:"libraries"`
}

// Artifact is a file produced by a build and stored in an entry.
type Artifact struct {
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

// Metadata describes an entry of the store.
type Metadata struct {
	FormatVersion int        `json:"format_version"`
	Key           string     `json:"key"`
	Manifest      Manifest   `json:"manifest"`
	Artifacts     []Artifact `json:"artifacts"`
	FQBN          string     `json:"fqbn,omitempty"`
	SketchPath    string     `json:"sketch_path,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// Entry is a valid entry of the store.
type Entry struct {
	Metadata
	dir      *paths.Path
	lastUsed time.Time
}

// Dir returns the directory containing the entry.
func (e *Entry) Dir() *paths.Path {
	return e.dir
}

// entryID returns the deterministic name of the entry directory for the given
// key and manifest: the same build is always stored in the same directory.
func entryID(key string, manifest *Manifest) (string, error) {
	m := normalizedManifest(manifest)
	data, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	h.Write([]byte(key))
	h.Write([]byte{0})
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))[:32], nil
}

func normalizedManifest(m *Manifest) *Manifest {
	res := &Manifest{
		Files:     append([]FileHash{}, m.Files...),
		Libraries: append([]DirListing{}, m.Libraries...),
	}
	sort.Slice(res.Files, func(i, j int) bool { return res.Files[i].Path < res.Files[j].Path })
	sort.Slice(res.Libraries, func(i, j int) bool { return res.Libraries[i].Path < res.Libraries[j].Path })
	return res
}

// entries returns all the valid entries of the store, most recently used first.
func (s *Store) entries() []*Entry {
	dirs, err := s.dir.ReadDir()
	if err != nil {
		return nil
	}
	dirs.FilterDirs()
	var res []*Entry
	for _, dir := range dirs {
		if strings.HasPrefix(dir.Base(), ".") {
			continue
		}
		e, err := loadEntry(dir)
		if err != nil {
			logrus.WithError(err).Tracef("Ignoring invalid build history entry %s", dir)
			continue
		}
		res = append(res, e)
	}
	sort.SliceStable(res, func(i, j int) bool { return res[i].lastUsed.After(res[j].lastUsed) })
	return res
}

func loadEntry(dir *paths.Path) (*Entry, error) {
	data, err := dir.Join(metadataFileName).ReadFile()
	if err != nil {
		return nil, err
	}
	e := &Entry{dir: dir}
	if err := json.Unmarshal(data, &e.Metadata); err != nil {
		return nil, err
	}
	if e.FormatVersion != FormatVersion {
		return nil, fmt.Errorf("unsupported format version %d", e.FormatVersion)
	}
	if st, err := dir.Join(lastUsedFileName).Stat(); err == nil {
		e.lastUsed = st.ModTime()
	} else {
		e.lastUsed = e.CreatedAt
	}
	return e, nil
}

// Lookup returns the most recently used entry having the given key for which
// isValid returns true, or nil if there is no such entry.
func (s *Store) Lookup(key string, isValid func(*Metadata) bool) *Entry {
	for _, e := range s.entries() {
		if e.Key != key {
			continue
		}
		if isValid(&e.Metadata) {
			return e
		}
	}
	return nil
}

// Touch marks the entry as recently used, so that it is evicted last.
func (s *Store) Touch(e *Entry) {
	touch(e.dir)
}

func touch(dir *paths.Path) {
	if err := dir.Join(lastUsedFileName).WriteFile([]byte{}); err != nil {
		// The entry may have been evicted concurrently: nothing to do.
		logrus.WithError(err).Tracef("Could not mark build history entry %s as used", dir)
	}
}

// Restore copies the artifacts of the entry into destDir. The checksum of every
// artifact is verified before any file in destDir is replaced; if the entry is
// corrupted, or is removed while it is being read, an error is returned and
// destDir is left untouched.
func (e *Entry) Restore(destDir *paths.Path) error {
	if len(e.Artifacts) == 0 {
		return errors.New("build history entry contains no artifacts")
	}
	if err := destDir.MkdirAll(); err != nil {
		return err
	}
	suffix, err := randomSuffix()
	if err != nil {
		return err
	}
	type staged struct{ tmp, dest *paths.Path }
	var stagedFiles []staged
	cleanup := func() {
		for _, f := range stagedFiles {
			_ = f.tmp.Remove()
		}
	}
	for _, a := range e.Artifacts {
		if !isPlainFileName(a.Name) {
			cleanup()
			return fmt.Errorf("invalid artifact name %q", a.Name)
		}
		tmp := destDir.Join(".history-" + suffix + "-" + a.Name)
		stagedFiles = append(stagedFiles, staged{tmp: tmp, dest: destDir.Join(a.Name)})
		size, sum, err := copyAndHash(e.dir.Join(artifactsDirName, a.Name), tmp)
		if err != nil {
			cleanup()
			return err
		}
		if size != a.Size || sum != a.SHA256 {
			cleanup()
			return fmt.Errorf("artifact %s of build history entry %s is corrupted", a.Name, e.dir)
		}
	}
	for i, f := range stagedFiles {
		if err := f.tmp.Rename(f.dest); err != nil {
			for _, f := range stagedFiles[i:] {
				_ = f.tmp.Remove()
			}
			return err
		}
	}
	return nil
}

// Add stores a new entry with the given metadata and artifacts (the artifacts
// listed in the metadata are ignored and recomputed). If an identical entry
// (same key and manifest) is already present, it is only marked as recently
// used and false is returned. After adding a new entry the least recently used
// entries exceeding the configured maximum are evicted.
func (s *Store) Add(md Metadata, artifacts paths.PathList) (bool, error) {
	if s.maxEntries <= 0 {
		return false, nil
	}
	if len(artifacts) == 0 {
		return false, errors.New("no artifacts to store")
	}
	id, err := entryID(md.Key, &md.Manifest)
	if err != nil {
		return false, err
	}
	finalDir := s.dir.Join(id)
	if finalDir.Join(metadataFileName).Exist() {
		touch(finalDir)
		return false, nil
	}
	if err := s.dir.MkdirAll(); err != nil {
		return false, err
	}
	s.removeStaleTempDirs()

	suffix, err := randomSuffix()
	if err != nil {
		return false, err
	}
	staging := s.dir.Join(stagingPrefix + id + "-" + suffix)
	if err := staging.Join(artifactsDirName).MkdirAll(); err != nil {
		return false, err
	}
	defer staging.RemoveAll() // no-op after a successful rename

	md.FormatVersion = FormatVersion
	md.Manifest = *normalizedManifest(&md.Manifest)
	md.Artifacts = nil
	if md.CreatedAt.IsZero() {
		md.CreatedAt = time.Now()
	}
	for _, artifact := range artifacts {
		name := artifact.Base()
		size, sum, err := copyAndHash(artifact, staging.Join(artifactsDirName, name))
		if err != nil {
			return false, err
		}
		md.Artifacts = append(md.Artifacts, Artifact{Name: name, Size: size, SHA256: sum})
	}
	data, err := json.MarshalIndent(md, "", "  ")
	if err != nil {
		return false, err
	}
	// The metadata file is written last: an entry is valid only if it exists.
	if err := writeFileSync(staging.Join(metadataFileName), data); err != nil {
		return false, err
	}
	touch(staging)

	if err := staging.Rename(finalDir); err != nil {
		if finalDir.Join(metadataFileName).Exist() {
			// Stored concurrently by another process: keep that one.
			touch(finalDir)
			return false, nil
		}
		return false, err
	}
	s.evict()
	return true, nil
}

// evict removes the least recently used entries exceeding the maximum.
func (s *Store) evict() {
	entries := s.entries()
	if len(entries) <= s.maxEntries {
		return
	}
	for _, e := range entries[s.maxEntries:] {
		logrus.Tracef("Evicting build history entry %s", e.dir)
		removeDir(e.dir)
	}
}

// removeDir atomically makes the directory invisible to readers and then deletes it.
func removeDir(dir *paths.Path) {
	suffix, err := randomSuffix()
	if err != nil {
		return
	}
	trash := dir.Parent().Join(trashPrefix + dir.Base() + "-" + suffix)
	if err := dir.Rename(trash); err != nil {
		// Probably already removed by a concurrent process.
		logrus.WithError(err).Tracef("Could not remove build history entry %s", dir)
		return
	}
	_ = trash.RemoveAll()
}

// removeStaleTempDirs deletes staging and trash directories left behind by
// interrupted processes.
func (s *Store) removeStaleTempDirs() {
	dirs, err := s.dir.ReadDir()
	if err != nil {
		return
	}
	dirs.FilterDirs()
	for _, dir := range dirs {
		name := dir.Base()
		if !strings.HasPrefix(name, stagingPrefix) && !strings.HasPrefix(name, trashPrefix) {
			continue
		}
		if st, err := dir.Stat(); err == nil && time.Since(st.ModTime()) > staleTempAge {
			_ = dir.RemoveAll()
		}
	}
}

func isPlainFileName(name string) bool {
	return name != "" && name != "." && name != ".." && !strings.ContainsAny(name, `/\`)
}

func randomSuffix() (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

// HashFile returns the hex-encoded SHA-256 checksum of the file content.
func HashFile(p *paths.Path) (string, error) {
	f, err := p.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// copyAndHash copies src to dst (synced to disk) and returns the size and the
// SHA-256 checksum of the copied data.
func copyAndHash(src, dst *paths.Path) (int64, string, error) {
	in, err := src.Open()
	if err != nil {
		return 0, "", err
	}
	defer in.Close()
	out, err := dst.Create()
	if err != nil {
		return 0, "", err
	}
	h := sha256.New()
	size, err := io.Copy(io.MultiWriter(out, h), in)
	if err == nil {
		err = out.Sync()
	}
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		_ = dst.Remove()
		return 0, "", err
	}
	return size, hex.EncodeToString(h.Sum(nil)), nil
}

func writeFileSync(p *paths.Path, data []byte) error {
	f, err := os.OpenFile(p.String(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
