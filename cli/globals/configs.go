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
// a commercial license. Buying such a license is mandatory if you want to modify or
// otherwise use the software for commercial activities involving the Arduino
// software without disclosing the source code of your own applications. To purchase
// a commercial license, send an email to license@arduino.cc.

package globals

import (
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/configs"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
)

// InitConfigs initializes the configuration from the specified file.
func InitConfigs() {
	// Start with default configuration
	if conf, err := configs.NewConfiguration(); err != nil {
		logrus.WithError(err).Error("Error creating default configuration")
		feedback.Errorf("Error creating default configuration: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	} else {
		Config = conf
	}

	// Read configuration from global config file
	logrus.Info("Checking for config file in: " + Config.ConfigFile.String())
	if Config.ConfigFile.Exist() {
		readConfigFrom(Config.ConfigFile)
	}

	if Config.IsBundledInDesktopIDE() {
		logrus.Info("CLI is bundled into the IDE")
		err := Config.LoadFromDesktopIDEPreferences()
		if err != nil {
			logrus.WithError(err).Warn("Did not manage to get config file of IDE, using default configuration")
		}
	} else {
		logrus.Info("CLI is not bundled into the IDE")
	}

	// Read configuration from parent folders (project config)
	if pwd, err := paths.Getwd(); err != nil {
		logrus.WithError(err).Warn("Did not manage to find current path")
		if path := paths.New("arduino-yaml"); path.Exist() {
			readConfigFrom(path)
		}
	} else {
		Config.Navigate(pwd)
	}

	// Read configuration from old configuration file if found, but output a warning.
	if old := paths.New(".cli-config.yml"); old.Exist() {
		logrus.Errorf("Old configuration file detected: %s.", old)
		logrus.Info("The name of this file has been changed to `arduino-cli.yaml`, please rename the file fix it.")
		feedback.Error(
			fmt.Errorf("WARNING: Old configuration file detected: %s", old),
			"The name of this file has been changed to `arduino-yaml`, in a future release we will not support"+
				"the old name `.cli-config.yml` anymore. Please rename the file to `arduino-cli.yaml` to silence this warning.")
		readConfigFrom(old)
	}

	// Read configuration from environment vars
	Config.LoadFromEnv()

	// Read configuration from user specified file
	if YAMLConfigFile != "" {
		Config.ConfigFile = paths.New(YAMLConfigFile)
		readConfigFrom(Config.ConfigFile)
	}

	logrus.Info("Configuration set")
}

func readConfigFrom(path *paths.Path) {
	logrus.Infof("Reading configuration from %s", path)
	if err := Config.LoadFromYAML(path); err != nil {
		logrus.WithError(err).Warnf("Could not read configuration from %s", path)
	}
}
