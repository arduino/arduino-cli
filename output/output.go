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

package output

import (
	"os"
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/i18n"
	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
)

var tr = i18n.Tr

// Setup the feedback system
func Setup(outputFormat string, noColor bool) {
	// Set feedback format output
	outputFormat = strings.ToLower(outputFormat)
	output.OutputFormat = outputFormat
	format, found := parseFormatString(outputFormat)
	if !found {
		feedback.Errorf(tr("Invalid output format: %s"), outputFormat)
		os.Exit(errorcodes.ErrBadCall)
	}
	feedback.SetFormat(format)

	// Set default feedback output to colorable
	feedback.SetOut(colorable.NewColorableStdout())
	feedback.SetErr(colorable.NewColorableStderr())

	// https://no-color.org/
	color.NoColor = noColor

}

func parseFormatString(arg string) (feedback.OutputFormat, bool) {
	f, found := map[string]feedback.OutputFormat{
		"json":     feedback.JSON,
		"jsonmini": feedback.JSONMini,
		"text":     feedback.Text,
		"yaml":     feedback.YAML,
	}[strings.ToLower(arg)]

	return f, found
}
