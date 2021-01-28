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

package sketch

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/arduino/arduino-cli/arduino/globals"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	paths "github.com/arduino/go-paths-helper"
	"github.com/pkg/errors"
)

// ArchiveSketch FIXMEDOC
func ArchiveSketch(ctx context.Context, req *rpc.ArchiveSketchReq) (*rpc.ArchiveSketchResp, error) {
	// sketchName is the name of the sketch without extension, for example "MySketch"
	var sketchName string

	sketchPath := paths.New(req.SketchPath)
	if sketchPath == nil {
		sketchPath = paths.New(".")
	}

	sketchPath, err := sketchPath.Clean().Abs()
	if err != nil {
		return nil, fmt.Errorf("Error getting absolute sketch path %v", err)
	}

	// Make sketchPath point to the sketch folder
	if sketchPath.IsNotDir() {
		sketchPath = sketchPath.Parent()
	}

	sketchName = sketchPath.Base()

	// sketchMainFile is the path to the sketch main file
	var sketchMainFile *paths.Path

	for ext := range globals.MainFileValidExtensions {
		candidateSketchMainFile := sketchPath.Join(sketchPath.Base() + ext)
		if candidateSketchMainFile.Exist() {
			if sketchMainFile == nil {
				sketchMainFile = candidateSketchMainFile
			} else {
				return nil, errors.Errorf("multiple main sketch files found (%v, %v)",
					sketchMainFile,
					candidateSketchMainFile,
				)
			}
		}
	}

	// Checks if it's really a sketch
	if sketchMainFile == nil {
		return nil, fmt.Errorf("specified path is not a sketch: %v", sketchPath.String())
	}

	archivePath := paths.New(req.ArchivePath)
	if archivePath == nil {
		archivePath = sketchPath.Parent()
	}

	archivePath, err = archivePath.Clean().Abs()
	if err != nil {
		return nil, fmt.Errorf("Error getting absolute archive path %v", err)
	}

	// Makes archivePath point to a zip file
	if archivePath.IsDir() {
		archivePath = archivePath.Join(sketchName + ".zip")
	} else if archivePath.Ext() == "" {
		archivePath = paths.New(archivePath.String() + ".zip")
	}

	if archivePath.Exist() {
		return nil, fmt.Errorf("archive already exists")
	}

	filesToZip, err := sketchPath.ReadDirRecursive()
	if err != nil {
		return nil, fmt.Errorf("Error retrieving sketch files: %v", err)
	}
	filesToZip.FilterOutDirs()

	archive, err := archivePath.Create()
	if err != nil {
		return nil, fmt.Errorf("Error creating archive: %v", err)
	}
	defer archive.Close()

	zipWriter := zip.NewWriter(archive)
	defer zipWriter.Close()

	for _, f := range filesToZip {

		if !req.IncludeBuildDir {
			filePath, err := sketchPath.Parent().RelTo(f)
			if err != nil {
				return nil, fmt.Errorf("Error calculating relative file path: %v", err)
			}

			// Skips build folder
			if strings.HasPrefix(filePath.String(), sketchName+string(filepath.Separator)+"build") {
				continue
			}
		}

		// We get the parent path since we want the archive to unpack as a folder.
		// If we don't do this the archive would contain all the sketch files as top level.
		err = addFileToSketchArchive(zipWriter, f, sketchPath.Parent())
		if err != nil {
			return nil, fmt.Errorf("Error adding file to archive: %v", err)
		}
	}

	return &rpc.ArchiveSketchResp{}, nil
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
