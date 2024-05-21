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

package commands

import (
	"context"
	"errors"
	"regexp"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/internal/arduino/globals"
	"github.com/arduino/arduino-cli/internal/i18n"
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

var invalidNames = []string{"CON", "PRN", "AUX", "NUL", "COM0", "COM1", "COM2", "COM3", "COM4", "COM5",
	"COM6", "COM7", "COM8", "COM9", "LPT0", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}

// NewSketch creates a new sketch via gRPC
func (s *arduinoCoreServerImpl) NewSketch(ctx context.Context, req *rpc.NewSketchRequest) (*rpc.NewSketchResponse, error) {
	var sketchesDir string
	if len(req.GetSketchDir()) > 0 {
		sketchesDir = req.GetSketchDir()
	} else {
		sketchesDir = s.settings.GetString("directories.User")
	}

	if err := validateSketchName(req.GetSketchName()); err != nil {
		return nil, err
	}

	sketchDirPath := paths.New(sketchesDir).Join(req.GetSketchName())
	if err := sketchDirPath.MkdirAll(); err != nil {
		return nil, &cmderrors.CantCreateSketchError{Cause: err}
	}
	sketchName := sketchDirPath.Base()
	sketchMainFilePath := sketchDirPath.Join(sketchName + globals.MainFileValidExtension)
	if !req.GetOverwrite() {
		if sketchMainFilePath.Exist() {
			return nil, &cmderrors.CantCreateSketchError{Cause: errors.New(i18n.Tr(".ino file already exists"))}
		}
	}
	if err := sketchMainFilePath.WriteFile(emptySketch); err != nil {
		return nil, &cmderrors.CantCreateSketchError{Cause: err}
	}

	return &rpc.NewSketchResponse{MainFile: sketchMainFilePath.String()}, nil
}

func validateSketchName(name string) error {
	if name == "" {
		return &cmderrors.CantCreateSketchError{Cause: errors.New(i18n.Tr("sketch name cannot be empty"))}
	}
	if len(name) > sketchNameMaxLength {
		return &cmderrors.CantCreateSketchError{Cause: errors.New(i18n.Tr("sketch name too long (%[1]d characters). Maximum allowed length is %[2]d",
			len(name),
			sketchNameMaxLength))}
	}
	if !sketchNameValidationRegex.MatchString(name) {
		return &cmderrors.CantCreateSketchError{Cause: errors.New(i18n.Tr(`invalid sketch name "%[1]s": the first character must be alphanumeric or "_", the following ones can also contain "-" and ".". The last one cannot be ".".`,
			name))}
	}
	for _, invalid := range invalidNames {
		if name == invalid {
			return &cmderrors.CantCreateSketchError{Cause: errors.New(i18n.Tr(`sketch name cannot be the reserved name "%[1]s"`, invalid))}
		}
	}
	return nil
}
