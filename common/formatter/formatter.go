/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package formatter

import (
	"errors"
	"fmt"

	"github.com/cavaliercoder/grab"
	"github.com/sirupsen/logrus"
)

// Formatter interface represents a generic formatter. It allows to print and format Messages.
type Formatter interface {
	// Format formats a parameter if possible, otherwise it returns an error.
	Format(interface{}) (string, error)

	// DownloadProgressBar outputs a progress bar if possible. Waits until the download ends.
	DownloadProgressBar(resp *grab.Response, prefix string)
}

// PrintFunc represents a function used to print formatted data.
type PrintFunc func(Formatter, interface{}) error

var formatters map[string]Formatter
var defaultFormatter Formatter

var logger *logrus.Logger

var debug bool

func init() {
	formatters = make(map[string]Formatter, 2)
	AddCustomFormatter("text", &TextFormatter{})
	AddCustomFormatter("json", &JSONFormatter{})
	defaultFormatter = formatters["text"]
}

// SetFormatter sets the defaults format to the one specified, if valid. Otherwise it returns an error.
func SetFormatter(formatName string) error {
	if !IsSupported(formatName) {
		return fmt.Errorf("formatter for %s format not implemented", formatName)
	}
	defaultFormatter = formatters[formatName]
	return nil
}

// SetLogger sets the logger for printed errors.
func SetLogger(log *logrus.Logger) {
	logger = log
}

// IsSupported returns whether the format specified is supported or not by the current set of formatters.
func IsSupported(formatName string) bool {
	_, supported := formatters[formatName]
	return supported
}

// IsCurrentFormat returns if the specified format is the one currently set.
func IsCurrentFormat(formatName string) bool {
	return formatters[formatName] == defaultFormatter
}

// AddCustomFormatter adds a custom formatter to the list of available formatters of this package.
//
// If a key is already present, it is replaced and old Value is returned.
//
// If format was not already added as supported, the custom formatter is
// simply added, and oldValue returns nil.
func AddCustomFormatter(formatName string, form Formatter) Formatter {
	oldValue := formatters[formatName]
	formatters[formatName] = form
	return oldValue
}

// Format formats a message formatted using a Formatter specified by SetFormatter(...) function.
func Format(msg interface{}) (string, error) {
	if defaultFormatter == nil {
		return "", errors.New("no formatter set")
	}
	return defaultFormatter.Format(msg)
}

// Print prints a message formatted using a Formatter specified by SetFormatter(...) function.
func Print(msg interface{}) error {
	output, err := defaultFormatter.Format(msg)
	if err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}
