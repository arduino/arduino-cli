package builder

import (
	"github.com/arduino/arduino-cli/arduino/builder/internal/utils"
	"github.com/arduino/go-paths-helper"
	"github.com/pkg/errors"
)

// ArchiveCompiledFiles fixdoc
func (b *Builder) archiveCompiledFiles(buildPath *paths.Path, archiveFile *paths.Path, objectFilesToArchive paths.PathList) (*paths.Path, error) {
	archiveFilePath := buildPath.JoinPath(archiveFile)

	if b.onlyUpdateCompilationDatabase {
		if b.logger.Verbose() {
			b.logger.Info(tr("Skipping archive creation of: %[1]s", archiveFilePath))
		}
		return archiveFilePath, nil
	}

	if archiveFileStat, err := archiveFilePath.Stat(); err == nil {
		rebuildArchive := false
		for _, objectFile := range objectFilesToArchive {
			objectFileStat, err := objectFile.Stat()
			if err != nil || objectFileStat.ModTime().After(archiveFileStat.ModTime()) {
				// need to rebuild the archive
				rebuildArchive = true
				break
			}
		}

		// something changed, rebuild the core archive
		if rebuildArchive {
			if err := archiveFilePath.Remove(); err != nil {
				return nil, errors.WithStack(err)
			}
		} else {
			if b.logger.Verbose() {
				b.logger.Info(tr("Using previously compiled file: %[1]s", archiveFilePath))
			}
			return archiveFilePath, nil
		}
	}

	for _, objectFile := range objectFilesToArchive {
		properties := b.buildProperties.Clone()
		properties.Set("archive_file", archiveFilePath.Base())
		properties.SetPath("archive_file_path", archiveFilePath)
		properties.SetPath("object_file", objectFile)

		command, err := b.prepareCommandForRecipe(properties, "recipe.ar.pattern", false)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		verboseInfo, _, _, err := utils.ExecCommand(b.logger.Verbose(), b.logger.Stdout(), b.logger.Stderr(), command, utils.ShowIfVerbose /* stdout */, utils.Show /* stderr */)
		if b.logger.Verbose() {
			b.logger.Info(string(verboseInfo))
		}
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return archiveFilePath, nil
}
