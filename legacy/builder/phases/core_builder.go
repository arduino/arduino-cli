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
	"os"
	"strings"

	"github.com/arduino/arduino-cli/buildcache"
	"github.com/arduino/arduino-cli/i18n"
	f "github.com/arduino/arduino-cli/internal/algorithms"
	"github.com/arduino/arduino-cli/legacy/builder/builder_utils"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
)

type CoreBuilder struct{}

var tr = i18n.Tr

func (s *CoreBuilder) Run(ctx *types.Context) error {
	coreBuildPath := ctx.CoreBuildPath
	coreBuildCachePath := ctx.CoreBuildCachePath
	buildProperties := ctx.BuildProperties

	if err := coreBuildPath.MkdirAll(); err != nil {
		return errors.WithStack(err)
	}

	if coreBuildCachePath != nil {
		if _, err := coreBuildCachePath.RelTo(ctx.BuildPath); err != nil {
			ctx.Info(tr("Couldn't deeply cache core build: %[1]s", err))
			ctx.Info(tr("Running normal build of the core..."))
			coreBuildCachePath = nil
			ctx.CoreBuildCachePath = nil
		} else if err := coreBuildCachePath.MkdirAll(); err != nil {
			return errors.WithStack(err)
		}
	}

	archiveFile, objectFiles, err := compileCore(ctx, coreBuildPath, coreBuildCachePath, buildProperties)
	if err != nil {
		return errors.WithStack(err)
	}

	ctx.CoreArchiveFilePath = archiveFile
	ctx.CoreObjectsFiles = objectFiles

	return nil
}

func compileCore(ctx *types.Context, buildPath *paths.Path, buildCachePath *paths.Path, buildProperties *properties.Map) (*paths.Path, paths.PathList, error) {
	coreFolder := buildProperties.GetPath("build.core.path")
	variantFolder := buildProperties.GetPath("build.variant.path")

	targetCoreFolder := buildProperties.GetPath(constants.BUILD_PROPERTIES_RUNTIME_PLATFORM_PATH)

	includes := []string{}
	includes = append(includes, coreFolder.String())
	if variantFolder != nil && variantFolder.IsDir() {
		includes = append(includes, variantFolder.String())
	}
	includes = f.Map(includes, utils.WrapWithHyphenI)

	var err error

	variantObjectFiles := paths.NewPathList()
	if variantFolder != nil && variantFolder.IsDir() {
		variantObjectFiles, err = builder_utils.CompileFilesRecursive(ctx, variantFolder, buildPath, buildProperties, includes)
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
			realCoreFolder)
		targetArchivedCore = buildCachePath.Join(archivedCoreName, "core.a")

		if _, err := buildcache.New(buildCachePath).GetOrCreate(archivedCoreName); errors.Is(err, buildcache.CreateDirErr) {
			return nil, nil, fmt.Errorf(tr("creating core cache folder: %s", err))
		}

		var canUseArchivedCore bool
		if ctx.OnlyUpdateCompilationDatabase || ctx.Clean {
			canUseArchivedCore = false
		} else if isOlder, err := builder_utils.DirContentIsOlderThan(realCoreFolder, targetArchivedCore); err != nil || !isOlder {
			// Recreate the archive if ANY of the core files (including platform.txt) has changed
			canUseArchivedCore = false
		} else if targetCoreFolder == nil || realCoreFolder.EquivalentTo(targetCoreFolder) {
			canUseArchivedCore = true
		} else if isOlder, err := builder_utils.DirContentIsOlderThan(targetCoreFolder, targetArchivedCore); err != nil || !isOlder {
			// Recreate the archive if ANY of the build core files (including platform.txt) has changed
			canUseArchivedCore = false
		} else {
			canUseArchivedCore = true
		}

		if canUseArchivedCore {
			// use archived core
			if ctx.Verbose {
				ctx.Info(tr("Using precompiled core: %[1]s", targetArchivedCore))
			}
			return targetArchivedCore, variantObjectFiles, nil
		}
	}

	coreObjectFiles, err := builder_utils.CompileFilesRecursive(ctx, coreFolder, buildPath, buildProperties, includes)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	archiveFile, err := builder_utils.ArchiveCompiledFiles(ctx, buildPath, paths.New("core.a"), coreObjectFiles, buildProperties)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	// archive core.a
	if targetArchivedCore != nil && !ctx.OnlyUpdateCompilationDatabase {
		err := archiveFile.CopyTo(targetArchivedCore)
		if ctx.Verbose {
			if err == nil {
				ctx.Info(tr("Archiving built core (caching) in: %[1]s", targetArchivedCore))
			} else if os.IsNotExist(err) {
				ctx.Info(tr("Unable to cache built core, please tell %[1]s maintainers to follow %[2]s",
					ctx.ActualPlatform,
					"https://arduino.github.io/arduino-cli/latest/platform-specification/#recipes-to-build-the-corea-archive-file"))
			} else {
				ctx.Info(tr("Error archiving built core (caching) in %[1]s: %[2]s", targetArchivedCore, err))
			}
		}
	}

	return archiveFile, variantObjectFiles, nil
}

// GetCachedCoreArchiveDirName returns the directory name to be used to store
// the global cached core.a.
func GetCachedCoreArchiveDirName(fqbn string, optimizationFlags string, coreFolder *paths.Path) string {
	fqbnToUnderscore := strings.Replace(fqbn, ":", "_", -1)
	fqbnToUnderscore = strings.Replace(fqbnToUnderscore, "=", "_", -1)
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
