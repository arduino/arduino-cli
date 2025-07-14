// This file is part of arduino-cli.
//
// Copyright 2024 ARDUINO SA (http://www.arduino.cc/)
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

package runner

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/arduino/go-paths-helper"
)

// Task is a command to be executed
type Task struct {
	Args        []string `json:"args"`
	LimitStderr int      `json:"-"`
}

// NewTask creates a new Task
func NewTask(args ...string) *Task {
	return &Task{Args: args}
}

// NewTaskWithLimitedStderr creates a new Task with a hard-limit on the stderr output
func NewTaskWithLimitedStderr(limit int, args ...string) *Task {
	return &Task{Args: args, LimitStderr: limit}
}

func (t *Task) String() string {
	return strings.Join(t.Args, " ")
}

// Result contains the output of a command execution
type Result struct {
	Args   []string
	Stdout []byte
	Stderr []byte
	Error  error
}

// Run executes the command and returns the result
func (t *Task) Run(ctx context.Context) *Result {
	proc, err := paths.NewProcess(nil, t.Args...)
	if err != nil {
		return &Result{Args: t.Args, Error: err}
	}

	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	proc.RedirectStdoutTo(stdout)

	if t.LimitStderr > 0 {
		innerCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		count := 0
		stderrLimited := writerFunc(func(p []byte) (int, error) {
			n, err := stderr.Write(p)
			count += n
			if count > t.LimitStderr {
				fmt.Fprintln(stderr, i18n.Tr("Compiler error output has been truncated."))
				cancel()
			}
			return n, err
		})

		ctx = innerCtx
		proc.RedirectStderrTo(stderrLimited)
	} else {
		proc.RedirectStderrTo(stderr)
	}

	// Append arguments to stdout
	fmt.Fprintln(stdout, t.String())

	// Execute command and wait for the process to finish
	if err := proc.Start(); err != nil {
		return &Result{Error: err}
	}
	err = proc.WaitWithinContext(ctx)
	return &Result{Args: proc.GetArgs(), Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), Error: err}
}

type writerFunc func(p []byte) (n int, err error)

func (f writerFunc) Write(p []byte) (n int, err error) {
	return f(p)
}
