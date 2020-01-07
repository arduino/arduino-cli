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
	"bufio"
	"fmt"
	"os"
	"os/exec"
)

// OutputListener is a callback interface to receive output messages from process
type OutputListener interface {
	Output(msg string)
}

// AttachStdoutListener adds an OutputListener to the stdout of the process
func AttachStdoutListener(cmd *exec.Cmd, listener OutputListener) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("can't retrieve standard output stream: %s", err)
	}

	stdoutCopy := bufio.NewScanner(stdout)
	stdoutCopy.Split(bufio.ScanLines)
	go func() {
		for stdoutCopy.Scan() {
			listener.Output(stdoutCopy.Text())
		}
	}()

	return nil
}

// AttachStderrListener adds an OutputListener to the stderr of the process
func AttachStderrListener(cmd *exec.Cmd, listener OutputListener) error {
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("can't retrieve standard error stream: %s", err)
	}

	stderrCopy := bufio.NewScanner(stderr)
	stderrCopy.Split(bufio.ScanLines)
	go func() {
		for stderrCopy.Scan() {
			listener.Output(stderrCopy.Text())
		}
	}()

	return nil
}

// PrintToStdout is an OutputListener that outputs messages to standard output
var PrintToStdout = &printToStdout{}

type printToStdout struct{}

func (*printToStdout) Output(msg string) {
	fmt.Fprintln(os.Stdout, msg)
}

// PrintToStderr is an OutputListener that outputs messages to standard error
var PrintToStderr = &printToStderr{}

type printToStderr struct{}

func (*printToStderr) Output(msg string) {
	fmt.Fprintln(os.Stderr, msg)
}
