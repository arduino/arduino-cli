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
