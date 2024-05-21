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

	f "github.com/arduino/arduino-cli/internal/algorithms"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
)

// InitSketchPath returns an instance of paths.Path pointing to sketchPath.
// If sketchPath is an empty string returns the current working directory.
func InitSketchPath(path string) (sketchPath *paths.Path) {
	if path != "" {
		sketchPath = paths.New(path)
	} else {
		wd, err := paths.Getwd()
		if err != nil {
			feedback.Fatal(i18n.Tr("Couldn't get current working directory: %v", err), feedback.ErrGeneric)
		}
		logrus.Infof("Reading sketch from dir: %s", wd)
		sketchPath = wd
	}
	return sketchPath
}

// GetSketchProfiles is an helper function useful to autocomplete.
// It returns the profile names set in the sketch.yaml
func GetSketchProfiles(ctx context.Context, srv rpc.ArduinoCoreServiceServer, sketchPath string) []string {
	if sketchPath == "" {
		if wd, _ := paths.Getwd(); wd != nil && wd.String() != "" {
			sketchPath = wd.String()
		} else {
			return nil
		}
	}
	resp, err := srv.LoadSketch(ctx, &rpc.LoadSketchRequest{SketchPath: sketchPath})
	if err != nil {
		return nil
	}
	profiles := resp.GetSketch().GetProfiles()
	return f.Map(profiles, (*rpc.SketchProfile).GetName)
}
