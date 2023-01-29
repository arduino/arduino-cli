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

package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/arduino/arduino-cli/i18n"
	"github.com/arduino/arduino-cli/legacy/builder/gohasissues"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	paths "github.com/arduino/go-paths-helper"
	"github.com/pkg/errors"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type filterFiles func([]os.FileInfo) []os.FileInfo

var tr = i18n.Tr

func ReadDirFiltered(folder string, fn filterFiles) ([]os.FileInfo, error) {
	files, err := gohasissues.ReadDir(folder)
	if err != nil {
		return nil, errors.WithStack(err)
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

var SOURCE_CONTROL_FOLDERS = map[string]bool{"CVS": true, "RCS": true, ".git": true, ".github": true, ".svn": true, ".hg": true, ".bzr": true, ".vscode": true, ".settings": true, ".pioenvs": true, ".piolibdeps": true}

// FilterOutHiddenFiles is a ReadDirFilter that exclude files with a "." prefix in their name
var FilterOutHiddenFiles = paths.FilterOutPrefixes(".")

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

func IsSCCSOrHiddenFile(file os.FileInfo) bool {
	return IsSCCSFile(file) || IsHiddenFile(file)
}

func IsHiddenFile(file os.FileInfo) bool {
	name := filepath.Base(file.Name())
	return name[0] == '.'
}

func IsSCCSFile(file os.FileInfo) bool {
	name := filepath.Base(file.Name())
	return SOURCE_CONTROL_FOLDERS[name]
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

func printableArgument(arg string) string {
	if strings.ContainsAny(arg, "\"\\ \t") {
		arg = strings.Replace(arg, "\\", "\\\\", -1)
		arg = strings.Replace(arg, "\"", "\\\"", -1)
		return "\"" + arg + "\""
	} else {
		return arg
	}
}

// Convert a command and argument slice back to a printable string.
// This adds basic escaping which is sufficient for debug output, but
// probably not for shell interpretation. This essentially reverses
// ParseCommandLine.
func PrintableCommand(parts []string) string {
	return strings.Join(Map(parts, printableArgument), " ")
}

const (
	Ignore        = 0 // Redirect to null
	Show          = 1 // Show on stdout/stderr as normal
	ShowIfVerbose = 2 // Show if verbose is set, Ignore otherwise
	Capture       = 3 // Capture into buffer
)

func ExecCommand(ctx *types.Context, command *exec.Cmd, stdout int, stderr int) ([]byte, []byte, error) {
	if ctx.Stdout == nil {
		ctx.Stdout = os.Stdout
	}
	if ctx.Stderr == nil {
		ctx.Stderr = os.Stderr
	}

	if ctx.Verbose {
		ctx.Info(PrintableCommand(command.Args))
	}

	if stdout == Capture {
		buffer := &bytes.Buffer{}
		command.Stdout = buffer
	} else if stdout == Show || stdout == ShowIfVerbose && ctx.Verbose {
		command.Stdout = ctx.Stdout
	}

	if stderr == Capture {
		buffer := &bytes.Buffer{}
		command.Stderr = buffer
	} else if stderr == Show || stderr == ShowIfVerbose && ctx.Verbose {
		command.Stderr = ctx.Stderr
	}

	err := command.Start()
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	err = command.Wait()

	var outbytes, errbytes []byte
	if buf, ok := command.Stdout.(*bytes.Buffer); ok {
		outbytes = buf.Bytes()
	}
	if buf, ok := command.Stderr.(*bytes.Buffer); ok {
		errbytes = buf.Bytes()
	}

	return outbytes, errbytes, errors.WithStack(err)
}

func AbsolutizePaths(files []string) ([]string, error) {
	for idx, file := range files {
		if file == "" {
			continue
		}
		absFile, err := filepath.Abs(file)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		files[idx] = absFile
	}

	return files, nil
}

func FindFilesInFolder(dir *paths.Path, recurse bool, extensions []string) (paths.PathList, error) {
	fileFilter := paths.AndFilter(
		paths.FilterSuffixes(extensions...),
		FilterOutHiddenFiles,
		FilterOutSCCS,
		paths.FilterOutDirectories(),
		FilterReadableFiles,
	)
	if recurse {
		dirFilter := paths.AndFilter(
			FilterOutHiddenFiles,
			FilterOutSCCS,
		)
		return dir.ReadDirRecursiveFiltered(dirFilter, fileFilter)
	}
	return dir.ReadDir(fileFilter)
}

func AppendIfNotPresent(target []string, elements ...string) []string {
	for _, element := range elements {
		if !SliceContains(target, element) {
			target = append(target, element)
		}
	}
	return target
}

func MD5Sum(data []byte) string {
	md5sumBytes := md5.Sum(data)
	return hex.EncodeToString(md5sumBytes[:])
}

type loggerAction struct {
	onlyIfVerbose bool
	warn          bool
	msg           string
}

func (l *loggerAction) Run(ctx *types.Context) error {
	if !l.onlyIfVerbose || ctx.Verbose {
		if l.warn {
			ctx.Warn(l.msg)
		} else {
			ctx.Info(l.msg)
		}
	}
	return nil
}

func LogIfVerbose(warn bool, msg string) types.Command {
	return &loggerAction{onlyIfVerbose: true, warn: warn, msg: msg}
}

// Returns the given string as a quoted string for use with the C
// preprocessor. This adds double quotes around it and escapes any
// double quotes and backslashes in the string.
func QuoteCppString(str string) string {
	str = strings.Replace(str, "\\", "\\\\", -1)
	str = strings.Replace(str, "\"", "\\\"", -1)
	return "\"" + str + "\""
}

func QuoteCppPath(path *paths.Path) string {
	return QuoteCppString(path.String())
}

// Parse a C-preprocessor string as emitted by the preprocessor. This
// is a string contained in double quotes, with any backslashes or
// quotes escaped with a backslash. If a valid string was present at the
// start of the given line, returns the unquoted string contents, the
// remainder of the line (everything after the closing "), and true.
// Otherwise, returns the empty string, the entire line and false.
func ParseCppString(line string) (string, string, bool) {
	// For details about how these strings are output by gcc, see:
	// https://github.com/gcc-mirror/gcc/blob/a588355ab948cf551bc9d2b89f18e5ae5140f52c/libcpp/macro.c#L491-L511
	// Note that the documentation suggests all non-printable
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
		case '\\':
			// Backslash, next character is used unmodified
			i += width
			if i >= len(line) {
				return "", line, false
			}
			res += string(line[i])
		case '"':
			// Quote, end of string
			return res, line[i+width:], true
		default:
			res += string(c)
		}

		i += width
	}
}

// Normalizes an UTF8 byte slice
// TODO: use it more often troughout all the project (maybe on logger interface?)
func NormalizeUTF8(buf []byte) []byte {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.Bytes(t, buf)
	return result
}

// CopyFile copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file. The file mode will be copied from the source and
// the copied data is synced/flushed to stable storage.
func CopyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
// Symlinks are ignored and skipped.
func CopyDir(src string, dst string, extensions []string) (err error) {
	isAcceptedExtension := func(ext string) bool {
		ext = strings.ToLower(ext)
		for _, valid := range extensions {
			if ext == valid {
				return true
			}
		}
		return false
	}

	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf(tr("source is not a directory"))
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	if err == nil {
		return fmt.Errorf(tr("destination already exists"))
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return
	}

	infos := make([]fs.FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, scopeErr := entry.Info()
		if scopeErr != nil {
			return
		}
		infos = append(infos, info)
	}

	for _, entry := range infos {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath, extensions)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			if !isAcceptedExtension(filepath.Ext(srcPath)) {
				continue
			}

			err = CopyFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}

	return
}
