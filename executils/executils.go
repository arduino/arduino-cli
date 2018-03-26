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
	"bytes"
	"fmt"
	"io"
	"os/exec"
)

// PipeCommands executes the commands received as input by feeding the output of
// one to the input of the other, exactly like Unix Pipe (|).
// Returns the output of the final command and the eventual error.
//
// code inspired by https://gist.github.com/tyndyll/89fbb2c2273f83a074dc
func PipeCommands(commands ...*exec.Cmd) ([]byte, error) {
	var errorBuffer, outputBuffer bytes.Buffer
	pipeStack := make([]*io.PipeWriter, len(commands)-1)
	i := 0
	for ; i < len(commands)-1; i++ {
		stdinPipe, stdoutPipe := io.Pipe()
		commands[i].Stdout = stdoutPipe
		commands[i].Stderr = &errorBuffer
		commands[i+1].Stdin = stdinPipe
		pipeStack[i] = stdoutPipe
	}
	commands[i].Stdout = &outputBuffer
	commands[i].Stderr = &errorBuffer

	if err := call(commands, pipeStack); err != nil {
		return nil, err
	}

	return outputBuffer.Bytes(), nil
}

func call(stack []*exec.Cmd, pipes []*io.PipeWriter) (err error) {
	if stack[0].Process == nil {
		if err = stack[0].Start(); err != nil {
			return err
		}
	}
	if len(stack) > 1 {
		if err = stack[1].Start(); err != nil {
			return err
		}
		defer func() {
			pipes[0].Close()
			err = call(stack[1:], pipes[1:])
		}()
	}
	return stack[0].Wait()
}

// TellCommandNotToSpawnShell avoids that the specified Cmd display a small
// command prompt while runnning on Windows. It has no effects on other OS.
func TellCommandNotToSpawnShell(cmd *exec.Cmd) {
	tellCommandNotToSpawnShell(cmd)
}

// Command creates a command with the provided command line arguments.
// The first argument is the path to the executable, the remainder are the
// arguments to the command.
func Command(args []string) (*exec.Cmd, error) {
	if args == nil || len(args) == 0 {
		return nil, fmt.Errorf("no executable specified")
	}
	cmd := exec.Command(args[0], args[1:]...)
	TellCommandNotToSpawnShell(cmd)
	return cmd, nil
}
