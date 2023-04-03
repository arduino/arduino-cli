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
	"strings"

	f "github.com/arduino/arduino-cli/internal/algorithms"
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

func link(ctx *types.Context, objectFiles paths.PathList, coreDotARelPath *paths.Path, coreArchiveFilePath *paths.Path, buildProperties *properties.Map) error {
	objectFileList := strings.Join(f.Map(objectFiles.AsStrings(), wrapWithDoubleQuotes), " ")

	// If command line length is too big (> 30000 chars), try to collect the object files into archives
	// and use that archives to complete the build.
	if len(objectFileList) > 30000 {

		// We must create an object file for each visited directory: this is required becuase gcc-ar checks
		// if an object file is already in the archive by looking ONLY at the filename WITHOUT the path, so
		// it may happen that a subdir/spi.o inside the archive may be overwritten by a anotherdir/spi.o
		// because thery are both named spi.o.

		properties := buildProperties.Clone()
		archives := paths.NewPathList()
		for _, object := range objectFiles {
			if object.HasSuffix(".a") {
				archives.Add(object)
				continue
			}
			archive := object.Parent().Join("objs.a")
			if !archives.Contains(archive) {
				archives.Add(archive)
				// Cleanup old archives
				_ = archive.Remove()
			}
			properties.Set("archive_file", archive.Base())
			properties.SetPath("archive_file_path", archive)
			properties.SetPath("object_file", object)

			command, err := builder_utils.PrepareCommandForRecipe(properties, constants.RECIPE_AR_PATTERN, false, ctx.PackageManager.GetEnvVarsForSpawnedProcess())
			if err != nil {
				return errors.WithStack(err)
			}

			if _, _, err := utils.ExecCommand(ctx, command, utils.ShowIfVerbose /* stdout */, utils.Show /* stderr */); err != nil {
				return errors.WithStack(err)
			}
		}

		objectFileList = strings.Join(f.Map(archives.AsStrings(), wrapWithDoubleQuotes), " ")
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
