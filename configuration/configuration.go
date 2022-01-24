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

	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/i18n"
	paths "github.com/arduino/go-paths-helper"
	"github.com/arduino/go-win32-utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

var tr = i18n.Tr

// Init initialize defaults and read the configuration file.
// Please note the logging system hasn't been configured yet,
// so logging shouldn't be used here.
func Init(configFile string) *viper.Viper {
	jww.SetStdoutThreshold(jww.LevelFatal)

	// Create a new viper instance with default values for all the settings
	settings := viper.New()
	SetDefaults(settings)

	// Set config name and config path
	if configFilePath := paths.New(configFile); configFilePath != nil {
		settings.SetConfigName(strings.TrimSuffix(configFilePath.Base(), configFilePath.Ext()))
		settings.AddConfigPath(configFilePath.Parent().String())
	}

	// Attempt to read config file
	if err := settings.ReadInConfig(); err != nil {
		// ConfigFileNotFoundError is acceptable, anything else
		// should be reported to the user
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			feedback.Errorf(tr("Error reading config file: %v"), err)
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
	if dataDir := os.Getenv("ARDUINO_DATA_DIR"); dataDir != "" {
		return dataDir
	}
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		feedback.Errorf(tr("Unable to get user home dir: %v"), err)
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
			feedback.Errorf(tr("Unable to get Local App Data Folder: %v"), err)
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
		feedback.Errorf(tr("Unable to get user home dir: %v"), err)
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
			feedback.Errorf(tr("Unable to get Documents Folder: %v"), err)
			return "."
		}
		return filepath.Join(documentsPath, "Arduino")
	default:
		return "."
	}
}

// IsBundledInDesktopIDE returns true if the CLI is bundled with the Arduino IDE.
func IsBundledInDesktopIDE(settings *viper.Viper) bool {
	// value is cached the first time we run the check
	if settings.IsSet("IDE.Bundled") {
		return settings.GetBool("IDE.Bundled")
	}

	settings.Set("IDE.Bundled", false)
	settings.Set("IDE.Portable", false)

	executable, err := os.Executable()
	if err != nil {
		feedback.Errorf(tr("Cannot get executable path: %v"), err)
		return false
	}

	executablePath, err := filepath.EvalSymlinks(executable)
	if err != nil {
		feedback.Errorf(tr("Cannot get executable path: %v"), err)
		return false
	}

	ideDir := paths.New(executablePath).Parent()
	logrus.WithField("dir", ideDir).Trace("Candidate IDE directory")

	// To determine if the CLI is bundled with an IDE, We check an arbitrary
	// number of folders that are part of the IDE install tree
	tests := []string{
		"tools-builder",
		"examples/01.Basics/Blink",
	}

	for _, test := range tests {
		if !ideDir.Join(test).Exist() {
			// the test folder doesn't exist or is not accessible
			return false
		}
	}

	// Persist IDE-related config settings
	settings.Set("IDE.Bundled", true)
	settings.Set("IDE.Directory", ideDir)

	// Check whether this is a portable install
	if ideDir.Join("portable").Exist() {
		logrus.Info("The IDE installation is 'portable'")
		settings.Set("IDE.Portable", true)
	}

	return true
}

// FindConfigFileInArgsOrWorkingDirectory returns the config file path using the
// argument '--config-file' (if specified) or looking in the current working dir.
// Returns default path for the current OS if no config file can be found.
func FindConfigFileInArgsOrWorkingDirectory(args []string) string {
	// Look for '--config-file' argument
	for i, arg := range args {
		if arg == "--config-file" {
			if len(args) > i+1 {
				return args[i+1]
			}
		}
	}

	// Look into current working directory
	if cwd, err := paths.Getwd(); err != nil {
		panic(err)
	} else if configFile := searchConfigTree(cwd); configFile != nil {
		return configFile.Join("arduino-cli.yaml").String()
	}
	return paths.New(getDefaultArduinoDataDir(), "arduino-cli.yaml").String()
}

func searchConfigTree(cwd *paths.Path) *paths.Path {
	// go back up to root and search for the config file
	for _, path := range cwd.Parents() {
		if path.Join("arduino-cli.yaml").Exist() {
			return path
		}
	}

	return nil
}
