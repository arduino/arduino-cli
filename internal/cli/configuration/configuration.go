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
	"path/filepath"
	"runtime"
	"strings"

	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/i18n"
	paths "github.com/arduino/go-paths-helper"
	"github.com/arduino/go-win32-utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Settings is a global instance of viper holding configurations for the CLI and the gRPC consumers
var Settings *viper.Viper

var tr = i18n.Tr

// Init initialize defaults and read the configuration file.
// Please note the logging system hasn't been configured yet,
// so logging shouldn't be used here.
func Init(configFile string) *viper.Viper {
	// Create a new viper instance with default values for all the settings
	settings := viper.New()
	SetDefaults(settings)

	// Set config name and config path
	if configFilePath := paths.New(configFile); configFilePath != nil {
		settings.SetConfigName(strings.TrimSuffix(configFilePath.Base(), configFilePath.Ext()))
		settings.AddConfigPath(configFilePath.Parent().String())
	} else {
		configDir := settings.GetString("directories.Data")
		// Get default data path if none was provided
		if configDir == "" {
			configDir = getDefaultArduinoDataDir()
		}

		settings.SetConfigName("arduino-cli")
		settings.AddConfigPath(configDir)
	}

	// Attempt to read config file
	if err := settings.ReadInConfig(); err != nil {
		// ConfigFileNotFoundError is acceptable, anything else
		// should be reported to the user
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			feedback.Warning(tr("Error reading config file: %v", err))
		}
	}

	return settings
}

// BindFlags creates all the flags binding between the cobra Command and the instance of viper
func BindFlags(cmd *cobra.Command, settings *viper.Viper) {
	settings.BindPFlag("logging.level", cmd.Flag("log-level"))
	settings.BindPFlag("logging.file", cmd.Flag("log-file"))
	settings.BindPFlag("logging.format", cmd.Flag("log-format"))
	settings.BindPFlag("board_manager.additional_urls", cmd.Flag("additional-urls"))
	settings.BindPFlag("output.no_color", cmd.Flag("no-color"))
}

// getDefaultArduinoDataDir returns the full path to the default arduino folder
func getDefaultArduinoDataDir() string {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		feedback.Warning(tr("Unable to get user home dir: %v", err))
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
			feedback.Warning(tr("Unable to get Local App Data Folder: %v", err))
			return "."
		}
		return filepath.Join(localAppDataPath, "Arduino15")
	default:
		return "."
	}
}

// getDefaultUserDir returns the full path to the default user folder
func getDefaultUserDir() string {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		feedback.Warning(tr("Unable to get user home dir: %v", err))
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
			feedback.Warning(tr("Unable to get Documents Folder: %v", err))
			return "."
		}
		return filepath.Join(documentsPath, "Arduino")
	default:
		return "."
	}
}

// GetDefaultBuiltinLibrariesDir returns the full path to the default builtin libraries dir
func GetDefaultBuiltinLibrariesDir() string {
	return filepath.Join(getDefaultArduinoDataDir(), "libraries")
}

// FindConfigFileInArgsFallbackOnEnv returns the config file path using the
// argument '--config-file' (if specified), if empty looks for the ARDUINO_CONFIG_FILE env,
// or looking in the current working dir
func FindConfigFileInArgsFallbackOnEnv(args []string) string {
	// Look for '--config-file' argument
	for i, arg := range args {
		if arg == "--config-file" {
			if len(args) > i+1 {
				return args[i+1]
			}
		}
	}
	return os.Getenv("ARDUINO_CONFIG_FILE")
}
