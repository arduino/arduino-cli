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
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/go-configmap"
	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-win32-utils"
)

// Settings contains the configuration of the Arduino CLI core service
type Settings struct {
	*configmap.Map
	Defaults *configmap.Map
}

// NewSettings creates a new instance of Settings with the default values set
func NewSettings() *Settings {
	res := &Settings{
		Map:      configmap.New(),
		Defaults: configmap.New(),
	}
	SetDefaults(res)
	return res
}

var userProvidedDefaultDataDir *string

// getDefaultArduinoDataDir returns the full path to the default arduino folder
func getDefaultArduinoDataDir() string {
	// This is overridden by --config-dir flag
	if userProvidedDefaultDataDir != nil {
		return *userProvidedDefaultDataDir
	}

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		feedback.Warning(i18n.Tr("Unable to get user home dir: %v", err))
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
			feedback.Warning(i18n.Tr("Unable to get Local App Data Folder: %v", err))
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
		feedback.Warning(i18n.Tr("Unable to get user home dir: %v", err))
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
			feedback.Warning(i18n.Tr("Unable to get Documents Folder: %v", err))
			return "."
		}
		return filepath.Join(documentsPath, "Arduino")
	default:
		return "."
	}
}

// FindConfigFlagsInArgsOrFallbackOnEnv returns the config file path using the
// argument '--config-file' (if specified), if empty looks for the ARDUINO_CONFIG_FILE env,
// or looking in the current working dir
func FindConfigFlagsInArgsOrFallbackOnEnv(args []string) string {
	// Look for '--config-dir' argument
	for i, arg := range args {
		if arg == "--config-dir" {
			if len(args) > i+1 {
				absArgs, err := paths.New(args[i+1]).Abs()
				if err != nil {
					feedback.FatalError(fmt.Errorf("invalid --config-dir value: %w", err), feedback.ErrBadArgument)
				}
				configDir := absArgs.String()
				userProvidedDefaultDataDir = &configDir
				break
			}
		}
	}

	// Look for '--config-file' argument
	for i, arg := range args {
		if arg == "--config-file" {
			if len(args) > i+1 {
				return args[i+1]
			}
		}
	}
	if p, ok := os.LookupEnv("ARDUINO_CONFIG_FILE"); ok {
		return p
	}
	if p, ok := os.LookupEnv("ARDUINO_DIRECTORIES_DATA"); ok {
		return filepath.Join(p, "arduino-cli.yaml")
	}
	if p, ok := os.LookupEnv("ARDUINO_DATA_DIR"); ok {
		return filepath.Join(p, "arduino-cli.yaml")
	}
	return filepath.Join(getDefaultArduinoDataDir(), "arduino-cli.yaml")
}
