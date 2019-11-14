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
	"os"
	"path/filepath"
	"runtime"

	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/go-win32-utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Init initialize defaults and read the configuration file
func Init() {
	// Config file metadata
	viper.SetConfigName("arduino-cli")

	// Add paths where to search for a config file
	configPath := GetDefaultArduinoDataDir()
	logrus.Infof("Checking for config file in: %s", configPath)
	viper.AddConfigPath(configPath)

	// Set configuration defaults
	setDefaults()

	// Attempt to read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logrus.Info("Config file not found, using default values")
		} else {
			feedback.Errorf("Error reading config file: %v", err)
		}
	}

	// Bind env vars
	viper.SetEnvPrefix("ARDUINO")
	viper.AutomaticEnv()
}

// GetDefaultArduinoDataDir returns the full path to the default arduino folder
func GetDefaultArduinoDataDir() string {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		logrus.Errorf("Unable to get user home dir: %v", err)
		return "."
	}

	switch runtime.GOOS {
	case "linux":
		return filepath.Join(userHomeDir, ".arduino15")
	case "darwin":
		return filepath.Join(userHomeDir, "Library", "Arduino15")
	case "windows":
		localAppDataPath, err := win32.GetLocalAppDataFolder()
		if err != nil {
			logrus.Errorf("Unable to get Local App Data Folder: %v", err)
			return "."
		}
		return filepath.Join(localAppDataPath, "Arduino15")
	default:
		return "."
	}
}
