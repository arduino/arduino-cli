// This file is part of arduino-cli.
//
// Copyright 2022 ARDUINO SA (http://www.arduino.cc/)
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

package feedback

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/mattn/go-isatty"
	"golang.org/x/term"
)

// InteractiveStreams returns the underlying io.Reader and io.Writer to directly stream to
// stdin and stdout. It errors if the selected output format is not Text or the terminal is
// not interactive.
func InteractiveStreams() (io.Reader, io.Writer, error) {
	if !formatSelected {
		panic("output format not yet selected")
	}
	if format != Text {
		return nil, nil, errors.New(i18n.Tr("interactive terminal not supported for the '%s' output format", format))
	}
	return os.Stdin, stdOut, nil
}

var oldStateStdin *term.State

// SetRawModeStdin sets the stdin stream in RAW mode (no buffering, echo disabled,
// no terminal escape codes nor signals interpreted)
func SetRawModeStdin() error {
	if oldStateStdin != nil {
		panic("terminal already in RAW mode")
	}
	if !IsInteractive() {
		return errors.New(i18n.Tr("not running in a terminal"))
	}
	old, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	oldStateStdin = old
	return nil
}

// RestoreModeStdin restore the terminal settings to the normal non-RAW state. This
// function must be called after SetRawModeStdin to not leave the terminal in an
// undefined state.
func RestoreModeStdin() {
	if oldStateStdin == nil {
		return
	}
	_ = term.Restore(int(os.Stdin.Fd()), oldStateStdin)
	oldStateStdin = nil
}

// IsInteractive returns true if the CLI is interactive (it can receive inputs from terminal/console)
func IsInteractive() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
}

// HasConsole returns true if the CLI outputs to a terminal/console
func HasConsole() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

// IsCI returns true if running on CI environments.
// (based on https://github.com/watson/ci-info/blob/HEAD/index.js)
func IsCI() bool {
	return os.Getenv("CI") != "" || // GitHub Actions, Travis CI, CircleCI, Cirrus CI, GitLab CI, AppVeyor, CodeShip, dsari
		os.Getenv("BUILD_NUMBER") != "" || // Jenkins, TeamCity
		os.Getenv("RUN_ID") != "" // TaskCluster, dsari
}

// InputUserField prompts the user to input the provided user field.
func InputUserField(prompt string, secret bool) (string, error) {
	if format != Text {
		return "", errors.New(i18n.Tr("user input not supported for the '%s' output format", format))
	}
	if !IsInteractive() {
		return "", errors.New(i18n.Tr("user input not supported in non interactive mode"))
	}

	fmt.Fprintf(stdOut, "%s: ", prompt)

	if secret {
		// Read and return a password (no characted echoed on terminal)
		value, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(stdOut)
		return string(value), err
	}

	// Read and return an input line
	sc := bufio.NewScanner(os.Stdin)
	sc.Scan()
	return sc.Text(), sc.Err()
}

// ExitWhenParentProcessEnds waits until the controlling parent process ends and then exits
// the current process. This is useful to terminate the current process when it is daemonized
// and the controlling parent process is terminated to avoid leaving zombie processes.
// It is recommended to call this function as a goroutine.
func ExitWhenParentProcessEnds() {
	// Stdin is closed when the controlling parent process ends
	_, _ = io.Copy(io.Discard, os.Stdin)
	os.Exit(0)
}
