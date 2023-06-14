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

package sketch

import (
	"context"
	"errors"
	"regexp"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/arduino/globals"
	"github.com/arduino/arduino-cli/configuration"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	paths "github.com/arduino/go-paths-helper"
)

var emptySketch = []byte(`
void setup() {
}

void loop() {
}
`)

// sketchNameMaxLength could be part of the regex, but it's intentionally left out for clearer error reporting
var sketchNameMaxLength = 63
var sketchNameValidationRegex = regexp.MustCompile(`^[0-9a-zA-Z_](?:[0-9a-zA-Z_\.-]*[0-9a-zA-Z_-]|)$`)

// NewSketch creates a new sketch via gRPC
func NewSketch(ctx context.Context, req *rpc.NewSketchRequest) (*rpc.NewSketchResponse, error) {
	var sketchesDir string
	if len(req.SketchDir) > 0 {
		sketchesDir = req.SketchDir
	} else {
		sketchesDir = configuration.Settings.GetString("directories.User")
	}

	if err := validateSketchName(req.SketchName); err != nil {
		return nil, err
	}

	sketchDirPath := paths.New(sketchesDir).Join(req.SketchName)
	if err := sketchDirPath.MkdirAll(); err != nil {
		return nil, &arduino.CantCreateSketchError{Cause: err}
	}
	sketchName := sketchDirPath.Base()
	sketchMainFilePath := sketchDirPath.Join(sketchName + globals.MainFileValidExtension)
	if !req.Overwrite {
		if sketchMainFilePath.Exist() {
			return nil, &arduino.CantCreateSketchError{Cause: errors.New(tr(".ino file already exists"))}
		}
	}
	if err := sketchMainFilePath.WriteFile(emptySketch); err != nil {
		return nil, &arduino.CantCreateSketchError{Cause: err}
	}

	return &rpc.NewSketchResponse{MainFile: sketchMainFilePath.String()}, nil
}

func validateSketchName(name string) error {
	if name == "" {
		return &arduino.CantCreateSketchError{Cause: errors.New(tr("sketch name cannot be empty"))}
	}
	if len(name) > sketchNameMaxLength {
		return &arduino.CantCreateSketchError{Cause: errors.New(tr("sketch name too long (%[1]d characters). Maximum allowed length is %[2]d",
			len(name),
			sketchNameMaxLength))}
	}
	if !sketchNameValidationRegex.MatchString(name) {
		return &arduino.CantCreateSketchError{Cause: errors.New(tr(`invalid sketch name "%[1]s": the first character must be alphanumeric or "_", the following ones can also contain "-" and ".". The last one cannot be ".".`,
			name))}
	}
	return nil
}
