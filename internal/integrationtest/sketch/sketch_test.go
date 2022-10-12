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
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

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

func TestSketchArchiveNoArgs(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	cli.SetWorkingDir(cli.CopySketch("sketch_simple"))

	_, _, err := cli.Run("sketch", "archive")
	require.NoError(t, err)

	cli.SetWorkingDir(env.RootDir())

	archive, err := zip.OpenReader(cli.WorkingDir().Join("sketch_simple.zip").String())
	require.NoError(t, err)
	defer require.NoError(t, archive.Close())
	verifyZipContainsSketchExcludingBuildDir(t, archive.File)
}

func TestSketchArchiveDotArg(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	cli.SetWorkingDir(cli.CopySketch("sketch_simple"))

	_, _, err := cli.Run("sketch", "archive", ".")
	require.NoError(t, err)

	cli.SetWorkingDir(env.RootDir())

	archive, err := zip.OpenReader(cli.WorkingDir().Join("sketch_simple.zip").String())
	require.NoError(t, err)
	defer require.NoError(t, archive.Close())
	verifyZipContainsSketchExcludingBuildDir(t, archive.File)
}

func TestSketchDotArgRelativeZipPath(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Creates a folder where to save the zip
	archivesFolder := cli.WorkingDir().Join("my_archives")
	require.NoError(t, archivesFolder.Mkdir())

	cli.SetWorkingDir(cli.CopySketch("sketch_simple"))
	_, _, err := cli.Run("sketch", "archive", ".", "../my_archives")
	require.NoError(t, err)

	archive, err := zip.OpenReader(archivesFolder.Join("sketch_simple.zip").String())
	require.NoError(t, err)
	defer require.NoError(t, archive.Close())
	verifyZipContainsSketchExcludingBuildDir(t, archive.File)
}

func TestSketchDotArgAbsoluteZipPath(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Creates a folder where to save the zip
	archivesFolder := cli.WorkingDir().Join("my_archives")
	require.NoError(t, archivesFolder.Mkdir())

	cli.SetWorkingDir(cli.CopySketch("sketch_simple"))
	_, _, err := cli.Run("sketch", "archive", ".", archivesFolder.String())
	require.NoError(t, err)
	archive, err := zip.OpenReader(archivesFolder.Join("sketch_simple.zip").String())
	require.NoError(t, err)
	defer require.NoError(t, archive.Close())
	verifyZipContainsSketchExcludingBuildDir(t, archive.File)
}

func TestSketchArchiveDotArgRelativeZipPathAndNameWithoutExtension(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Creates a folder where to save the zip
	archivesFolder := cli.WorkingDir().Join("my_archives")
	require.NoError(t, archivesFolder.Mkdir())

	cli.SetWorkingDir(cli.CopySketch("sketch_simple"))
	_, _, err := cli.Run("sketch", "archive", ".", "../my_archives/my_custom_sketch")
	require.NoError(t, err)

	archive, err := zip.OpenReader(archivesFolder.Join("my_custom_sketch.zip").String())
	require.NoError(t, err)
	defer require.NoError(t, archive.Close())
	verifyZipContainsSketchExcludingBuildDir(t, archive.File)
}

func TestSketchArchiveDotArgAbsoluteZipPathAndNameWithoutExtension(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Creates a folder where to save the zip
	archivesFolder := cli.WorkingDir().Join("my_archives")
	require.NoError(t, archivesFolder.Mkdir())

	cli.SetWorkingDir(cli.CopySketch("sketch_simple"))
	_, _, err := cli.Run("sketch", "archive", ".", archivesFolder.Join("my_custom_sketch").String())
	require.NoError(t, err)

	archive, err := zip.OpenReader(archivesFolder.Join("my_custom_sketch.zip").String())
	require.NoError(t, err)
	defer require.NoError(t, archive.Close())
	verifyZipContainsSketchExcludingBuildDir(t, archive.File)
}

func TestSketchArchiveDotArgCustomZipPathAndNameWithExtension(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Creates a folder where to save the zip
	archivesFolder := cli.WorkingDir().Join("my_archives")
	require.NoError(t, archivesFolder.Mkdir())

	cli.SetWorkingDir(cli.CopySketch("sketch_simple"))
	_, _, err := cli.Run("sketch", "archive", ".", archivesFolder.Join("my_custom_sketch.zip").String())
	require.NoError(t, err)

	archive, err := zip.OpenReader(archivesFolder.Join("my_custom_sketch.zip").String())
	require.NoError(t, err)
	defer require.NoError(t, archive.Close())
	verifyZipContainsSketchExcludingBuildDir(t, archive.File)
}

func TestSketchArchiveRelativeSketchPath(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_ = cli.CopySketch("sketch_simple")
	_, _, err := cli.Run("sketch", "archive", "./sketch_simple")
	require.NoError(t, err)

	archive, err := zip.OpenReader(cli.WorkingDir().Join("sketch_simple.zip").String())
	require.NoError(t, err)
	defer require.NoError(t, archive.Close())
	verifyZipContainsSketchExcludingBuildDir(t, archive.File)
}

func TestSketchArchiveAbsoluteSketchPath(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	cli.SetWorkingDir(cli.CopySketch("sketch_simple"))
	_, _, err := cli.Run("sketch", "archive", env.RootDir().Join("sketch_simple").String())
	require.NoError(t, err)

	archive, err := zip.OpenReader(env.RootDir().Join("sketch_simple.zip").String())
	require.NoError(t, err)
	defer require.NoError(t, archive.Close())
	verifyZipContainsSketchExcludingBuildDir(t, archive.File)
}

func TestSketchArchiveRelativeSketchPathWithRelativeZipPath(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_ = cli.CopySketch("sketch_simple")
	// Creates a folder where to save the zip
	archivesFolder := cli.WorkingDir().Join("my_archives")
	require.NoError(t, archivesFolder.Mkdir())

	_, _, err := cli.Run("sketch", "archive", "./sketch_simple", "./my_archives")
	require.NoError(t, err)

	archive, err := zip.OpenReader(cli.WorkingDir().Join("my_archives", "sketch_simple.zip").String())
	require.NoError(t, err)
	defer require.NoError(t, archive.Close())
	verifyZipContainsSketchExcludingBuildDir(t, archive.File)
}

func TestSketchArchiveRelativeSketchPathWithAbsoluteZipPath(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_ = cli.CopySketch("sketch_simple")
	// Creates a folder where to save the zip
	archivesFolder := cli.WorkingDir().Join("my_archives")
	require.NoError(t, archivesFolder.Mkdir())

	_, _, err := cli.Run("sketch", "archive", "./sketch_simple", archivesFolder.String())
	require.NoError(t, err)

	archive, err := zip.OpenReader(archivesFolder.Join("sketch_simple.zip").String())
	require.NoError(t, err)
	defer require.NoError(t, archive.Close())
	verifyZipContainsSketchExcludingBuildDir(t, archive.File)
}

func TestSketchArchiveRelativeSketchPathWithRelativeZipPathAndNameWithoutExtension(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_ = cli.CopySketch("sketch_simple")
	// Creates a folder where to save the zip
	archivesFolder := cli.WorkingDir().Join("my_archives")
	require.NoError(t, archivesFolder.Mkdir())

	_, _, err := cli.Run("sketch", "archive", "./sketch_simple", "./my_archives/my_custom_sketch")
	require.NoError(t, err)

	archive, err := zip.OpenReader(cli.WorkingDir().Join("my_archives", "my_custom_sketch.zip").String())
	require.NoError(t, err)
	defer require.NoError(t, archive.Close())
	verifyZipContainsSketchExcludingBuildDir(t, archive.File)
}

func TestSketchArchiveRelativeSketchPathWithRelativeZipPathAndNameWithExtension(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_ = cli.CopySketch("sketch_simple")
	// Creates a folder where to save the zip
	archivesFolder := cli.WorkingDir().Join("my_archives")
	require.NoError(t, archivesFolder.Mkdir())

	_, _, err := cli.Run("sketch", "archive", "./sketch_simple", "./my_archives/my_custom_sketch.zip")
	require.NoError(t, err)

	archive, err := zip.OpenReader(cli.WorkingDir().Join("my_archives", "my_custom_sketch.zip").String())
	require.NoError(t, err)
	defer require.NoError(t, archive.Close())
	verifyZipContainsSketchExcludingBuildDir(t, archive.File)
}

func TestSketchArchiveRelativeSketchPathWithAbsoluteZipPathAndNameWithoutExtension(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_ = cli.CopySketch("sketch_simple")
	// Creates a folder where to save the zip
	archivesFolder := cli.WorkingDir().Join("my_archives")
	require.NoError(t, archivesFolder.Mkdir())

	_, _, err := cli.Run("sketch", "archive", "./sketch_simple", archivesFolder.Join("my_custom_sketch").String())
	require.NoError(t, err)

	archive, err := zip.OpenReader(archivesFolder.Join("my_custom_sketch.zip").String())
	require.NoError(t, err)
	defer require.NoError(t, archive.Close())
	verifyZipContainsSketchExcludingBuildDir(t, archive.File)
}

func TestSketchArchiveRelativeSketchPathWithAbsoluteZipPathAndNameWithExtension(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_ = cli.CopySketch("sketch_simple")
	// Creates a folder where to save the zip
	archivesFolder := cli.WorkingDir().Join("my_archives")
	require.NoError(t, archivesFolder.Mkdir())

	_, _, err := cli.Run("sketch", "archive", "./sketch_simple", archivesFolder.Join("my_custom_sketch.zip").String())
	require.NoError(t, err)

	archive, err := zip.OpenReader(archivesFolder.Join("my_custom_sketch.zip").String())
	require.NoError(t, err)
	defer require.NoError(t, archive.Close())
	verifyZipContainsSketchExcludingBuildDir(t, archive.File)
}

func TestSketchArchiveAbsoluteSketchPathWithRelativeZipPath(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_ = cli.CopySketch("sketch_simple")
	// Creates a folder where to save the zip
	archivesFolder := cli.WorkingDir().Join("my_archives")
	require.NoError(t, archivesFolder.Mkdir())

	_, _, err := cli.Run("sketch", "archive", cli.WorkingDir().Join("sketch_simple").String(), "./my_archives")
	require.NoError(t, err)

	archive, err := zip.OpenReader(cli.WorkingDir().Join("my_archives", "sketch_simple.zip").String())
	require.NoError(t, err)
	defer require.NoError(t, archive.Close())
	verifyZipContainsSketchExcludingBuildDir(t, archive.File)
}

func TestSketchArchiveAbsoluteSketchPathWithAbsoluteZipPath(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Creates a folder where to save the zip
	archivesFolder := cli.WorkingDir().Join("my_archives")
	require.NoError(t, archivesFolder.Mkdir())

	cli.SetWorkingDir(cli.CopySketch("sketch_simple"))
	_, _, err := cli.Run("sketch", "archive", env.RootDir().Join("sketch_simple").String(), archivesFolder.String())
	require.NoError(t, err)

	archive, err := zip.OpenReader(archivesFolder.Join("sketch_simple.zip").String())
	require.NoError(t, err)
	defer require.NoError(t, archive.Close())
	verifyZipContainsSketchExcludingBuildDir(t, archive.File)
}

func TestSketchArchiveAbsoluteSketchPathWithRelativeZipPathAndNameWithoutExtension(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_ = cli.CopySketch("sketch_simple")
	// Creates a folder where to save the zip
	archivesFolder := cli.WorkingDir().Join("my_archives")
	require.NoError(t, archivesFolder.Mkdir())

	_, _, err := cli.Run("sketch", "archive", cli.WorkingDir().Join("sketch_simple").String(), "./my_archives/my_custom_sketch")
	require.NoError(t, err)

	archive, err := zip.OpenReader(cli.WorkingDir().Join("my_archives", "my_custom_sketch.zip").String())
	require.NoError(t, err)
	defer require.NoError(t, archive.Close())
	verifyZipContainsSketchExcludingBuildDir(t, archive.File)
}

func TestSketchArchiveAbsoluteSketchPathWithRelativeZipPathAndNameWithExtension(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_ = cli.CopySketch("sketch_simple")
	// Creates a folder where to save the zip
	archivesFolder := cli.WorkingDir().Join("my_archives")
	require.NoError(t, archivesFolder.Mkdir())

	_, _, err := cli.Run("sketch", "archive", cli.WorkingDir().Join("sketch_simple").String(), "./my_archives/my_custom_sketch.zip")
	require.NoError(t, err)

	archive, err := zip.OpenReader(cli.WorkingDir().Join("my_archives", "my_custom_sketch.zip").String())
	require.NoError(t, err)
	defer require.NoError(t, archive.Close())
	verifyZipContainsSketchExcludingBuildDir(t, archive.File)
}

func TestSketchArchiveAbsoluteSketchPathWithAbsoluteZipPathAndNameWithoutExtension(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Creates a folder where to save the zip
	archivesFolder := cli.WorkingDir().Join("my_archives")
	require.NoError(t, archivesFolder.Mkdir())

	cli.SetWorkingDir(cli.CopySketch("sketch_simple"))
	_, _, err := cli.Run("sketch", "archive", env.RootDir().Join("sketch_simple").String(), archivesFolder.Join("my_custom_sketch").String())
	require.NoError(t, err)

	archive, err := zip.OpenReader(archivesFolder.Join("my_custom_sketch.zip").String())
	require.NoError(t, err)
	defer require.NoError(t, archive.Close())
	verifyZipContainsSketchExcludingBuildDir(t, archive.File)
}
