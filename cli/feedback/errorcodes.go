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

// ExitCode to be used for Fatal.
type ExitCode int

const (
	// Success (0 is the no-error return code in Unix)
	Success ExitCode = iota

	// ErrGeneric Generic error (1 is the reserved "catchall" code in Unix)
	ErrGeneric

	_ // (2 Is reserved in Unix)

	// ErrNoConfigFile is returned when the config file is not found (3)
	ErrNoConfigFile

	_ // (4 was ErrBadCall and has been removed)

	// ErrNetwork is returned when a network error occurs (5)
	ErrNetwork

	// ErrCoreConfig represents an error in the cli core config, for example some basic
	// files shipped with the installation are missing, or cannot create or get basic
	// directories vital for the CLI to work. (6)
	ErrCoreConfig

	// ErrBadArgument is returned when the arguments are not valid (7)
	ErrBadArgument
)
