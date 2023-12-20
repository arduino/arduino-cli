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
	"strings"

	f "github.com/arduino/arduino-cli/internal/algorithms"
	"github.com/arduino/go-paths-helper"
)

// link fixdoc
func (b *Builder) link() error {
	if b.onlyUpdateCompilationDatabase {
		if b.logger.Verbose() {
			b.logger.Info(tr("Skip linking of final executable."))
		}
		return nil
	}

	objectFiles := paths.NewPathList()
	objectFiles.AddAll(b.buildArtifacts.sketchObjectFiles)
	objectFiles.AddAll(b.buildArtifacts.librariesObjectFiles)
	objectFiles.AddAll(b.buildArtifacts.coreObjectsFiles)

	coreDotARelPath, err := b.buildPath.RelTo(b.buildArtifacts.coreArchiveFilePath)
	if err != nil {
		return err
	}

	wrapWithDoubleQuotes := func(value string) string { return "\"" + value + "\"" }
	objectFileList := strings.Join(f.Map(objectFiles.AsStrings(), wrapWithDoubleQuotes), " ")

	// If command line length is too big (> 30000 chars), try to collect the object files into archives
	// and use that archives to complete the build.
	if len(objectFileList) > 30000 {

		// We must create an object file for each visited directory: this is required because gcc-ar checks
		// if an object file is already in the archive by looking ONLY at the filename WITHOUT the path, so
		// it may happen that a subdir/spi.o inside the archive may be overwritten by a anotherdir/spi.o
		// because thery are both named spi.o.

		// Put all the existing archives apart from the other object files
		existingArchives := objectFiles.Clone()
		existingArchives.FilterSuffix(".a")
		objectFiles.FilterOutSuffix(".a")

		// Generate an archive for each directory from the remaining object files
		newArchives := paths.NewPathList()
		for _, object := range objectFiles {
			archive := object.Parent().Join("objs.a")
			newArchives.AddIfMissing(archive)
		}
		for _, archive := range newArchives {
			archiveDir := archive.Parent()
			relatedObjectFiles := objectFiles.Clone()
			relatedObjectFiles.Filter(func(object *paths.Path) bool {
				// extract all the object files that are in the same directory of the archive
				return object.Parent().EquivalentTo(archiveDir)
			})
			b.archiveCompiledFiles(archive, relatedObjectFiles)
		}

		// Put everything together
		allArchives := existingArchives.Clone()
		allArchives.AddAll(newArchives)
		objectFileList = strings.Join(f.Map(allArchives.AsStrings(), wrapWithDoubleQuotes), " ")
		objectFileList = "-Wl,--whole-archive " + objectFileList + " -Wl,--no-whole-archive"
	}

	properties := b.buildProperties.Clone()
	properties.Set("compiler.c.elf.flags", properties.Get("compiler.c.elf.flags"))
	properties.Set("compiler.warning_flags", properties.Get("compiler.warning_flags."+b.logger.WarningsLevel()))
	properties.Set("archive_file", coreDotARelPath.String())
	properties.Set("archive_file_path", b.buildArtifacts.coreArchiveFilePath.String())
	properties.Set("object_files", objectFileList)

	command, err := b.prepareCommandForRecipe(properties, "recipe.c.combine.pattern", false)
	if err != nil {
		return err
	}

	return b.execCommand(command)
}
