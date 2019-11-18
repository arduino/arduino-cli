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
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package configuration

import (
	"path/filepath"

	"github.com/spf13/viper"
)

func setDefaults(dataDir, sketchBookDir string) {
	// board manager settings
	viper.SetDefault("board_manager.additional_urls", []string{})

	// arduino directories
	viper.SetDefault("directories.Data", dataDir)
	viper.SetDefault("directories.Downloads", filepath.Join(dataDir, "staging"))
	viper.SetDefault("directories.Packages", filepath.Join(dataDir, "packages"))
	viper.SetDefault("directories.SketchBook", sketchBookDir)
	viper.SetDefault("directories.Libraries", filepath.Join(sketchBookDir, "libraries"))
}
