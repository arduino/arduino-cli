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

package builder

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/arduino/arduino-cli/internal/arduino/builder/cpp"
	"github.com/arduino/arduino-cli/internal/arduino/builder/internal/utils"
	"github.com/arduino/arduino-cli/internal/arduino/builder/logger"
	"github.com/arduino/arduino-cli/internal/arduino/globals"
	"github.com/arduino/arduino-cli/internal/arduino/libraries"
	"github.com/arduino/arduino-cli/internal/buildhistory"
	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/arduino/arduino-cli/internal/version"
	"github.com/arduino/go-paths-helper"
)

// Build history
//
// A build is identified by a two-level fingerprint:
//
//   - the configuration key, computed before building, covers everything that is
//     known upfront: all the resolved build properties (FQBN and board options,
//     platform, toolchain paths and versions, recipes, compiler flags, custom
//     build properties), the warnings level, the content of all the sketch files,
//     the inventory of the installed libraries, the state of the platform, core
//     and variant folders, the signing keys and the CLI version;
//
//   - the manifest, recorded after a successful build, lists the content
//     checksum of every file the compiler reported as a dependency in the
//     dependency (.d) files of the build path (library sources, all the included
//     headers...) and the list of source files of every used library.
//
// A stored build is reused only if both the key matches and every file in its
// manifest still has the same content.

// historyKeyFormat must be changed whenever the fingerprint computation changes.
const historyKeyFormat = "1"

// nonArtifactFiles are the files in the root of the build path that hold the
// state of the incremental build and are not part of the build output.
var nonArtifactFiles = map[string]bool{
	"build.options.json":    true,
	"compile_commands.json": true,
	"includes.cache":        true,
	"libraries.cache":       true,
	".last-used":            true,
}

// timestampMacros make the compiler output depend on the time of the build.
var timestampMacros = [][]byte{[]byte("__DATE__"), []byte("__TIME__"), []byte("__TIMESTAMP__")}

type historyBuild struct {
	key   string
	start time.Time
}

// prepareHistory returns the fingerprint of the current build, or nil if the
// build history is disabled or can not be used for this build.
func (b *Builder) prepareHistory() *historyBuild {
	if b.history == nil {
		return nil
	}
	start := time.Now()
	key, notCacheableReason, err := b.computeHistoryKey()
	if err != nil {
		notCacheableReason = err.Error()
	}
	if notCacheableReason != "" {
		b.logIfVerbose(false, i18n.Tr("Build history not used: %s", notCacheableReason))
		return nil
	}
	return &historyBuild{key: key, start: start}
}

// restoreFromHistory looks for a previous build identical to the current one and,
// if found, restores its artifacts in the build path. Returns true if the build
// has been restored.
func (b *Builder) restoreFromHistory(h *historyBuild) bool {
	hashes := map[string]string{}
	entry := b.history.Lookup(h.key, func(md *buildhistory.Metadata) bool {
		return b.historyManifestIsValid(&md.Manifest, hashes)
	})
	if entry == nil {
		b.logIfVerbose(false, i18n.Tr("No matching build found in the build history."))
		return false
	}

	importedLibs, ok := b.findInstalledLibraries(entry.Manifest.Libraries)
	if !ok {
		return false
	}
	if err := entry.Restore(b.buildPath); err != nil {
		b.logIfVerbose(true, i18n.Tr("Could not restore the build from the build history: %s", err))
		return false
	}
	b.history.Touch(entry)
	b.libsDetector.SetImportedLibraries(importedLibs)

	if b.logger.VerbosityLevel() != logger.VerbosityQuiet {
		b.logger.Info(i18n.Tr("Using a previous identical build from the build history, compilation skipped."))
	}
	b.logIfVerbose(false, i18n.Tr("Build restored from: %s", entry.Dir()))
	return true
}

// saveToHistory stores the artifacts of the current (successful) build in the
// build history. Errors are not fatal: the build is simply not stored.
func (b *Builder) saveToHistory(h *historyBuild) {
	if err := b.doSaveToHistory(h); err != nil {
		b.logIfVerbose(true, i18n.Tr("Could not save the build in the build history: %s", err))
	}
}

func (b *Builder) doSaveToHistory(h *historyBuild) error {
	// Ensure that the inputs did not change while building
	if key, _, err := b.computeHistoryKey(); err != nil {
		return err
	} else if key != h.key {
		return errors.New(i18n.Tr("the build inputs have been modified during the build"))
	}
	manifest, err := b.collectHistoryManifest(h.start)
	if err != nil {
		return err
	}
	artifacts, err := b.collectHistoryArtifacts()
	if err != nil {
		return err
	}
	added, err := b.history.Add(buildhistory.Metadata{
		Key:        h.key,
		Manifest:   *manifest,
		FQBN:       b.buildProperties.Get("build.fqbn"),
		SketchPath: b.sketch.FullPath.String(),
	}, artifacts)
	if err != nil {
		return err
	}
	if added {
		b.logIfVerbose(false, i18n.Tr("Build saved in the build history."))
	}
	return nil
}

// computeHistoryKey returns the configuration key of the build. If the build
// output is not reproducible (for example because it embeds the build time) a
// non-empty reason is returned instead.
func (b *Builder) computeHistoryKey() (string, string, error) {
	h := sha256.New()
	add := func(kind, name, value string) {
		fmt.Fprintf(h, "%s %q %q\n", kind, name, value)
	}
	add("format", "", historyKeyFormat)
	add("cli", "", version.VersionInfo.VersionString)
	add("warnings", "", b.logger.WarningsLevel())

	// All the resolved build properties: FQBN, board options, platform and
	// tools paths (that include their versions), recipes and flags...
	keys := b.buildProperties.Keys()
	sort.Strings(keys)
	for _, k := range keys {
		if strings.HasPrefix(k, "extra.time.") {
			continue
		}
		v := b.buildProperties.Get(k)
		if strings.Contains(v, "{extra.time.") {
			return "", i18n.Tr("the platform embeds the build time in the firmware"), nil
		}
		add("property", k, v)
	}

	// Content of all the sketch files, including the unsaved editor buffers
	files := paths.PathList{b.sketch.MainFile}
	files.AddAllMissing(b.sketch.OtherSketchFiles)
	files.AddAllMissing(b.sketch.AdditionalFiles)
	files.AddAllMissing(b.sketch.RootFolderFiles)
	files.Sort()
	for _, f := range files {
		rel, err := b.sketch.FullPath.RelTo(f)
		if err != nil {
			return "", "", err
		}
		var data []byte
		if override, ok := b.sourceOverrides[rel.String()]; ok {
			data = []byte(override)
		} else if data, err = f.ReadFile(); err != nil {
			return "", "", err
		}
		if isCompiledSource(f) {
			for _, macro := range timestampMacros {
				if bytes.Contains(data, macro) {
					return "", i18n.Tr("the sketch uses the %s macro", string(macro)), nil
				}
			}
		}
		sum := sha256.Sum256(data)
		add("sketch", filepath.ToSlash(rel.String()), hex.EncodeToString(sum[:]))
	}

	// Inventory of the libraries that may be selected by the library discovery:
	// installing, removing or upgrading a library may change the resolution.
	var libs []string
	for _, lib := range b.libsDetector.InstalledLibraries() {
		libs = append(libs, fmt.Sprintf("%s|%s|%v|%s", lib.Name, lib.Version, lib.Location, lib.InstallDir))
	}
	sort.Strings(libs)
	for _, lib := range libs {
		add("library", "", lib)
	}

	// The platform, core and variant files (boards.txt, platform.txt, core
	// sources, linker scripts, bootloaders...). Like the core build cache, file
	// changes are detected using size and modification time.
	for _, dir := range b.historyPlatformDirs() {
		err := filepath.WalkDir(dir.String(), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			add("platform-file", path, fmt.Sprintf("%d %d", info.Size(), info.ModTime().UnixNano()))
			return nil
		})
		if err != nil {
			return "", "", err
		}
	}

	// The content of the signing and encryption keys
	if keychain := b.buildProperties.ExpandPropsInString(b.buildProperties.Get("build.keys.keychain")); keychain != "" {
		keychainDir := paths.New(keychain)
		if keychainDir.IsDir() {
			keyFiles, err := keychainDir.ReadDirRecursiveFiltered(nil, paths.FilterOutDirectories())
			if err != nil {
				return "", "", err
			}
			keyFiles.Sort()
			for _, f := range keyFiles {
				sum, err := buildhistory.HashFile(f)
				if err != nil {
					return "", "", err
				}
				add("key-file", f.String(), sum)
			}
		}
	}

	return hex.EncodeToString(h.Sum(nil)), "", nil
}

// historyPlatformDirs returns the platform, core and variant folders used by
// the build, without duplicates or folders contained in other folders.
func (b *Builder) historyPlatformDirs() paths.PathList {
	candidates := paths.PathList{
		b.targetPlatform.InstallDir,
		b.buildPlatform.InstallDir,
		b.buildProperties.GetPath("build.core.path"),
		b.buildProperties.GetPath("build.variant.path"),
	}
	var res paths.PathList
	for _, dir := range candidates {
		if dir == nil || !dir.IsDir() {
			continue
		}
		contained := false
		for _, other := range candidates {
			if other != nil && !other.EqualsTo(dir) {
				if inside, _ := dir.IsInsideDir(other); inside {
					contained = true
					break
				}
			}
		}
		if !contained {
			res.AddIfMissing(dir)
		}
	}
	res.Sort()
	return res
}

func isCompiledSource(f *paths.Path) bool {
	ext := f.Ext()
	_, isSource := globals.SourceFilesValidExtensions[ext]
	return isSource || globals.MainFileValidExtensions[ext] || globals.HeaderFilesValidExtensions[ext]
}

// collectHistoryManifest returns the manifest of the inputs of the completed
// build, extracted from the dependency files generated by the compiler.
func (b *Builder) collectHistoryManifest(buildStart time.Time) (*buildhistory.Manifest, error) {
	depFiles, err := b.buildPath.ReadDirRecursiveFiltered(nil,
		paths.FilterOutDirectories(),
		paths.FilterSuffixes(".d"))
	if err != nil {
		return nil, err
	}
	inputs := map[string]*paths.Path{}
	for _, depFile := range depFiles {
		deps, err := cpp.ReadDepFile(depFile)
		if err != nil {
			return nil, err
		}
		for _, dep := range deps.Dependencies {
			p := paths.New(dep)
			if !p.IsAbs() {
				p = b.buildPath.Join(dep)
			}
			p = p.Clean()
			// Files in the build path are generated from the sketch, that is
			// already part of the configuration key.
			if inside, _ := p.IsInsideDir(b.buildPath); inside {
				continue
			}
			inputs[p.String()] = p
		}
	}

	// The precompiled archives of the libraries are linked but are not
	// reported in the dependency files.
	manifest := &buildhistory.Manifest{}
	for _, lib := range b.libsDetector.ImportedLibraries() {
		digest, err := libraryListingDigest(lib)
		if err != nil {
			return nil, err
		}
		manifest.Libraries = append(manifest.Libraries, buildhistory.DirListing{Path: lib.InstallDir.String(), Digest: digest})
		if lib.Precompiled && lib.SourceDir != nil {
			archives, err := lib.SourceDir.ReadDirRecursiveFiltered(nil,
				paths.FilterOutDirectories(),
				paths.FilterSuffixes(".a", ".so"))
			if err != nil {
				return nil, err
			}
			for _, archive := range archives {
				inputs[archive.String()] = archive
			}
		}
	}

	for _, p := range inputs {
		st, err := p.Stat()
		if err != nil {
			// Stale dependency file of a removed object, not an input of this build.
			continue
		}
		if st.ModTime().After(buildStart) {
			return nil, errors.New(i18n.Tr("the build inputs have been modified during the build"))
		}
		sum, err := buildhistory.HashFile(p)
		if err != nil {
			return nil, err
		}
		manifest.Files = append(manifest.Files, buildhistory.FileHash{Path: p.String(), SHA256: sum})
	}
	return manifest, nil
}

// libraryListingDigest returns a digest of the list of source, header and
// archive files of the library, used to detect added or removed files.
func libraryListingDigest(lib *libraries.Library) (string, error) {
	extensions := []string{".a", ".so"}
	for ext := range globals.SourceFilesValidExtensions {
		extensions = append(extensions, ext)
	}
	for ext := range globals.HeaderFilesValidExtensions {
		extensions = append(extensions, ext)
	}
	files, err := utils.FindFilesInFolder(lib.InstallDir, true, nil, extensions...)
	if err != nil {
		return "", err
	}
	var rels []string
	for _, f := range files {
		rel, err := lib.InstallDir.RelTo(f)
		if err != nil {
			return "", err
		}
		rels = append(rels, filepath.ToSlash(rel.String()))
	}
	sort.Strings(rels)
	sum := sha256.Sum256([]byte(strings.Join(rels, "\n")))
	return hex.EncodeToString(sum[:]), nil
}

// historyManifestIsValid returns true if all the inputs listed in the manifest
// still have the same content. Computed checksums are memoized in hashes.
func (b *Builder) historyManifestIsValid(m *buildhistory.Manifest, hashes map[string]string) bool {
	for _, f := range m.Files {
		sum, ok := hashes[f.Path]
		if !ok {
			var err error
			if sum, err = buildhistory.HashFile(paths.New(f.Path)); err != nil {
				sum = ""
			}
			hashes[f.Path] = sum
		}
		if sum != f.SHA256 {
			return false
		}
	}
	installed := b.libsDetector.InstalledLibraries()
	for _, l := range m.Libraries {
		lib := findLibraryByInstallDir(installed, l.Path)
		if lib == nil {
			return false
		}
		if digest, err := libraryListingDigest(lib); err != nil || digest != l.Digest {
			return false
		}
	}
	return true
}

// findInstalledLibraries returns the installed libraries corresponding to the
// given listings, in the same order.
func (b *Builder) findInstalledLibraries(listings []buildhistory.DirListing) (libraries.List, bool) {
	installed := b.libsDetector.InstalledLibraries()
	res := libraries.List{}
	for _, l := range listings {
		lib := findLibraryByInstallDir(installed, l.Path)
		if lib == nil {
			return nil, false
		}
		res = append(res, lib)
	}
	return res, true
}

func findLibraryByInstallDir(libs libraries.List, dir string) *libraries.Library {
	for _, lib := range libs {
		if lib.InstallDir != nil && lib.InstallDir.String() == dir {
			return lib
		}
	}
	return nil
}

// collectHistoryArtifacts returns the files produced by the build that must be
// stored: all the files in the root of the build path, where the upload recipes
// look for them, except the ones holding the incremental build state.
func (b *Builder) collectHistoryArtifacts() (paths.PathList, error) {
	files, err := b.buildPath.ReadDir()
	if err != nil {
		return nil, err
	}
	files.FilterOutDirs()
	files.Filter(func(f *paths.Path) bool { return !nonArtifactFiles[f.Base()] })

	projectName := b.buildProperties.Get("build.project_name")
	for _, f := range files {
		if strings.HasPrefix(f.Base(), projectName) {
			return files, nil
		}
	}
	return nil, errors.New(i18n.Tr("no build artifacts found in %s", b.buildPath))
}
