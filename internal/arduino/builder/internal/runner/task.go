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
	"context"
	"fmt"
	"strings"

	"github.com/arduino/go-paths-helper"
)

// Task is a command to be executed
type Task struct {
	Args []string `json:"args"`
}

// NewTask creates a new Task
func NewTask(args ...string) *Task {
	return &Task{Args: args}
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
	stdout, stderr, err := proc.RunAndCaptureOutput(ctx)

	// Append arguments to stdout
	stdout = append([]byte(fmt.Sprintln(t)), stdout...)

	return &Result{Args: proc.GetArgs(), Stdout: stdout, Stderr: stderr, Error: err}
}
