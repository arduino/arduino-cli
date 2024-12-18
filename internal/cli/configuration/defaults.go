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
	"os"
	"time"
)

// SetDefaults sets the default values for certain keys
func SetDefaults(settings *Settings) {
	setKeyTypeSchema := func(k string, v any) {
		settings.SetKeyTypeSchema(k, v)
		settings.Defaults.SetKeyTypeSchema(k, v)
	}
	setDefaultValueAndKeyTypeSchema := func(k string, v any) {
		setKeyTypeSchema(k, v)
		settings.Defaults.Set(k, v)
	}

	// logging
	setDefaultValueAndKeyTypeSchema("logging.level", "info")
	setDefaultValueAndKeyTypeSchema("logging.format", "text")
	setKeyTypeSchema("logging.file", "")

	// Libraries
	setDefaultValueAndKeyTypeSchema("library.enable_unsafe_install", false)

	// Boards Manager
	setDefaultValueAndKeyTypeSchema("board_manager.additional_urls", []string{})

	// arduino directories
	setDefaultValueAndKeyTypeSchema("directories.data", getDefaultArduinoDataDir())
	setDefaultValueAndKeyTypeSchema("directories.downloads", "")
	setDefaultValueAndKeyTypeSchema("directories.user", getDefaultUserDir())
	setKeyTypeSchema("directories.builtin.libraries", "")

	// Sketch compilation
	setDefaultValueAndKeyTypeSchema("sketch.always_export_binaries", false)
	setDefaultValueAndKeyTypeSchema("build_cache.ttl", (time.Hour * 24 * 30).String())
	setDefaultValueAndKeyTypeSchema("build_cache.compilations_before_purge", uint(10))
	setDefaultValueAndKeyTypeSchema("build_cache.path", getDefaultBuildCacheDir())
	setKeyTypeSchema("build_cache.extra_paths", []string{})

	// daemon settings
	setDefaultValueAndKeyTypeSchema("daemon.port", "50051")

	// metrics settings
	setDefaultValueAndKeyTypeSchema("metrics.enabled", true)
	setDefaultValueAndKeyTypeSchema("metrics.addr", ":9090")

	// output settings
	setDefaultValueAndKeyTypeSchema("output.no_color", false)

	// updater settings
	setDefaultValueAndKeyTypeSchema("updater.enable_notification", true)

	// network settings
	setKeyTypeSchema("network.proxy", "")
	setKeyTypeSchema("network.user_agent_ext", "")
	setDefaultValueAndKeyTypeSchema("network.connection_timeout", (time.Second * 60).String())
	// network: Arduino Cloud API settings
	setKeyTypeSchema("network.cloud_api.skip_board_detection_calls", false)

	// locale
	setKeyTypeSchema("locale", "")
}

// InjectEnvVars change settings based on the environment variables values
func InjectEnvVars(settings *Settings) {
	// Bind env vars
	settings.InjectEnvVars(os.Environ(), "ARDUINO")

	// Bind env aliases to keep backward compatibility
	setIfEnvExists := func(key, env string) {
		if v, ok := os.LookupEnv(env); ok {
			settings.SetFromENV(key, v)
		}
	}
	setIfEnvExists("library.enable_unsafe_install", "ARDUINO_ENABLE_UNSAFE_LIBRARY_INSTALL")
	setIfEnvExists("directories.user", "ARDUINO_SKETCHBOOK_DIR")
	setIfEnvExists("directories.downloads", "ARDUINO_DOWNLOADS_DIR")
	setIfEnvExists("directories.data", "ARDUINO_DATA_DIR")
	setIfEnvExists("sketch.always_export_binaries", "ARDUINO_SKETCH_ALWAYS_EXPORT_BINARIES")
}
