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

package builder

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/arduino/builder/cpp"
	"github.com/arduino/arduino-cli/arduino/builder/internal/utils"
	f "github.com/arduino/arduino-cli/internal/algorithms"
	"github.com/arduino/arduino-cli/internal/buildcache"
	"github.com/arduino/go-paths-helper"
	"github.com/pkg/errors"
)

// buildCore fixdoc
func (b *Builder) buildCore() error {
	if err := b.coreBuildPath.MkdirAll(); err != nil {
		return errors.WithStack(err)
	}

	if b.coreBuildCachePath != nil {
		if _, err := b.coreBuildCachePath.RelTo(b.buildPath); err != nil {
			b.logger.Info(tr("Couldn't deeply cache core build: %[1]s", err))
			b.logger.Info(tr("Running normal build of the core..."))
			// TODO decide if we want to override this or not. (It's only used by the
			// compileCore function).
			b.coreBuildCachePath = nil
		} else if err := b.coreBuildCachePath.MkdirAll(); err != nil {
			return errors.WithStack(err)
		}
	}

	archiveFile, objectFiles, err := b.compileCore()
	if err != nil {
		return errors.WithStack(err)
	}
	b.buildArtifacts.coreObjectsFiles = objectFiles
	b.buildArtifacts.coreArchiveFilePath = archiveFile
	return nil
}

func (b *Builder) compileCore() (*paths.Path, paths.PathList, error) {
	coreFolder := b.buildProperties.GetPath("build.core.path")
	variantFolder := b.buildProperties.GetPath("build.variant.path")
	targetCoreFolder := b.buildProperties.GetPath("runtime.platform.path")

	includes := []string{coreFolder.String()}
	if variantFolder != nil && variantFolder.IsDir() {
		includes = append(includes, variantFolder.String())
	}
	includes = f.Map(includes, cpp.WrapWithHyphenI)

	var err error
	variantObjectFiles := paths.NewPathList()
	if variantFolder != nil && variantFolder.IsDir() {
		variantObjectFiles, err = b.compileFiles(
			variantFolder, b.coreBuildPath,
			true, /** recursive **/
			includes,
		)
		if err != nil {
			return nil, nil, errors.WithStack(err)
		}
	}

	var targetArchivedCore *paths.Path
	if b.coreBuildCachePath != nil {
		realCoreFolder := coreFolder.Parent().Parent()
		archivedCoreName := getCachedCoreArchiveDirName(
			b.buildProperties.Get("build.fqbn"),
			b.buildProperties.Get("compiler.optimization_flags"),
			realCoreFolder,
		)
		targetArchivedCore = b.coreBuildCachePath.Join(archivedCoreName, "core.a")

		if _, err := buildcache.New(b.coreBuildCachePath).GetOrCreate(archivedCoreName); errors.Is(err, buildcache.CreateDirErr) {
			return nil, nil, fmt.Errorf(tr("creating core cache folder: %s", err))
		}

		var canUseArchivedCore bool
		if b.onlyUpdateCompilationDatabase || b.clean {
			canUseArchivedCore = false
		} else if isOlder, err := utils.DirContentIsOlderThan(realCoreFolder, targetArchivedCore); err != nil || !isOlder {
			// Recreate the archive if ANY of the core files (including platform.txt) has changed
			canUseArchivedCore = false
		} else if targetCoreFolder == nil || realCoreFolder.EquivalentTo(targetCoreFolder) {
			canUseArchivedCore = true
		} else if isOlder, err := utils.DirContentIsOlderThan(targetCoreFolder, targetArchivedCore); err != nil || !isOlder {
			// Recreate the archive if ANY of the build core files (including platform.txt) has changed
			canUseArchivedCore = false
		} else {
			canUseArchivedCore = true
		}

		if canUseArchivedCore {
			// use archived core
			if b.logger.Verbose() {
				b.logger.Info(tr("Using precompiled core: %[1]s", targetArchivedCore))
			}
			return targetArchivedCore, variantObjectFiles, nil
		}
	}

	coreObjectFiles, err := b.compileFiles(
		coreFolder, b.coreBuildPath,
		true, /** recursive **/
		includes,
	)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	archiveFile, err := b.archiveCompiledFiles(b.coreBuildPath, paths.New("core.a"), coreObjectFiles)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	// archive core.a
	if targetArchivedCore != nil && !b.onlyUpdateCompilationDatabase {
		err := archiveFile.CopyTo(targetArchivedCore)
		if b.logger.Verbose() {
			if err == nil {
				b.logger.Info(tr("Archiving built core (caching) in: %[1]s", targetArchivedCore))
			} else if os.IsNotExist(err) {
				b.logger.Info(tr("Unable to cache built core, please tell %[1]s maintainers to follow %[2]s",
					b.actualPlatform,
					"https://arduino.github.io/arduino-cli/latest/platform-specification/#recipes-to-build-the-corea-archive-file"))
			} else {
				b.logger.Info(tr("Error archiving built core (caching) in %[1]s: %[2]s", targetArchivedCore, err))
			}
		}
	}

	return archiveFile, variantObjectFiles, nil
}

// getCachedCoreArchiveDirName returns the directory name to be used to store
// the global cached core.a.
func getCachedCoreArchiveDirName(fqbn string, optimizationFlags string, coreFolder *paths.Path) string {
	fqbnToUnderscore := strings.ReplaceAll(fqbn, ":", "_")
	fqbnToUnderscore = strings.ReplaceAll(fqbnToUnderscore, "=", "_")
	if absCoreFolder, err := coreFolder.Abs(); err == nil {
		coreFolder = absCoreFolder
	} // silently continue if absolute path can't be detected

	md5Sum := func(data []byte) string {
		md5sumBytes := md5.Sum(data)
		return hex.EncodeToString(md5sumBytes[:])
	}
	hash := md5Sum([]byte(coreFolder.String() + optimizationFlags))
	realName := fqbnToUnderscore + "_" + hash
	if len(realName) > 100 {
		// avoid really long names, simply hash the name again
		realName = md5Sum([]byte(realName))
	}
	return realName
}
