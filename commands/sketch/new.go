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

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/arduino/globals"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	paths "github.com/arduino/go-paths-helper"
)

var emptySketch = []byte(`
void setup() {
}

void loop() {
}
`)

// NewSketch creates a new sketch via gRPC
func NewSketch(ctx context.Context, req *rpc.NewSketchRequest) (*rpc.NewSketchResponse, error) {
	var sketchesDir string
	if len(req.SketchDir) > 0 {
		sketchesDir = req.SketchDir
	} else {
		instance := commands.GetInstance(req.Instance.Id)
		sketchesDir = instance.Settings.GetString("directories.User")
	}
	sketchDirPath := paths.New(sketchesDir).Join(req.SketchName)
	if err := sketchDirPath.MkdirAll(); err != nil {
		return nil, &arduino.CantCreateSketchError{Cause: err}
	}
	sketchName := sketchDirPath.Base()
	sketchMainFilePath := sketchDirPath.Join(sketchName + globals.MainFileValidExtension)
	if err := sketchMainFilePath.WriteFile(emptySketch); err != nil {
		return nil, &arduino.CantCreateSketchError{Cause: err}
	}

	return &rpc.NewSketchResponse{MainFile: sketchMainFilePath.String()}, nil
}
