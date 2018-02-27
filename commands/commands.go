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
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */

package commands

import (
	"github.com/sirupsen/logrus"
	"github.com/bcmi-labs/arduino-cli/common/releases"
	"github.com/bcmi-labs/arduino-cli/common/formatter"
)

// Error codes to be used for os.Exit().
const (
	ErrNoConfigFile int = iota
	ErrBadCall      int = iota
	ErrGeneric      int = iota
	ErrNetwork      int = iota
	ErrCoreConfig   int = iota // Represents an error in the cli core config, for example some basic files shipped with the installation are missing, or cannot create or get basic folder vital for the CLI to work.

	Version = "0.1.0-alpha.preview"
)

// ErrLogrus represents the logrus instance, which has the role to
// log all non info messages.
var ErrLogrus = logrus.New()

// GlobalFlags represents flags available in all the program.
var GlobalFlags struct {
	Debug  bool   // If true, dump debug output to stderr.
	Format string // The Output format (e.g. text, json).
}

// FIXME: Move away? Where should the display logic reside; in the formatter? That causes an import cycle BTW...
func GenerateDownloadProgressFormatter() releases.ParallelDownloadProgressHandler {
	if formatter.IsCurrentFormat("text") {
		return &ProgressBarFormatter{}
	}
	return nil
}
