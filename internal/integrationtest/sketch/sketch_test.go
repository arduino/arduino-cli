// This file is part of arduino-cli.
//
// Copyright 2022 ARDUINO SA (http://www.arduino.cc/)
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

package sketch_test

import (
	"archive/zip"
	"fmt"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

type archiveTest struct {
	SubTestName         string
	SketchPathParam     string
	TargetPathParam     string
	WorkingDir          *paths.Path
	ExpectedArchivePath *paths.Path
}

func TestSketchNew(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Create a test sketch in current directory
	currentPath := cli.WorkingDir()
	sketchName := "SketchNewIntegrationTest"
	currentSketchPath := currentPath.Join(sketchName)
	stdout, _, err := cli.Run("sketch", "new", sketchName)
	require.NoError(t, err)
	require.Contains(t, string(stdout), "Sketch created in: "+currentSketchPath.String())
	require.FileExists(t, currentSketchPath.Join(sketchName).String()+".ino")

	// Create a test sketch in current directory but using an absolute path
	sketchName = "SketchNewIntegrationTestAbsolute"
	currentSketchPath = currentPath.Join(sketchName)
	stdout, _, err = cli.Run("sketch", "new", currentSketchPath.String())
	require.NoError(t, err)
	require.Contains(t, string(stdout), "Sketch created in: "+currentSketchPath.String())
	require.FileExists(t, currentSketchPath.Join(sketchName).String()+".ino")

	// Create a test sketch in current directory subpath but using an absolute path
	sketchName = "SketchNewIntegrationTestSubpath"
	sketchSubpath := paths.New("subpath", sketchName)
	currentSketchPath = currentPath.JoinPath(sketchSubpath)
	stdout, _, err = cli.Run("sketch", "new", sketchSubpath.String())
	require.NoError(t, err)
	require.Contains(t, string(stdout), "Sketch created in: "+currentSketchPath.String())
	require.FileExists(t, currentSketchPath.Join(sketchName).String()+".ino")

	// Create a test sketch in current directory using .ino extension
	sketchName = "SketchNewIntegrationTestDotIno"
	currentSketchPath = currentPath.Join(sketchName)
	stdout, _, err = cli.Run("sketch", "new", sketchName+".ino")
	require.NoError(t, err)
	require.Contains(t, string(stdout), "Sketch created in: "+currentSketchPath.String())
	require.FileExists(t, currentSketchPath.Join(sketchName).String()+".ino")
}

func verifyZipContainsSketchExcludingBuildDir(t *testing.T, files []*zip.File) {
	require.Len(t, files, 8)
	require.Equal(t, paths.New("sketch_simple", "doc.txt").String(), files[0].Name)
	require.Equal(t, paths.New("sketch_simple", "header.h").String(), files[1].Name)
	require.Equal(t, paths.New("sketch_simple", "merged_sketch.txt").String(), files[2].Name)
	require.Equal(t, paths.New("sketch_simple", "old.pde").String(), files[3].Name)
	require.Equal(t, paths.New("sketch_simple", "other.ino").String(), files[4].Name)
	require.Equal(t, paths.New("sketch_simple", "s_file.S").String(), files[5].Name)
	require.Equal(t, paths.New("sketch_simple", "sketch_simple.ino").String(), files[6].Name)
	require.Equal(t, paths.New("sketch_simple", "src", "helper.h").String(), files[7].Name)
}

func verifyZipContainsSketchIncludingBuildDir(t *testing.T, files []*zip.File) {
	require.Len(t, files, 13)
	require.Equal(t, paths.New("sketch_simple", "doc.txt").String(), files[5].Name)
	require.Equal(t, paths.New("sketch_simple", "header.h").String(), files[6].Name)
	require.Equal(t, paths.New("sketch_simple", "merged_sketch.txt").String(), files[7].Name)
	require.Equal(t, paths.New("sketch_simple", "old.pde").String(), files[8].Name)
	require.Equal(t, paths.New("sketch_simple", "other.ino").String(), files[9].Name)
	require.Equal(t, paths.New("sketch_simple", "s_file.S").String(), files[10].Name)
	require.Equal(t, paths.New("sketch_simple", "sketch_simple.ino").String(), files[11].Name)
	require.Equal(t, paths.New("sketch_simple", "src", "helper.h").String(), files[12].Name)
	require.Equal(t, paths.New("sketch_simple", "build", "adafruit.samd.adafruit_feather_m0", "sketch_simple.ino.hex").String(), files[0].Name)
	require.Equal(t, paths.New("sketch_simple", "build", "adafruit.samd.adafruit_feather_m0", "sketch_simple.ino.map").String(), files[1].Name)
	require.Equal(t, paths.New("sketch_simple", "build", "arduino.avr.uno", "sketch_simple.ino.eep").String(), files[2].Name)
	require.Equal(t, paths.New("sketch_simple", "build", "arduino.avr.uno", "sketch_simple.ino.hex").String(), files[3].Name)
	require.Equal(t, paths.New("sketch_simple", "build", "arduino.avr.uno", "sketch_simple.ino.with_bootloader.hex").String(), files[4].Name)
}

func TestSketchArchive(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	sketchSimple := cli.CopySketch("sketch_simple")

	// Creates a folder where to save the zip
	archivesFolder := cli.WorkingDir().Join("my_archives")
	require.NoError(t, archivesFolder.Mkdir())

	archiveTests := []archiveTest{
		{
			SubTestName:         "ArchiveNoArgs",
			SketchPathParam:     "",
			TargetPathParam:     "",
			WorkingDir:          sketchSimple,
			ExpectedArchivePath: env.RootDir().Join("sketch_simple.zip"),
		},
		{
			SubTestName:         "ArchiveDotArg",
			SketchPathParam:     ".",
			TargetPathParam:     "",
			WorkingDir:          sketchSimple,
			ExpectedArchivePath: env.RootDir().Join("sketch_simple.zip"),
		},
		{
			SubTestName:         "DotArgRelativeZipPath",
			SketchPathParam:     ".",
			TargetPathParam:     "../my_archives",
			WorkingDir:          sketchSimple,
			ExpectedArchivePath: archivesFolder.Join("sketch_simple.zip"),
		},
		{
			SubTestName:         "DotArgAbsoluteZiptPath",
			SketchPathParam:     ".",
			TargetPathParam:     archivesFolder.String(),
			WorkingDir:          sketchSimple,
			ExpectedArchivePath: archivesFolder.Join("sketch_simple.zip"),
		},
		{
			SubTestName:         "ArchiveDotArgRelativeZipPathAndNameWithoutExtension",
			SketchPathParam:     ".",
			TargetPathParam:     "../my_archives/my_custom_sketch",
			WorkingDir:          sketchSimple,
			ExpectedArchivePath: archivesFolder.Join("my_custom_sketch.zip"),
		},
		{
			SubTestName:         "ArchiveDotArgAbsoluteZipPathAndNameWithoutExtension",
			SketchPathParam:     ".",
			TargetPathParam:     archivesFolder.Join("my_custom_sketch").String(),
			WorkingDir:          sketchSimple,
			ExpectedArchivePath: archivesFolder.Join("my_custom_sketch.zip"),
		},
		{
			SubTestName:         "ArchiveDotArgCustomZipPathAndNameWithExtension",
			SketchPathParam:     ".",
			TargetPathParam:     archivesFolder.Join("my_custom_sketch.zip").String(),
			WorkingDir:          sketchSimple,
			ExpectedArchivePath: archivesFolder.Join("my_custom_sketch.zip"),
		},
		{
			SubTestName:         "ArchiveRelativeSketchPath",
			SketchPathParam:     "./sketch_simple",
			TargetPathParam:     "",
			WorkingDir:          env.RootDir(),
			ExpectedArchivePath: env.RootDir().Join("sketch_simple.zip"),
		},
		{
			SubTestName:         "ArchiveAbsoluteSketchPath",
			SketchPathParam:     env.RootDir().Join("sketch_simple").String(),
			TargetPathParam:     "",
			WorkingDir:          env.RootDir(),
			ExpectedArchivePath: env.RootDir().Join("sketch_simple.zip"),
		},
		{
			SubTestName:         "ArchiveRelativeSketchPathWithRelativeZipPath",
			SketchPathParam:     "./sketch_simple",
			TargetPathParam:     "./my_archives",
			WorkingDir:          env.RootDir(),
			ExpectedArchivePath: archivesFolder.Join("sketch_simple.zip"),
		},
		{
			SubTestName:         "ArchiveRelativeSketchPathWithAbsoluteZipPath",
			SketchPathParam:     "./sketch_simple",
			TargetPathParam:     archivesFolder.String(),
			WorkingDir:          env.RootDir(),
			ExpectedArchivePath: archivesFolder.Join("sketch_simple.zip"),
		},
		{
			SubTestName:         "ArchiveRelativeSketchPathWithRelativeZipPathAndNameWithoutExtension",
			SketchPathParam:     "./sketch_simple",
			TargetPathParam:     "./my_archives/my_custom_sketch",
			WorkingDir:          env.RootDir(),
			ExpectedArchivePath: archivesFolder.Join("my_custom_sketch.zip"),
		},
		{
			SubTestName:         "ArchiveRelativeSketchPathWithRelativeZipPathAndNameWithExtension",
			SketchPathParam:     "./sketch_simple",
			TargetPathParam:     "./my_archives/my_custom_sketch.zip",
			WorkingDir:          env.RootDir(),
			ExpectedArchivePath: archivesFolder.Join("my_custom_sketch.zip"),
		},
		{
			SubTestName:         "ArchiveRelativeSketchPathWithAbsoluteZipPathAndNameWithoutExtension",
			SketchPathParam:     "./sketch_simple",
			TargetPathParam:     archivesFolder.Join("my_custom_sketch").String(),
			WorkingDir:          env.RootDir(),
			ExpectedArchivePath: archivesFolder.Join("my_custom_sketch.zip"),
		},
		{
			SubTestName:         "ArchiveRelativeSketchPathWithAbsoluteZipPathAndNameWithoutExtension",
			SketchPathParam:     "./sketch_simple",
			TargetPathParam:     archivesFolder.Join("my_custom_sketch.zip").String(),
			WorkingDir:          env.RootDir(),
			ExpectedArchivePath: archivesFolder.Join("my_custom_sketch.zip"),
		},
		{
			SubTestName:         "ArchiveAbsoluteSketchPathWithRelativeZipPath",
			SketchPathParam:     cli.WorkingDir().Join("sketch_simple").String(),
			TargetPathParam:     "./my_archives",
			WorkingDir:          env.RootDir(),
			ExpectedArchivePath: archivesFolder.Join("sketch_simple.zip"),
		},
		{
			SubTestName:         "ArchiveAbsoluteSketchPathWithAbsoluteZipPath",
			SketchPathParam:     env.RootDir().Join("sketch_simple").String(),
			TargetPathParam:     archivesFolder.String(),
			WorkingDir:          sketchSimple,
			ExpectedArchivePath: archivesFolder.Join("sketch_simple.zip"),
		},
		{
			SubTestName:         "ArchiveAbsoluteSketchPathWithRelativeZipPathAndNameWithoutExtension",
			SketchPathParam:     cli.WorkingDir().Join("sketch_simple").String(),
			TargetPathParam:     "./my_archives/my_custom_sketch",
			WorkingDir:          env.RootDir(),
			ExpectedArchivePath: archivesFolder.Join("my_custom_sketch.zip"),
		},
		{
			SubTestName:         "ArchiveAbsoluteSketchPathWithRelativeZipPathAndNameWithExtension",
			SketchPathParam:     cli.WorkingDir().Join("sketch_simple").String(),
			TargetPathParam:     "./my_archives/my_custom_sketch.zip",
			WorkingDir:          env.RootDir(),
			ExpectedArchivePath: archivesFolder.Join("my_custom_sketch.zip"),
		},
		{
			SubTestName:         "ArchiveAbsoluteSketchPathWithAbsoluteZipPathAndNameWithoutExtension",
			SketchPathParam:     env.RootDir().Join("sketch_simple").String(),
			TargetPathParam:     archivesFolder.Join("my_custom_sketch").String(),
			WorkingDir:          sketchSimple,
			ExpectedArchivePath: archivesFolder.Join("my_custom_sketch.zip"),
		},
		{
			SubTestName:         "ArchiveAbsoluteSketchPathWithAbsoluteZipPathAndNameWithExtension",
			SketchPathParam:     env.RootDir().Join("sketch_simple").String(),
			TargetPathParam:     archivesFolder.Join("my_custom_sketch.zip").String(),
			WorkingDir:          sketchSimple,
			ExpectedArchivePath: archivesFolder.Join("my_custom_sketch.zip"),
		},
	}

	for _, test := range archiveTests {
		t.Run(fmt.Sprint(test.SubTestName), func(t *testing.T) {
			var err error
			cli.SetWorkingDir(test.WorkingDir)
			if test.TargetPathParam == "" {
				if test.SketchPathParam == "" {
					_, _, err = cli.Run("sketch", "archive")
				} else {
					_, _, err = cli.Run("sketch", "archive", test.SketchPathParam)
				}
			} else {
				_, _, err = cli.Run("sketch", "archive", test.SketchPathParam, test.TargetPathParam)
			}
			require.NoError(t, err)

			archive, err := zip.OpenReader(test.ExpectedArchivePath.String())
			require.NoError(t, err)
			defer require.NoError(t, archive.Close())
			defer require.NoError(t, test.ExpectedArchivePath.Remove())
			verifyZipContainsSketchExcludingBuildDir(t, archive.File)
		})
		t.Run(fmt.Sprint(test.SubTestName+"WithIncludeBuildDirFlag"), func(t *testing.T) {
			var err error
			cli.SetWorkingDir(test.WorkingDir)
			if test.TargetPathParam == "" {
				if test.SketchPathParam == "" {
					_, _, err = cli.Run("sketch", "archive", "--include-build-dir")
				} else {
					_, _, err = cli.Run("sketch", "archive", test.SketchPathParam, "--include-build-dir")
				}
			} else {
				_, _, err = cli.Run("sketch", "archive", test.SketchPathParam, test.TargetPathParam, "--include-build-dir")
			}
			require.NoError(t, err)

			archive, err := zip.OpenReader(test.ExpectedArchivePath.String())
			require.NoError(t, err)
			defer require.NoError(t, archive.Close())
			defer require.NoError(t, test.ExpectedArchivePath.Remove())
			verifyZipContainsSketchIncludingBuildDir(t, archive.File)
		})
	}
}

func TestSketchArchiveWithPdeMainFile(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	sketchName := "sketch_pde_main_file"
	sketchDir := cli.CopySketch(sketchName)
	sketchFile := sketchDir.Join(sketchName + ".pde")
	relPath, err := sketchFile.RelFrom(sketchDir)
	require.NoError(t, err)
	cli.SetWorkingDir(sketchDir)
	_, stderr, err := cli.Run("sketch", "archive")
	require.NoError(t, err)
	require.Contains(t, string(stderr), "Sketches with .pde extension are deprecated, please rename the following files to .ino")
	require.Contains(t, string(stderr), relPath.String())
	cli.SetWorkingDir(env.RootDir())

	archive, err := zip.OpenReader(cli.WorkingDir().Join(sketchName + ".zip").String())
	require.NoError(t, err)
	defer require.NoError(t, archive.Close())
	require.Contains(t, archive.File[0].Name, paths.New(sketchName, sketchName+".pde").String())
}

func TestSketchArchiveWithMultipleMainFiles(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	sketchName := "sketch_multiple_main_files"
	sketchDir := cli.CopySketch(sketchName)
	sketchFile := sketchDir.Join(sketchName + ".pde")
	relPath, err := sketchFile.RelFrom(sketchDir)
	require.NoError(t, err)
	cli.SetWorkingDir(sketchDir)
	_, stderr, err := cli.Run("sketch", "archive")
	require.Error(t, err)
	require.Contains(t, string(stderr), "Sketches with .pde extension are deprecated, please rename the following files to .ino")
	require.Contains(t, string(stderr), relPath.String())
	require.Contains(t, string(stderr), "Error archiving: Can't open sketch: multiple main sketch files found")
}

func TestSketchArchiveCaseMismatchFails(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	sketchName := "ArchiveSketchCaseMismatch"
	sketchPath := cli.SketchbookDir().Join(sketchName)

	_, _, err := cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	// Rename main .ino file so casing is different from sketch name
	require.NoError(t, sketchPath.Join(sketchName+".ino").Rename(sketchPath.Join(strings.ToLower(sketchName)+".ino")))

	_, stderr, err := cli.Run("sketch", "archive", sketchPath.String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "Error archiving: Can't open sketch:")
}

func TestSketchNewDotArgOverwrite(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	sketchNew := "SketchNewDotArgOverwrite"
	sketchPath := cli.SketchbookDir().Join(sketchNew)
	require.NoError(t, sketchPath.MkdirAll())

	cli.SetWorkingDir(sketchPath)
	require.NoFileExists(t, sketchPath.Join(sketchNew+".ino").String())
	// Create a new sketch
	_, _, err := cli.Run("sketch", "new", ".")
	require.NoError(t, err)

	require.FileExists(t, sketchPath.Join(sketchNew+".ino").String())
	// Tries to overwrite the existing sketch with a new one, but it should fail
	_, stderr, err := cli.Run("sketch", "new", ".")
	require.Error(t, err)
	require.Contains(t, string(stderr), ".ino file already exists")
}
