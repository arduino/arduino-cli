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

package executils

import (
	"io"
	"os"
	"os/exec"

	"github.com/arduino/go-paths-helper"
	"github.com/pkg/errors"
)

// Process is representation of an external process run
type Process struct {
	cmd *exec.Cmd
}

// NewProcess creates a command with the provided command line arguments.
// The first argument is the path to the executable, the remainder are the
// arguments to the command.
func NewProcess(args ...string) (*Process, error) {
	if args == nil || len(args) == 0 {
		return nil, errors.New("no executable specified")
	}
	p := &Process{
		cmd: exec.Command(args[0], args[1:]...),
	}
	TellCommandNotToSpawnShell(p.cmd)

	// This is required because some tools detects if the program is running
	// from terminal by looking at the stdin/out bindings.
	// https://github.com/arduino/arduino-cli/issues/844
	p.cmd.Stdin = NullReader
	return p, nil
}

// NewProcessFromPath creates a command from the provided executable path and
// command line arguments.
func NewProcessFromPath(executable *paths.Path, args ...string) (*Process, error) {
	processArgs := []string{executable.String()}
	processArgs = append(processArgs, args...)
	return NewProcess(processArgs...)
}

// RedirectStdoutTo will redirect the process' stdout to the specified
// writer. Any previous redirection will be overwritten.
func (p *Process) RedirectStdoutTo(out io.Writer) {
	p.cmd.Stdout = out
}

// RedirectStderrTo will redirect the process' stdout to the specified
// writer. Any previous redirection will be overwritten.
func (p *Process) RedirectStderrTo(out io.Writer) {
	p.cmd.Stderr = out
}

// StdinPipe returns a pipe that will be connected to the command's standard
// input when the command starts. The pipe will be closed automatically after
// Wait sees the command exit. A caller need only call Close to force the pipe
// to close sooner. For example, if the command being run will not exit until
// standard input is closed, the caller must close the pipe.
func (p *Process) StdinPipe() (io.WriteCloser, error) {
	if p.cmd.Stdin == NullReader {
		p.cmd.Stdin = nil
	}
	return p.cmd.StdinPipe()
}

// StdoutPipe returns a pipe that will be connected to the command's standard
// output when the command starts.
func (p *Process) StdoutPipe() (io.ReadCloser, error) {
	return p.cmd.StdoutPipe()
}

// StderrPipe returns a pipe that will be connected to the command's standard
// error when the command starts.
func (p *Process) StderrPipe() (io.ReadCloser, error) {
	return p.cmd.StderrPipe()
}

// Start will start the underliyng process.
func (p *Process) Start() error {
	return p.cmd.Start()
}

// Wait waits for the command to exit and waits for any copying to stdin or copying
// from stdout or stderr to complete.
func (p *Process) Wait() error {
	// TODO: make some helpers to retrieve exit codes out of *ExitError.
	return p.cmd.Wait()
}

// Signal sends a signal to the Process. Sending Interrupt on Windows is not implemented.
func (p *Process) Signal(sig os.Signal) error {
	return p.cmd.Process.Signal(sig)
}

// Kill causes the Process to exit immediately. Kill does not wait until the Process has
// actually exited. This only kills the Process itself, not any other processes it may
// have started.
func (p *Process) Kill() error {
	return p.cmd.Process.Kill()
}

// SetDir sets the working directory of the command. If Dir is the empty string, Run
// runs the command in the calling process's current directory.
func (p *Process) SetDir(dir string) {
	p.cmd.Dir = dir
}

// SetDirFromPath sets the working directory of the command. If path is nil, Run
// runs the command in the calling process's current directory.
func (p *Process) SetDirFromPath(path *paths.Path) {
	if path == nil {
		p.cmd.Dir = ""
	} else {
		p.cmd.Dir = path.String()
	}
}

// Run starts the specified command and waits for it to complete.
func (p *Process) Run() error {
	return p.cmd.Run()
}
