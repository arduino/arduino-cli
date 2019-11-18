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
	"strings"

	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/go-win32-utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Init initialize defaults and read the configuration file
func Init(configPath string) {
	// Config file metadata
	viper.SetConfigName("arduino-cli")

	// Get default data path if none was provided
	if configPath == "" {
		configPath = getDefaultArduinoDataDir()
	}

	// Add paths where to search for a config file
	logrus.Infof("Checking for config file in: %s", configPath)
	viper.AddConfigPath(configPath)

	// Set configuration defaults
	setDefaults(configPath, getDefaultSketchbookDir())

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
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Bind env aliases to keep backward compatibility
	viper.BindEnv("directories.Sketchbook", "ARDUINO_SKETCHBOOK_DIR")
	viper.BindEnv("directories.Downloads", "ARDUINO_DOWNLOADS_DIR")
	viper.BindEnv("directories.Data", "ARDUINO_DATA_DIR")
}

// getDefaultArduinoDataDir returns the full path to the default arduino folder
func getDefaultArduinoDataDir() string {
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

// getDefaultSketchbookDir returns the full path to the default sketchbook folder
func getDefaultSketchbookDir() string {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		logrus.Errorf("Unable to get user home dir: %v", err)
		return "."
	}

	switch runtime.GOOS {
	case "linux":
		return filepath.Join(userHomeDir, "Arduino")
	case "darwin":
		return filepath.Join(userHomeDir, "Documents", "Arduino")
	case "windows":
		documentsPath, err := win32.GetDocumentsFolder()
		if err != nil {
			logrus.Errorf("Unable to get Documents Folder: %v", err)
			return "."
		}
		return filepath.Join(documentsPath, "Arduino")
	default:
		return "."
	}
}

// IsBundledInDesktopIDE returns true if the CLI is bundled with the Arduino IDE.
func IsBundledInDesktopIDE() bool {
	// value is cached the first time we run the check
	if viper.IsSet("IDE.Bundled") {
		return viper.GetBool("IDE.Bundled")
	}

	viper.Set("IDE.Bundled", false)
	viper.Set("IDE.Portable", false)

	logrus.Info("Checking if CLI is Bundled into the IDE")
	executable, err := os.Executable()
	if err != nil {
		logrus.WithError(err).Warn("Cannot get executable path")
		return viper.GetBool("IDE.Bundled")
	}

	executablePath, err := filepath.EvalSymlinks(executable)
	if err != nil {
		logrus.WithError(err).Warn("Cannot get executable path")
		return viper.GetBool("IDE.Bundled")
	}

	ideDir := filepath.Dir(executablePath)
	logrus.Info("Candidate IDE Directory: ", ideDir)

	// We check an arbitrary number of folders that are part of the IDE
	// install tree
	tests := []string{
		"tools-builder",
		"examples/01.Basics/Blink",
		"portable",
	}

	for _, test := range tests {
		if _, err := os.Stat(filepath.Join(ideDir, test)); err != nil {
			// the test folder doesn't exist or is not accessible
			return viper.GetBool("IDE.Bundled")
		}

		if test == "portable" {
			logrus.Info("IDE is portable")
			viper.Set("IDE.Portable", true)
		}
	}

	viper.Set("IDE.Directory", ideDir)
	return true
}
