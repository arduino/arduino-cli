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
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/arduino/arduino-cli/i18n"
	f "github.com/arduino/arduino-cli/internal/algorithms"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	paths "github.com/arduino/go-paths-helper"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type filterFiles func([]os.FileInfo) []os.FileInfo

var tr = i18n.Tr

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

func WrapWithHyphenI(value string) string {
	return "\"-I" + value + "\""
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
	return strings.Join(f.Map(parts, printableArgument), " ")
}

const (
	Ignore        = 0 // Redirect to null
	Show          = 1 // Show on stdout/stderr as normal
	ShowIfVerbose = 2 // Show if verbose is set, Ignore otherwise
	Capture       = 3 // Capture into buffer
)

func ExecCommand(ctx *types.Context, command *exec.Cmd, stdout int, stderr int) ([]byte, []byte, error) {
	if ctx.Verbose {
		ctx.Info(PrintableCommand(command.Args))
	}

	if stdout == Capture {
		buffer := &bytes.Buffer{}
		command.Stdout = buffer
	} else if stdout == Show || (stdout == ShowIfVerbose && ctx.Verbose) {
		if ctx.Stdout != nil {
			command.Stdout = ctx.Stdout
		} else {
			command.Stdout = os.Stdout
		}
	}

	if stderr == Capture {
		buffer := &bytes.Buffer{}
		command.Stderr = buffer
	} else if stderr == Show || (stderr == ShowIfVerbose && ctx.Verbose) {
		if ctx.Stderr != nil {
			command.Stderr = ctx.Stderr
		} else {
			command.Stderr = os.Stderr
		}
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
		if !slices.Contains(target, element) {
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

	for _, dirEntry := range entries {
		entry, scopeErr := dirEntry.Info()
		if scopeErr != nil {
			return
		}

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
