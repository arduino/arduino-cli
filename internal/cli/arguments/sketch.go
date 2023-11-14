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

package arguments

import (
	"context"

	"github.com/arduino/arduino-cli/commands/sketch"
	sk "github.com/arduino/arduino-cli/commands/sketch"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
)

// InitSketchPath returns an instance of paths.Path pointing to sketchPath.
// If sketchPath is an empty string returns the current working directory.
// In both cases it warns the user if he's using deprecated files
func InitSketchPath(path string, printWarnings bool) (sketchPath *paths.Path) {
	if path != "" {
		sketchPath = paths.New(path)
	} else {
		wd, err := paths.Getwd()
		if err != nil {
			feedback.Fatal(tr("Couldn't get current working directory: %v", err), feedback.ErrGeneric)
		}
		logrus.Infof("Reading sketch from dir: %s", wd)
		sketchPath = wd
	}
	if printWarnings {
		if msg := sk.WarnDeprecatedFiles(sketchPath); msg != "" {
			feedback.Warning(msg)
		}
	}
	return sketchPath
}

// GetSketchProfiles is an helper function useful to autocomplete.
// It returns the profile names set in the sketch.yaml
func GetSketchProfiles(sketchPath string) []string {
	if sketchPath == "" {
		if wd, _ := paths.Getwd(); wd != nil && wd.String() != "" {
			sketchPath = wd.String()
		} else {
			return nil
		}
	}
	list, _ := sketch.LoadSketch(context.Background(), &rpc.LoadSketchRequest{
		SketchPath: sketchPath,
	})
	profiles := list.GetProfiles()
	res := make([]string, len(profiles))
	for i, p := range list.GetProfiles() {
		res[i] = p.GetName()
	}
	return res
}
