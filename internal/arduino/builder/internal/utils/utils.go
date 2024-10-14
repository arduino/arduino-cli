// This file is part of arduino-cli.
//
// Copyright 2023 ARDUINO SA (http://www.arduino.cc/)
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

package utils

import (
	"os"
	"strings"
	"unicode"

	f "github.com/arduino/arduino-cli/internal/algorithms"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// ObjFileIsUpToDate fixdoc
func ObjFileIsUpToDate(sourceFile, objectFile, dependencyFile *paths.Path) (bool, error) {
	logrus.Debugf("Checking previous results for %v (result = %v, dep = %v)", sourceFile, objectFile, dependencyFile)
	if objectFile == nil || dependencyFile == nil {
		logrus.Debugf("Object file or dependency file not provided")
		return false, nil
	}

	sourceFile = sourceFile.Clean()
	sourceFileStat, err := sourceFile.Stat()
	if err != nil {
		logrus.Debugf("Could not stat source file: %s", err)
		return false, err
	}

	objectFile = objectFile.Clean()
	objectFileStat, err := objectFile.Stat()
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Debugf("Object file not found: %v", objectFile)
			return false, nil
		}
		logrus.Debugf("Could not stat object file: %s", err)
		return false, err
	}

	dependencyFile = dependencyFile.Clean()
	dependencyFileStat, err := dependencyFile.Stat()
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Debugf("Dependency file not found: %v", dependencyFile)
			return false, nil
		}
		logrus.Debugf("Could not stat dependency file: %s", err)
		return false, err
	}

	if sourceFileStat.ModTime().After(objectFileStat.ModTime()) {
		logrus.Debugf("%v newer than %v", sourceFile, objectFile)
		return false, nil
	}
	if sourceFileStat.ModTime().After(dependencyFileStat.ModTime()) {
		logrus.Debugf("%v newer than %v", sourceFile, dependencyFile)
		return false, nil
	}

	rows, err := dependencyFile.ReadFileAsLines()
	if err != nil {
		logrus.Debugf("Could not read dependency file: %s", dependencyFile)
		return false, err
	}

	rows = f.Map(rows, removeEndingBackSlash)
	rows = f.Map(rows, strings.TrimSpace)
	rows = f.Map(rows, unescapeDep)
	rows = f.Filter(rows, f.NotEquals(""))

	if len(rows) == 0 {
		return true, nil
	}

	firstRow := rows[0]
	if !strings.HasSuffix(firstRow, ":") {
		logrus.Debugf("No colon in first line of depfile")
		return false, nil
	}
	objFileInDepFile := firstRow[:len(firstRow)-1]
	if objFileInDepFile != objectFile.String() {
		logrus.Debugf("Depfile is about different object file: %v", objFileInDepFile)
		return false, nil
	}

	// The first line of the depfile contains the path to the object file to generate.
	// The second line of the depfile contains the path to the source file.
	// All subsequent lines contain the header files necessary to compile the object file.

	// If we don't do this check it might happen that trying to compile a source file
	// that has the same name but a different path wouldn't recreate the object file.
	if sourceFile.String() != strings.Trim(rows[1], " ") {
		logrus.Debugf("Depfile is about different source file: %v", strings.Trim(rows[1], " "))
		return false, nil
	}

	rows = rows[1:]
	for _, row := range rows {
		depStat, err := os.Stat(row)
		if err != nil && !os.IsNotExist(err) {
			// There is probably a parsing error of the dep file
			// Ignore the error and trigger a full rebuild anyway
			logrus.WithError(err).Debugf("Failed to read: %v", row)
			return false, nil
		}
		if os.IsNotExist(err) {
			logrus.Debugf("Not found: %v", row)
			return false, nil
		}
		if depStat.ModTime().After(objectFileStat.ModTime()) {
			logrus.Debugf("%v newer than %v", row, objectFile)
			return false, nil
		}
	}

	return true, nil
}

func removeEndingBackSlash(s string) string {
	return strings.TrimSuffix(s, "\\")
}

func unescapeDep(s string) string {
	s = strings.ReplaceAll(s, "\\ ", " ")
	s = strings.ReplaceAll(s, "\\\t", "\t")
	s = strings.ReplaceAll(s, "\\#", "#")
	s = strings.ReplaceAll(s, "$$", "$")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return s
}

// NormalizeUTF8 byte slice
// TODO: use it more often troughout all the project (maybe on logger interface?)
func NormalizeUTF8(buf []byte) []byte {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.Bytes(t, buf)
	return result
}

var sourceControlFolders = map[string]bool{"CVS": true, "RCS": true, ".git": true, ".github": true, ".svn": true, ".hg": true, ".bzr": true, ".vscode": true, ".settings": true, ".pioenvs": true, ".piolibdeps": true}

// filterOutSCCS is a ReadDirFilter that excludes known VSC or project files
func filterOutSCCS(file *paths.Path) bool {
	return !sourceControlFolders[file.Base()]
}

// filterOutHiddenFiles is a ReadDirFilter that exclude files with a "." prefix in their name
var filterOutHiddenFiles = paths.FilterOutPrefixes(".")

// FindFilesInFolder fixdoc
func FindFilesInFolder(dir *paths.Path, recurse bool, extensions ...string) (paths.PathList, error) {
	fileFilter := paths.AndFilter(
		filterOutHiddenFiles,
		filterOutSCCS,
		paths.FilterOutDirectories(),
	)
	if len(extensions) > 0 {
		fileFilter = paths.AndFilter(
			paths.FilterSuffixes(extensions...),
			fileFilter,
		)
	}
	if recurse {
		dirFilter := paths.AndFilter(
			filterOutHiddenFiles,
			filterOutSCCS,
		)
		return dir.ReadDirRecursiveFiltered(dirFilter, fileFilter)
	}
	return dir.ReadDir(fileFilter)
}

func printableArgument(arg string) string {
	if strings.ContainsAny(arg, "\"\\ \t") {
		arg = strings.ReplaceAll(arg, "\\", "\\\\")
		arg = strings.ReplaceAll(arg, "\"", "\\\"")
		return "\"" + arg + "\""
	}
	return arg
}

// PrintableCommand Convert a command and argument slice back to a printable string.
// This adds basic escaping which is sufficient for debug output, but
// probably not for shell interpretation. This essentially reverses
// ParseCommandLine.
func PrintableCommand(parts []string) string {
	return strings.Join(f.Map(parts, printableArgument), " ")
}

// DirContentIsOlderThan returns true if the content of the given directory is
// older than target file. If extensions are given, only the files with these
// extensions are tested.
func DirContentIsOlderThan(dir *paths.Path, target *paths.Path, extensions ...string) (bool, error) {
	targetStat, err := target.Stat()
	if err != nil {
		return false, err
	}
	targetModTime := targetStat.ModTime()

	files, err := FindFilesInFolder(dir, true, extensions...)
	if err != nil {
		return false, err
	}
	for _, file := range files {
		file, err := file.Stat()
		if err != nil {
			return false, err
		}
		if file.ModTime().After(targetModTime) {
			return false, nil
		}
	}
	return true, nil
}
