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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/arduino/arduino-cli/legacy/builder/builder_utils"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
)

type Linker struct{}

func (s *Linker) Run(ctx *types.Context) error {
	if ctx.OnlyUpdateCompilationDatabase {
		if ctx.Verbose {
			ctx.Info(tr("Skip linking of final executable."))
		}
		return nil
	}

	objectFilesSketch := ctx.SketchObjectFiles
	objectFilesLibraries := ctx.LibrariesObjectFiles
	objectFilesCore := ctx.CoreObjectsFiles

	objectFiles := paths.NewPathList()
	objectFiles.AddAll(objectFilesSketch)
	objectFiles.AddAll(objectFilesLibraries)
	objectFiles.AddAll(objectFilesCore)

	coreArchiveFilePath := ctx.CoreArchiveFilePath
	buildPath := ctx.BuildPath
	coreDotARelPath, err := buildPath.RelTo(coreArchiveFilePath)
	if err != nil {
		return errors.WithStack(err)
	}

	buildProperties := ctx.BuildProperties

	err = link(ctx, objectFiles, coreDotARelPath, coreArchiveFilePath, buildProperties)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// CppArchive represents a cpp archive (.a). It has a list of object files
// that must be part of the archive and has functions to build the archive
// and check if the archive is up-to-date.
type CppArchive struct {
	ArchivePath   *paths.Path
	CacheFilePath *paths.Path
	Objects       paths.PathList
}

// NewCppArchive creates an empty CppArchive
func NewCppArchive(archivePath *paths.Path) *CppArchive {
	return &CppArchive{
		ArchivePath:   archivePath,
		CacheFilePath: archivePath.Parent().Join(archivePath.Base() + ".cache"),
		Objects:       paths.NewPathList(),
	}
}

// AddObject adds an object file in the list of files to be archived
func (a *CppArchive) AddObject(object *paths.Path) {
	a.Objects.Add(object)
}

func (a *CppArchive) readCachedFilesList() paths.PathList {
	var cache paths.PathList
	if cacheData, err := a.CacheFilePath.ReadFile(); err != nil {
		return nil
	} else if err := json.Unmarshal(cacheData, &cache); err != nil {
		return nil
	} else {
		return cache
	}
}

func (a *CppArchive) writeCachedFilesList() error {
	if cacheData, err := json.Marshal(a.Objects); err != nil {
		panic(err) // should never happen
	} else if err := a.CacheFilePath.WriteFile(cacheData); err != nil {
		return err
	} else {
		return nil
	}
}

// IsUpToDate checks if an already made archive is up-to-date. If this
// method returns true, there is no need to Create the archive.
func (a *CppArchive) IsUpToDate() bool {
	archiveStat, err := a.ArchivePath.Stat()
	if err != nil {
		return false
	}

	cache := a.readCachedFilesList()
	if cache == nil || cache.Len() != a.Objects.Len() {
		return false
	}
	for _, object := range cache {
		objectStat, err := object.Stat()
		if err != nil {
			return false
		}
		if objectStat.ModTime().After(archiveStat.ModTime()) {
			return false
		}
	}

	return true
}

// Create will create the archive using the given arPattern
func (a *CppArchive) Create(ctx *types.Context, arPattern string) error {
	_ = a.ArchivePath.Remove()
	for _, object := range a.Objects {
		properties := properties.NewMap()
		properties.Set("archive_file", a.ArchivePath.Base())
		properties.SetPath("archive_file_path", a.ArchivePath)
		properties.SetPath("object_file", object)
		properties.Set("recipe.ar.pattern", arPattern)
		command, err := builder_utils.PrepareCommandForRecipe(properties, "recipe.ar.pattern", false, ctx.PackageManager.GetEnvVarsForSpawnedProcess())
		if err != nil {
			return errors.WithStack(err)
		}

		if _, _, err := utils.ExecCommand(ctx, command, utils.ShowIfVerbose /* stdout */, utils.Show /* stderr */); err != nil {
			return errors.WithStack(err)
		}
	}

	if err := a.writeCachedFilesList(); err != nil {
		ctx.Info("Error writing archive cache: " + err.Error())
	}
	return nil
}

func link(ctx *types.Context, objectFiles paths.PathList, coreDotARelPath *paths.Path, coreArchiveFilePath *paths.Path, buildProperties *properties.Map) error {
	objectFileList := strings.Join(utils.Map(objectFiles.AsStrings(), wrapWithDoubleQuotes), " ")

	// If command line length is too big (> 30000 chars), try to collect the object files into archives
	// and use that archives to complete the build.
	if len(objectFileList) > 30000 {
		buildObjectFiles := objectFiles.Clone()
		buildObjectFiles.FilterOutSuffix(".a")
		buildArchiveFiles := objectFiles.Clone()
		buildArchiveFiles.FilterSuffix(".a")

		// We must create an object file for each visited directory: this is required becuase gcc-ar checks
		// if an object file is already in the archive by looking ONLY at the filename WITHOUT the path, so
		// it may happen that a subdir/spi.o inside the archive may be overwritten by a anotherdir/spi.o
		// because thery are both named spi.o.

		// Split objects by directory and create a CppArchive for each directory
		archives := []*CppArchive{}
		{
			generatedArchivesFiles := map[string]*CppArchive{}
			for _, object := range buildObjectFiles {
				archive := object.Parent().Join("objs.a")
				a := generatedArchivesFiles[archive.String()]
				if a == nil {
					a = NewCppArchive(archive)
					archives = append(archives, a)
					generatedArchivesFiles[archive.String()] = a
				}
				a.AddObject(object)
			}
		}

		arPattern := buildProperties.ExpandPropsInString(buildProperties.Get("recipe.ar.pattern"))

		filesToLink := paths.NewPathList()
		for _, a := range archives {
			filesToLink.Add(a.ArchivePath)
			if a.IsUpToDate() {
				ctx.Info(fmt.Sprintf("%s %s", tr("Using previously build archive:"), a.ArchivePath))
				continue
			}
			if err := a.Create(ctx, arPattern); err != nil {
				return err
			}
		}

		// Add all remaining archives from the build
		filesToLink.AddAll(buildArchiveFiles)

		objectFileList = strings.Join(utils.Map(filesToLink.AsStrings(), wrapWithDoubleQuotes), " ")
		objectFileList = "-Wl,--whole-archive " + objectFileList + " -Wl,--no-whole-archive"
	}

	properties := buildProperties.Clone()
	properties.Set(constants.BUILD_PROPERTIES_COMPILER_C_ELF_FLAGS, properties.Get(constants.BUILD_PROPERTIES_COMPILER_C_ELF_FLAGS))
	properties.Set(constants.BUILD_PROPERTIES_COMPILER_WARNING_FLAGS, properties.Get(constants.BUILD_PROPERTIES_COMPILER_WARNING_FLAGS+"."+ctx.WarningsLevel))
	properties.Set(constants.BUILD_PROPERTIES_ARCHIVE_FILE, coreDotARelPath.String())
	properties.Set(constants.BUILD_PROPERTIES_ARCHIVE_FILE_PATH, coreArchiveFilePath.String())
	properties.Set("object_files", objectFileList)

	command, err := builder_utils.PrepareCommandForRecipe(properties, constants.RECIPE_C_COMBINE_PATTERN, false, ctx.PackageManager.GetEnvVarsForSpawnedProcess())
	if err != nil {
		return err
	}

	_, _, err = utils.ExecCommand(ctx, command, utils.ShowIfVerbose /* stdout */, utils.Show /* stderr */)
	return err
}

func wrapWithDoubleQuotes(value string) string {
	return "\"" + value + "\""
}
