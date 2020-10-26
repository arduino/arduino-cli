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
	objectFileList := strings.Join(utils.Map(objectFiles.AsStrings(), wrapWithDoubleQuotes), " ")

	// If command line length is too big (> 30000 chars), try to collect the object files into an archive
	// and use that archive to complete the build.
	if len(objectFileList) > 30000 {
		objectFilesArchive := ctx.BuildPath.Join("object_files.a")
		// Recreate the archive at every build
		_ = objectFilesArchive.Remove()
		properties := buildProperties.Clone()
		properties.Set("archive_file", objectFilesArchive.Base())
		properties.SetPath("archive_file_path", objectFilesArchive)
		for _, objFile := range objectFiles {
			properties.SetPath("object_file", objFile)
			_, _, err := builder_utils.ExecRecipe(ctx, properties, constants.RECIPE_AR_PATTERN, false, utils.ShowIfVerbose /* stdout */, utils.Show /* stderr */)
			if err != nil {
				return err
			}
		}
		objectFileList = "-Wl,--whole-archive " + wrapWithDoubleQuotes(objectFilesArchive.String()) + " -Wl,--no-whole-archive"
	}

	properties := buildProperties.Clone()
	properties.Set(constants.BUILD_PROPERTIES_COMPILER_C_ELF_FLAGS, properties.Get(constants.BUILD_PROPERTIES_COMPILER_C_ELF_FLAGS))
	properties.Set(constants.BUILD_PROPERTIES_COMPILER_WARNING_FLAGS, properties.Get(constants.BUILD_PROPERTIES_COMPILER_WARNING_FLAGS+"."+ctx.WarningsLevel))
	properties.Set(constants.BUILD_PROPERTIES_ARCHIVE_FILE, coreDotARelPath.String())
	properties.Set(constants.BUILD_PROPERTIES_ARCHIVE_FILE_PATH, coreArchiveFilePath.String())
	properties.Set("object_files", objectFileList)

	_, _, err := builder_utils.ExecRecipe(ctx, properties, constants.RECIPE_C_COMBINE_PATTERN, false /* stdout */, utils.ShowIfVerbose /* stderr */, utils.Show)
	return err
}

func wrapWithDoubleQuotes(value string) string {
	return "\"" + value + "\""
}
