package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
)

type BuilderLogger struct {
	stdLock sync.Mutex
	stdout  io.Writer
	stderr  io.Writer

	verbose       bool
	warningsLevel string
}

func New(stdout, stderr io.Writer, verbose bool, warningsLevel string) *BuilderLogger {
	return &BuilderLogger{
		stdout:        stdout,
		stderr:        stderr,
		verbose:       verbose,
		warningsLevel: warningsLevel,
	}
}

func (l *BuilderLogger) Info(msg string) {
	l.stdLock.Lock()
	if l.stdout == nil {
		fmt.Fprintln(os.Stdout, msg)
	} else {
		fmt.Fprintln(l.stdout, msg)
	}
	l.stdLock.Unlock()
}

func (l *BuilderLogger) Warn(msg string) {
	l.stdLock.Lock()
	if l.stderr == nil {
		fmt.Fprintln(os.Stderr, msg)
	} else {
		fmt.Fprintln(l.stderr, msg)
	}
	l.stdLock.Unlock()
}

func (l *BuilderLogger) WriteStdout(data []byte) (int, error) {
	l.stdLock.Lock()
	defer l.stdLock.Unlock()
	if l.stdout == nil {
		return os.Stdout.Write(data)
	}
	return l.stdout.Write(data)
}

func (l *BuilderLogger) WriteStderr(data []byte) (int, error) {
	l.stdLock.Lock()
	defer l.stdLock.Unlock()
	if l.stderr == nil {
		return os.Stderr.Write(data)
	}
	return l.stderr.Write(data)
}
