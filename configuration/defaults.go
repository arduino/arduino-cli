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

package configuration

import (
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// SetDefaults sets the default values for certain keys
func SetDefaults(settings *viper.Viper) {
	// logging
	settings.SetDefault("logging.level", "info")
	settings.SetDefault("logging.format", "text")

	// Libraries
	settings.SetDefault("library.enable_unsafe_install", false)

	// Boards Manager
	settings.SetDefault("board_manager.additional_urls", []string{})

	// arduino directories
	settings.SetDefault("directories.Data", getDefaultArduinoDataDir())
	settings.SetDefault("directories.Downloads", filepath.Join(getDefaultArduinoDataDir(), "staging"))
	settings.SetDefault("directories.User", getDefaultUserDir())

	// Sketch compilation
	settings.SetDefault("sketch.always_export_binaries", false)

	// daemon settings
	settings.SetDefault("daemon.port", "50051")

	// metrics settings
	settings.SetDefault("metrics.enabled", true)
	settings.SetDefault("metrics.addr", ":9090")

	// output settings
	settings.SetDefault("output.no_color", false)

	// updater settings
	settings.SetDefault("updater.enable_notification", true)

	// Bind env vars
	settings.SetEnvPrefix("ARDUINO")
	settings.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	settings.AutomaticEnv()

	// Bind env aliases to keep backward compatibility
	settings.BindEnv("library.enable_unsafe_install", "ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL")
	settings.BindEnv("directories.User", "ARDUINO_SKETCHBOOK_DIR")
	settings.BindEnv("directories.Downloads", "ARDUINO_DOWNLOADS_DIR")
	settings.BindEnv("directories.Data", "ARDUINO_DATA_DIR")
	settings.BindEnv("sketch.always_export_binaries", "ARDUINO_SKETCH_ALWAYS_EXPORT_BINARIES")
}
