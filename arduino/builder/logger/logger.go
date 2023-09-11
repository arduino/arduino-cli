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

package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// BuilderLogger fixdoc
type BuilderLogger struct {
	stdLock sync.Mutex
	stdout  io.Writer
	stderr  io.Writer

	verbose       bool
	warningsLevel string
}

// New fixdoc
func New(stdout, stderr io.Writer, verbose bool, warningsLevel string) *BuilderLogger {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	return &BuilderLogger{
		stdout:        stdout,
		stderr:        stderr,
		verbose:       verbose,
		warningsLevel: warningsLevel,
	}
}

// Info fixdoc
func (l *BuilderLogger) Info(msg string) {
	l.stdLock.Lock()
	defer l.stdLock.Unlock()
	fmt.Fprintln(l.stdout, msg)
}

// Warn fixdoc
func (l *BuilderLogger) Warn(msg string) {
	l.stdLock.Lock()
	defer l.stdLock.Unlock()
	fmt.Fprintln(l.stderr, msg)
}

// WriteStdout fixdoc
func (l *BuilderLogger) WriteStdout(data []byte) (int, error) {
	l.stdLock.Lock()
	defer l.stdLock.Unlock()
	return l.stdout.Write(data)
}

// WriteStderr fixdoc
func (l *BuilderLogger) WriteStderr(data []byte) (int, error) {
	l.stdLock.Lock()
	defer l.stdLock.Unlock()
	return l.stderr.Write(data)
}

// Verbose fixdoc
func (l *BuilderLogger) Verbose() bool {
	return l.verbose
}

// WarningsLevel fixdoc
func (l *BuilderLogger) WarningsLevel() string {
	return l.warningsLevel
}

// Stdout fixdoc
func (l *BuilderLogger) Stdout() io.Writer {
	return l.stdout
}

// Stderr fixdoc
func (l *BuilderLogger) Stderr() io.Writer {
	return l.stderr
}
