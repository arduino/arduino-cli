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

package phases

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/arduino/builder"
	"github.com/arduino/arduino-cli/arduino/builder/cpp"
	"github.com/arduino/arduino-cli/arduino/builder/progress"
	"github.com/arduino/arduino-cli/arduino/builder/utils"
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/buildcache"
	"github.com/arduino/arduino-cli/i18n"
	f "github.com/arduino/arduino-cli/internal/algorithms"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
)

var tr = i18n.Tr

func CoreBuilder(
	buildPath, coreBuildPath, coreBuildCachePath *paths.Path,
	buildProperties *properties.Map,
	actualPlatform *cores.PlatformRelease,
	verbose, onlyUpdateCompilationDatabase, clean bool,
	compilationDatabase *builder.CompilationDatabase,
	jobs int,
	warningsLevel string,
	stdoutWriter, stderrWriter io.Writer,
	verboseInfoFn func(msg string),
	verboseStdoutFn, verboseStderrFn func(data []byte),
	progress *progress.Struct, progressCB rpc.TaskProgressCB,
) (paths.PathList, *paths.Path, error) {
	if err := coreBuildPath.MkdirAll(); err != nil {
		return nil, nil, errors.WithStack(err)
	}

	if coreBuildCachePath != nil {
		if _, err := coreBuildCachePath.RelTo(buildPath); err != nil {
			verboseInfoFn(tr("Couldn't deeply cache core build: %[1]s", err))
			verboseInfoFn(tr("Running normal build of the core..."))
			coreBuildCachePath = nil
		} else if err := coreBuildCachePath.MkdirAll(); err != nil {
			return nil, nil, errors.WithStack(err)
		}
	}

	archiveFile, objectFiles, err := compileCore(
		verbose, onlyUpdateCompilationDatabase, clean,
		actualPlatform,
		coreBuildPath, coreBuildCachePath,
		buildProperties,
		compilationDatabase,
		jobs,
		warningsLevel,
		stdoutWriter, stderrWriter,
		verboseInfoFn,
		verboseStdoutFn, verboseStderrFn,
		progress, progressCB,
	)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	return objectFiles, archiveFile, nil
}

func compileCore(
	verbose, onlyUpdateCompilationDatabase, clean bool,
	actualPlatform *cores.PlatformRelease,
	buildPath, buildCachePath *paths.Path,
	buildProperties *properties.Map,
	compilationDatabase *builder.CompilationDatabase,
	jobs int,
	warningsLevel string,
	stdoutWriter, stderrWriter io.Writer,
	verboseInfoFn func(msg string),
	verboseStdoutFn, verboseStderrFn func(data []byte),
	progress *progress.Struct, progressCB rpc.TaskProgressCB,
) (*paths.Path, paths.PathList, error) {
	coreFolder := buildProperties.GetPath("build.core.path")
	variantFolder := buildProperties.GetPath("build.variant.path")
	targetCoreFolder := buildProperties.GetPath("runtime.platform.path")

	includes := []string{coreFolder.String()}
	if variantFolder != nil && variantFolder.IsDir() {
		includes = append(includes, variantFolder.String())
	}
	includes = f.Map(includes, cpp.WrapWithHyphenI)

	var err error
	variantObjectFiles := paths.NewPathList()
	if variantFolder != nil && variantFolder.IsDir() {
		variantObjectFiles, err = utils.CompileFilesRecursive(
			variantFolder, buildPath, buildProperties, includes,
			onlyUpdateCompilationDatabase,
			compilationDatabase,
			jobs,
			verbose,
			warningsLevel,
			stdoutWriter, stderrWriter,
			verboseInfoFn, verboseStdoutFn, verboseStderrFn,
			progress, progressCB,
		)
		if err != nil {
			return nil, nil, errors.WithStack(err)
		}
	}

	var targetArchivedCore *paths.Path
	if buildCachePath != nil {
		realCoreFolder := coreFolder.Parent().Parent()
		archivedCoreName := GetCachedCoreArchiveDirName(
			buildProperties.Get("build.fqbn"),
			buildProperties.Get("compiler.optimization_flags"),
			realCoreFolder,
		)
		targetArchivedCore = buildCachePath.Join(archivedCoreName, "core.a")

		if _, err := buildcache.New(buildCachePath).GetOrCreate(archivedCoreName); errors.Is(err, buildcache.CreateDirErr) {
			return nil, nil, fmt.Errorf(tr("creating core cache folder: %s", err))
		}

		var canUseArchivedCore bool
		if onlyUpdateCompilationDatabase || clean {
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
			if verbose {
				verboseInfoFn(tr("Using precompiled core: %[1]s", targetArchivedCore))
			}
			return targetArchivedCore, variantObjectFiles, nil
		}
	}

	coreObjectFiles, err := utils.CompileFilesRecursive(
		coreFolder, buildPath, buildProperties, includes,
		onlyUpdateCompilationDatabase,
		compilationDatabase,
		jobs,
		verbose,
		warningsLevel,
		stdoutWriter, stderrWriter,
		verboseInfoFn, verboseStdoutFn, verboseStderrFn,
		progress, progressCB,
	)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	archiveFile, verboseInfo, err := utils.ArchiveCompiledFiles(
		buildPath, paths.New("core.a"), coreObjectFiles, buildProperties,
		onlyUpdateCompilationDatabase, verbose, stdoutWriter, stderrWriter,
	)
	if verbose {
		verboseInfoFn(string(verboseInfo))
	}
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	// archive core.a
	if targetArchivedCore != nil && !onlyUpdateCompilationDatabase {
		err := archiveFile.CopyTo(targetArchivedCore)
		if verbose {
			if err == nil {
				verboseInfoFn(tr("Archiving built core (caching) in: %[1]s", targetArchivedCore))
			} else if os.IsNotExist(err) {
				verboseInfoFn(tr("Unable to cache built core, please tell %[1]s maintainers to follow %[2]s",
					actualPlatform,
					"https://arduino.github.io/arduino-cli/latest/platform-specification/#recipes-to-build-the-corea-archive-file"))
			} else {
				verboseInfoFn(tr("Error archiving built core (caching) in %[1]s: %[2]s", targetArchivedCore, err))
			}
		}
	}

	return archiveFile, variantObjectFiles, nil
}

// GetCachedCoreArchiveDirName returns the directory name to be used to store
// the global cached core.a.
func GetCachedCoreArchiveDirName(fqbn string, optimizationFlags string, coreFolder *paths.Path) string {
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
