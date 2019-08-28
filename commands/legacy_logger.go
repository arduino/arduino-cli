// This file is part of arduino-cli.
//
// Copyright 2019 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to modify or
// otherwise use the software for commercial activities involving the Arduino
// software without disclosing the source code of your own applications. To purchase
// a commercial license, send an email to license@arduino.cc.

package commands

import (
	"fmt"
	"io"

	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/i18n"
	"github.com/sirupsen/logrus"
)

// LegacyLogger wraps logrus but can be used within the `legacy` package
type LegacyLogger struct{}

// Fprintln is supposed to let the caller decide the target, let's not do it
// and just forward to Println
func (l LegacyLogger) Fprintln(w io.Writer, level string, format string, a ...interface{}) {
	l.Println(level, format, a...)
}

// Println is a regular log call
func (l LegacyLogger) Println(level string, format string, a ...interface{}) {
	msg := legacyFormat(format, a...)

	switch level {
	case constants.LOG_LEVEL_INFO:
		logrus.Info(msg)
	case constants.LOG_LEVEL_DEBUG:
		logrus.Debug(msg)
	case constants.LOG_LEVEL_ERROR:
		logrus.Error(msg)
	case constants.LOG_LEVEL_WARN:
		logrus.Warn(msg)
	default:
		logrus.Trace(msg)
	}
}

// UnformattedFprintln will do the same as Fprintln
func (l LegacyLogger) UnformattedFprintln(w io.Writer, str string) {
	l.Println(constants.LOG_LEVEL_INFO, str)
}

// UnformattedWrite will do the same as Fprintln
func (l LegacyLogger) UnformattedWrite(w io.Writer, data []byte) {
	l.Println(constants.LOG_LEVEL_INFO, string(data))
}

// Flush is a noop
func (l LegacyLogger) Flush() string {
	return ""
}

// Name doesn't matter
func (l LegacyLogger) Name() string {
	return "legacy"
}

func legacyFormat(format string, a ...interface{}) string {
	format = i18n.FromJavaToGoSyntax(format)
	message := fmt.Sprintf(format, a...)
	return message
}
