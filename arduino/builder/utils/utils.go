package utils

import (
	"os"
	"strings"
	"unicode"

	f "github.com/arduino/arduino-cli/internal/algorithms"
	"github.com/arduino/go-paths-helper"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func ObjFileIsUpToDate(sourceFile, objectFile, dependencyFile *paths.Path) (bool, error) {
	logrus.Debugf("Checking previous results for %v (result = %v, dep = %v)", sourceFile, objectFile, dependencyFile)
	if objectFile == nil || dependencyFile == nil {
		logrus.Debugf("Not found: nil")
		return false, nil
	}

	sourceFile = sourceFile.Clean()
	sourceFileStat, err := sourceFile.Stat()
	if err != nil {
		return false, errors.WithStack(err)
	}

	objectFile = objectFile.Clean()
	objectFileStat, err := objectFile.Stat()
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Debugf("Not found: %v", objectFile)
			return false, nil
		} else {
			return false, errors.WithStack(err)
		}
	}

	dependencyFile = dependencyFile.Clean()
	dependencyFileStat, err := dependencyFile.Stat()
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Debugf("Not found: %v", dependencyFile)
			return false, nil
		} else {
			return false, errors.WithStack(err)
		}
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
		return false, errors.WithStack(err)
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
		logrus.Debugf("Depfile is about different file: %v", objFileInDepFile)
		return false, nil
	}

	// The first line of the depfile contains the path to the object file to generate.
	// The second line of the depfile contains the path to the source file.
	// All subsequent lines contain the header files necessary to compile the object file.

	// If we don't do this check it might happen that trying to compile a source file
	// that has the same name but a different path wouldn't recreate the object file.
	if sourceFile.String() != strings.Trim(rows[1], " ") {
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

// Normalizes an UTF8 byte slice
// TODO: use it more often troughout all the project (maybe on logger interface?)
func NormalizeUTF8(buf []byte) []byte {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.Bytes(t, buf)
	return result
}

var SOURCE_CONTROL_FOLDERS = map[string]bool{"CVS": true, "RCS": true, ".git": true, ".github": true, ".svn": true, ".hg": true, ".bzr": true, ".vscode": true, ".settings": true, ".pioenvs": true, ".piolibdeps": true}

// FilterOutSCCS is a ReadDirFilter that excludes known VSC or project files
func FilterOutSCCS(file *paths.Path) bool {
	return !SOURCE_CONTROL_FOLDERS[file.Base()]
}

// FilterReadableFiles is a ReadDirFilter that accepts only readable files
func FilterReadableFiles(file *paths.Path) bool {
	// See if the file is readable by opening it
	f, err := file.Open()
	if err != nil {
		return false
	}
	f.Close()
	return true
}

// FilterOutHiddenFiles is a ReadDirFilter that exclude files with a "." prefix in their name
var FilterOutHiddenFiles = paths.FilterOutPrefixes(".")

func FindFilesInFolder(dir *paths.Path, recurse bool, extensions ...string) (paths.PathList, error) {
	fileFilter := paths.AndFilter(
		FilterOutHiddenFiles,
		FilterOutSCCS,
		paths.FilterOutDirectories(),
		FilterReadableFiles,
	)
	if len(extensions) > 0 {
		fileFilter = paths.AndFilter(
			paths.FilterSuffixes(extensions...),
			fileFilter,
		)
	}
	if recurse {
		dirFilter := paths.AndFilter(
			FilterOutHiddenFiles,
			FilterOutSCCS,
		)
		return dir.ReadDirRecursiveFiltered(dirFilter, fileFilter)
	}
	return dir.ReadDir(fileFilter)
}

func WrapWithHyphenI(value string) string {
	return "\"-I" + value + "\""
}
