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
	"bytes"
	"io"
	"strings"

	"github.com/arduino/arduino-cli/arduino/builder/utils"
	f "github.com/arduino/arduino-cli/internal/algorithms"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
)

func Linker(
	onlyUpdateCompilationDatabase, verbose bool,
	sketchObjectFiles, librariesObjectFiles, coreObjectsFiles paths.PathList,
	coreArchiveFilePath, buildPath *paths.Path,
	buildProperties *properties.Map,
	stdoutWriter, stderrWriter io.Writer,
	warningsLevel string,
) ([]byte, error) {
	verboseInfo := &bytes.Buffer{}
	if onlyUpdateCompilationDatabase {
		if verbose {
			verboseInfo.WriteString(tr("Skip linking of final executable."))
		}
		return verboseInfo.Bytes(), nil
	}

	objectFilesSketch := sketchObjectFiles
	objectFilesLibraries := librariesObjectFiles
	objectFilesCore := coreObjectsFiles

	objectFiles := paths.NewPathList()
	objectFiles.AddAll(objectFilesSketch)
	objectFiles.AddAll(objectFilesLibraries)
	objectFiles.AddAll(objectFilesCore)

	coreDotARelPath, err := buildPath.RelTo(coreArchiveFilePath)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	verboseInfoOut, err := link(
		objectFiles, coreDotARelPath, coreArchiveFilePath, buildProperties,
		verbose, stdoutWriter, stderrWriter, warningsLevel,
	)
	verboseInfo.Write(verboseInfoOut)
	if err != nil {
		return verboseInfo.Bytes(), errors.WithStack(err)
	}

	return verboseInfo.Bytes(), nil
}

func link(
	objectFiles paths.PathList, coreDotARelPath *paths.Path, coreArchiveFilePath *paths.Path, buildProperties *properties.Map,
	verbose bool,
	stdoutWriter, stderrWriter io.Writer,
	warningsLevel string,
) ([]byte, error) {
	verboseBuffer := &bytes.Buffer{}
	wrapWithDoubleQuotes := func(value string) string { return "\"" + value + "\"" }
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

			command, err := utils.PrepareCommandForRecipe(properties, "recipe.ar.pattern", false)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			if verboseInfo, _, _, err := utils.ExecCommand(verbose, stdoutWriter, stderrWriter, command, utils.ShowIfVerbose /* stdout */, utils.Show /* stderr */); err != nil {
				if verbose {
					verboseBuffer.WriteString(string(verboseInfo))
				}
				return verboseBuffer.Bytes(), errors.WithStack(err)
			}
		}

		objectFileList = strings.Join(f.Map(archives.AsStrings(), wrapWithDoubleQuotes), " ")
		objectFileList = "-Wl,--whole-archive " + objectFileList + " -Wl,--no-whole-archive"
	}

	properties := buildProperties.Clone()
	properties.Set("compiler.c.elf.flags", properties.Get("compiler.c.elf.flags"))
	properties.Set("compiler.warning_flags", properties.Get("compiler.warning_flags."+warningsLevel))
	properties.Set("archive_file", coreDotARelPath.String())
	properties.Set("archive_file_path", coreArchiveFilePath.String())
	properties.Set("object_files", objectFileList)

	command, err := utils.PrepareCommandForRecipe(properties, "recipe.c.combine.pattern", false)
	if err != nil {
		return verboseBuffer.Bytes(), err
	}

	verboseInfo, _, _, err := utils.ExecCommand(verbose, stdoutWriter, stderrWriter, command, utils.ShowIfVerbose /* stdout */, utils.Show /* stderr */)
	if verbose {
		verboseBuffer.WriteString(string(verboseInfo))
	}
	return verboseBuffer.Bytes(), err
}
