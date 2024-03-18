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

package commands

import (
	"archive/zip"
	"context"
	"io"
	"path/filepath"
	"strings"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/internal/arduino/sketch"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	paths "github.com/arduino/go-paths-helper"
)

// ArchiveSketch FIXMEDOC
func (s *arduinoCoreServerImpl) ArchiveSketch(ctx context.Context, req *rpc.ArchiveSketchRequest) (*rpc.ArchiveSketchResponse, error) {
	// sketchName is the name of the sketch without extension, for example "MySketch"
	var sketchName string

	sketchPath := paths.New(req.GetSketchPath())
	if sketchPath == nil {
		sketchPath = paths.New(".")
	}

	sk, err := sketch.New(sketchPath)
	if err != nil {
		return nil, &cmderrors.CantOpenSketchError{Cause: err}
	}

	sketchPath = sk.FullPath
	sketchName = sk.Name

	archivePath := paths.New(req.GetArchivePath())
	if archivePath == nil {
		archivePath = sketchPath.Parent()
	}

	archivePath, err = archivePath.Clean().Abs()
	if err != nil {
		return nil, &cmderrors.PermissionDeniedError{Message: tr("Error getting absolute path of sketch archive"), Cause: err}
	}

	// Makes archivePath point to a zip file
	if archivePath.IsDir() {
		archivePath = archivePath.Join(sketchName + ".zip")
	} else if archivePath.Ext() == "" {
		archivePath = paths.New(archivePath.String() + ".zip")
	}

	if !req.GetOverwrite() {
		if archivePath.Exist() {
			return nil, &cmderrors.InvalidArgumentError{Message: tr("Archive already exists")}
		}
	}

	filesToZip, err := sketchPath.ReadDirRecursive()
	if err != nil {
		return nil, &cmderrors.PermissionDeniedError{Message: tr("Error reading sketch files"), Cause: err}
	}
	filesToZip.FilterOutDirs()

	archive, err := archivePath.Create()
	if err != nil {
		return nil, &cmderrors.PermissionDeniedError{Message: tr("Error creating sketch archive"), Cause: err}
	}
	defer archive.Close()

	zipWriter := zip.NewWriter(archive)
	defer zipWriter.Close()

	for _, f := range filesToZip {

		if !req.GetIncludeBuildDir() {
			filePath, err := sketchPath.Parent().RelTo(f)
			if err != nil {
				return nil, &cmderrors.PermissionDeniedError{Message: tr("Error calculating relative file path"), Cause: err}
			}

			// Skips build folder
			if strings.HasPrefix(filePath.String(), sketchName+string(filepath.Separator)+"build") {
				continue
			}
		}

		// We get the parent path since we want the archive to unpack as a folder.
		// If we don't do this the archive would contain all the sketch files as top level.
		if err := addFileToSketchArchive(zipWriter, f, sketchPath.Parent()); err != nil {
			return nil, &cmderrors.PermissionDeniedError{Message: tr("Error adding file to sketch archive"), Cause: err}
		}
	}

	return &rpc.ArchiveSketchResponse{}, nil
}

// Adds a single file to an existing zip file
func addFileToSketchArchive(zipWriter *zip.Writer, filePath, sketchPath *paths.Path) error {
	f, err := filePath.Open()
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	filePath, err = sketchPath.RelTo(filePath)
	if err != nil {
		return err
	}

	header.Name = filePath.String()
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, f)
	return err
}
