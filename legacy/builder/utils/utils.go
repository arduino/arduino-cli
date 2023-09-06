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

package utils

import (
	"github.com/arduino/arduino-cli/legacy/builder/types"
)

type loggerAction struct {
	onlyIfVerbose bool
	warn          bool
	msg           string
}

func (l *loggerAction) Run(ctx *types.Context) error {
	if !l.onlyIfVerbose || ctx.Verbose {
		if l.warn {
			ctx.Warn(l.msg)
		} else {
			ctx.Info(l.msg)
		}
	}
	return nil
}

func LogIfVerbose(warn bool, msg string) types.Command {
	return &loggerAction{onlyIfVerbose: true, warn: warn, msg: msg}
}
