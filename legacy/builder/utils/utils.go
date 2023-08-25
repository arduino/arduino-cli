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
	"os"
	"os/exec"
	"strings"
	"unicode"

	f "github.com/arduino/arduino-cli/internal/algorithms"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	paths "github.com/arduino/go-paths-helper"
	"github.com/pkg/errors"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

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
