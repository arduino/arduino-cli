/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017-2018 ARDUINO AG (http://www.arduino.cc/)
 */

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
