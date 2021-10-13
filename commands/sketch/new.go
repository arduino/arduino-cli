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
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/configuration"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

var emptySketch = []byte(`
void setup() {
}

void loop() {
}
`)

// CreateSketch creates a new sketch
func CreateSketch(sketchDir string, sketchName string) (string, error) {
	if err := os.MkdirAll(sketchDir, os.FileMode(0755)); err != nil {
		return "", err
	}
	baseSketchName := filepath.Base(sketchDir)
	sketchFile := filepath.Join(sketchDir, baseSketchName+".ino")
	if err := ioutil.WriteFile(sketchFile, emptySketch, os.FileMode(0644)); err != nil {
		return "", err
	}
	return sketchFile, nil
}

// NewSketch FIXMEDOC
func NewSketch(ctx context.Context, req *rpc.NewSketchRequest) (*rpc.NewSketchResponse, error) {
	sketchesDir := configuration.Settings.GetString("directories.User")
	sketchDir := filepath.Join(sketchesDir, req.SketchName)
	sketchFile, err := CreateSketch(sketchDir, req.SketchName)
	if err != nil {
		return nil, &commands.CantCreateSketchError{Cause: err}
	}

	return &rpc.NewSketchResponse{MainFile: sketchFile}, nil
}
