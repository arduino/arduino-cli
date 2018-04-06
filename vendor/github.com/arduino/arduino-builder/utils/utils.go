/*
 * This file is part of Arduino Builder.
 *
 * Arduino Builder is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2015 Arduino LLC (http://www.arduino.cc/)
 */

package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"unicode/utf8"

	"github.com/arduino/arduino-builder/constants"
	"github.com/arduino/arduino-builder/gohasissues"
	"github.com/arduino/arduino-builder/i18n"
	"github.com/arduino/arduino-builder/types"
)

func KeysOfMapOfString(input map[string]string) []string {
	var keys []string
	for key, _ := range input {
		keys = append(keys, key)
	}
	return keys
}

func PrettyOSName() string {
	switch osName := runtime.GOOS; osName {
	case "darwin":
		return "macosx"
	case "freebsd":
		return "freebsd"
	case "linux":
		return "linux"
	case "windows":
		return "windows"
	default:
		return "other"
	}
}

func ParseCommandLine(input string, logger i18n.Logger) ([]string, error) {
	var parts []string
	escapingChar := constants.EMPTY_STRING
	escapedArg := constants.EMPTY_STRING
	for _, inputPart := range strings.Split(input, constants.SPACE) {
		inputPart = strings.TrimSpace(inputPart)
		if len(inputPart) == 0 {
			continue
		}

		if escapingChar == constants.EMPTY_STRING {
			if inputPart[0] != '"' && inputPart[0] != '\'' {
				parts = append(parts, inputPart)
				continue
			}

			escapingChar = string(inputPart[0])
			inputPart = inputPart[1:]
			escapedArg = constants.EMPTY_STRING
		}

		if inputPart[len(inputPart)-1] != '"' && inputPart[len(inputPart)-1] != '\'' {
			escapedArg = escapedArg + inputPart + " "
			continue
		}

		escapedArg = escapedArg + inputPart[:len(inputPart)-1]
		escapedArg = strings.TrimSpace(escapedArg)
		if len(escapedArg) > 0 {
			parts = append(parts, escapedArg)
		}
		escapingChar = constants.EMPTY_STRING
	}

	if escapingChar != constants.EMPTY_STRING {
		return nil, i18n.ErrorfWithLogger(logger, constants.MSG_INVALID_QUOTING, escapingChar)
	}

	return parts, nil
}

type filterFiles func([]os.FileInfo) []os.FileInfo

func ReadDirFiltered(folder string, fn filterFiles) ([]os.FileInfo, error) {
	files, err := gohasissues.ReadDir(folder)
	if err != nil {
		return nil, i18n.WrapError(err)
	}
	return fn(files), nil
}

func FilterDirs(files []os.FileInfo) []os.FileInfo {
	var filtered []os.FileInfo
	for _, info := range files {
		if info.IsDir() {
			filtered = append(filtered, info)
		}
	}
	return filtered
}

func FilterFilesWithExtensions(extensions ...string) filterFiles {
	return func(files []os.FileInfo) []os.FileInfo {
		var filtered []os.FileInfo
		for _, file := range files {
			if !file.IsDir() && SliceContains(extensions, filepath.Ext(file.Name())) {
				filtered = append(filtered, file)
			}
		}
		return filtered
	}
}

func FilterFiles() filterFiles {
	return func(files []os.FileInfo) []os.FileInfo {
		var filtered []os.FileInfo
		for _, file := range files {
			if !file.IsDir() {
				filtered = append(filtered, file)
			}
		}
		return filtered
	}
}

var SOURCE_CONTROL_FOLDERS = map[string]bool{"CVS": true, "RCS": true, ".git": true, ".github": true, ".svn": true, ".hg": true, ".bzr": true, ".vscode": true}

func IsSCCSOrHiddenFile(file os.FileInfo) bool {
	return IsSCCSFile(file) || IsHiddenFile(file)
}

func IsHiddenFile(file os.FileInfo) bool {
	name := filepath.Base(file.Name())

	if name[0] == '.' {
		return true
	}

	return false
}

func IsSCCSFile(file os.FileInfo) bool {
	name := filepath.Base(file.Name())

	if SOURCE_CONTROL_FOLDERS[name] {
		return true
	}

	return false
}

func SliceContains(slice []string, target string) bool {
	for _, value := range slice {
		if value == target {
			return true
		}
	}
	return false
}

type mapFunc func(string) string

func Map(slice []string, fn mapFunc) []string {
	newSlice := []string{}
	for _, elem := range slice {
		newSlice = append(newSlice, fn(elem))
	}
	return newSlice
}

type filterFunc func(string) bool

func Filter(slice []string, fn filterFunc) []string {
	newSlice := []string{}
	for _, elem := range slice {
		if fn(elem) {
			newSlice = append(newSlice, elem)
		}
	}
	return newSlice
}

func WrapWithHyphenI(value string) string {
	return "\"-I" + value + "\""
}

func TrimSpace(value string) string {
	return strings.TrimSpace(value)
}

type argFilterFunc func(int, string, []string) bool

func PrepareCommandFilteredArgs(pattern string, filter argFilterFunc, logger i18n.Logger) (*exec.Cmd, error) {
	parts, err := ParseCommandLine(pattern, logger)
	if err != nil {
		return nil, i18n.WrapError(err)
	}
	command := parts[0]
	parts = parts[1:]
	var args []string
	for idx, part := range parts {
		if filter(idx, part, parts) {
			args = append(args, part)
		}
	}

	return exec.Command(command, args...), nil
}

func filterEmptyArg(_ int, arg string, _ []string) bool {
	return arg != constants.EMPTY_STRING
}

func PrepareCommand(pattern string, logger i18n.Logger) (*exec.Cmd, error) {
	return PrepareCommandFilteredArgs(pattern, filterEmptyArg, logger)
}

func MapStringStringHas(aMap map[string]string, key string) bool {
	_, ok := aMap[key]
	return ok
}

func AbsolutizePaths(files []string) ([]string, error) {
	for idx, file := range files {
		absFile, err := filepath.Abs(file)
		if err != nil {
			return nil, i18n.WrapError(err)
		}
		files[idx] = absFile
	}

	return files, nil
}

func ReadFileToRows(file string) ([]string, error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, i18n.WrapError(err)
	}
	txt := string(bytes)
	txt = strings.Replace(txt, "\r\n", "\n", -1)

	return strings.Split(txt, "\n"), nil
}

type CheckExtensionFunc func(ext string) bool

func FindFilesInFolder(files *[]string, folder string, extensions CheckExtensionFunc, recurse bool) error {
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip source control and hidden files and directories
		if IsSCCSOrHiddenFile(info) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories unless recurse is on, or this is the
		// root directory
		if info.IsDir() {
			if recurse || path == folder {
				return nil
			} else {
				return filepath.SkipDir
			}
		}

		// Check (lowercased) extension against list of extensions
		if extensions != nil && !extensions(strings.ToLower(filepath.Ext(path))) {
			return nil
		}

		// See if the file is readable by opening it
		currentFile, err := os.Open(path)
		if err != nil {
			return nil
		}
		currentFile.Close()

		*files = append(*files, path)
		return nil
	}
	return gohasissues.Walk(folder, walkFunc)
}

func GetParentFolder(basefolder string, n int) string {
	tempFolder := basefolder
	i := 0
	for i < n {
		tempFolder = filepath.Dir(tempFolder)
		i++
	}
	return tempFolder
}

func AppendIfNotPresent(target []string, elements ...string) []string {
	for _, element := range elements {
		if !SliceContains(target, element) {
			target = append(target, element)
		}
	}
	return target
}

func EnsureFolderExists(folder string) error {
	return os.MkdirAll(folder, os.FileMode(0755))
}

func WriteFileBytes(targetFilePath string, data []byte) error {
	return ioutil.WriteFile(targetFilePath, data, os.FileMode(0644))
}

func WriteFile(targetFilePath string, data string) error {
	return WriteFileBytes(targetFilePath, []byte(data))
}

func TouchFile(targetFilePath string) error {
	return WriteFileBytes(targetFilePath, []byte{})
}

func NULLFile() string {
	if runtime.GOOS == "windows" {
		return "nul"
	}
	return "/dev/null"
}

func MD5Sum(data []byte) string {
	md5sumBytes := md5.Sum(data)
	return hex.EncodeToString(md5sumBytes[:])
}

type loggerAction struct {
	onlyIfVerbose bool
	level         string
	format        string
	args          []interface{}
}

func (l *loggerAction) Run(ctx *types.Context) error {
	if !l.onlyIfVerbose || ctx.Verbose {
		ctx.GetLogger().Println(l.level, l.format, l.args...)
	}
	return nil
}

func LogIfVerbose(level string, format string, args ...interface{}) types.Command {
	return &loggerAction{true, level, format, args}
}

// Returns the given string as a quoted string for use with the C
// preprocessor. This adds double quotes around it and escapes any
// double quotes and backslashes in the string.
func QuoteCppString(str string) string {
	str = strings.Replace(str, "\\", "\\\\", -1)
	str = strings.Replace(str, "\"", "\\\"", -1)
	return "\"" + str + "\""
}

// Parse a C-preprocessor string as emitted by the preprocessor. This
// is a string contained in double quotes, with any backslashes or
// quotes escaped with a backslash. If a valid string was present at the
// start of the given line, returns the unquoted string contents, the
// remaineder of the line (everything after the closing "), and true.
// Otherwise, returns the empty string, the entire line and false.
func ParseCppString(line string) (string, string, bool) {
	// For details about how these strings are output by gcc, see:
	// https://github.com/gcc-mirror/gcc/blob/a588355ab948cf551bc9d2b89f18e5ae5140f52c/libcpp/macro.c#L491-L511
	// Note that the documentaiton suggests all non-printable
	// characters are also escaped, but the implementation does not
	// actually do this. See https://gcc.gnu.org/bugzilla/show_bug.cgi?id=51259
	if len(line) < 1 || line[0] != '"' {
		return "", line, false
	}

	i := 1
	res := ""
	for {
		if i >= len(line) {
			return "", line, false
		}

		c, width := utf8.DecodeRuneInString(line[i:])

		switch c {
		// Backslash, next character is used unmodified
		case '\\':
			i += width
			if i >= len(line) {
				return "", line, false
			}
			res += string(line[i])
			break
		// Quote, end of string
		case '"':
			return res, line[i+width:], true
		default:
			res += string(c)
			break
		}

		i += width
	}
}
