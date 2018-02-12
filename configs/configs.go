/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */

// Package configs contains all CLI configurations handling.
//
// It is done via a YAML file which can be in a custom location,
// but is defaulted to "$EXECUTABLE_DIR/cli-config.yaml"
package configs

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/spf13/viper"

	"gopkg.in/yaml.v2"
)

// ConfigFilePath represents the default location of the config file (same directory as executable).
var ConfigFilePath = detectConfigFilePath()

func detectConfigFilePath() string {
	fileLocation, err := os.Executable()
	if err != nil {
		fileLocation = "."
	}
	fileLocation = filepath.Dir(fileLocation)
	fileLocation = filepath.Join(fileLocation, ".cli-config.yml")
	return fileLocation
}

// Configs represents the possible configurations for the CLI.
type Configs struct {
	ProxyType         string        `yaml:"proxy_type"`
	ProxyManualConfig *ProxyConfigs `yaml:"manual_configs,omitempty"`
	SketchbookPath    string        `yaml:"sketchbook_path,omitempty"`
	ArduinoDataFolder string        `yaml:"arduino_data,omitempty"`
}

// ProxyConfigs represents a possible manual proxy configuration.
type ProxyConfigs struct {
	Hostname string `yaml:"hostname"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"` // can be encrypted, see issue #71
}

// defaultConfig represents the default configuration.
var defaultConfig Configs

var envConfig = Configs{
	ProxyType:         os.Getenv("PROXY_TYPE"),
	SketchbookPath:    os.Getenv("SKETCHBOOK_FOLDER"),
	ArduinoDataFolder: os.Getenv("ARDUINO_DATA"),
}

func init() {
	defArduinoData, _ := ArduinoDataFolder.Get()
	defSketchbook, _ := ArduinoHomeFolder.Get()

	defaultConfig = Configs{
		ProxyType:         "auto",
		SketchbookPath:    defSketchbook,
		ArduinoDataFolder: defArduinoData,
	}
}

// Default returns a copy of the default configuration.
func Default() Configs {
	logrus.Info("Returning default configuration")
	return defaultConfig
}

// Unserialize loads the configs from a yaml file.
// Returns the default configuration if there is an
// error.
func Unserialize(path string) (Configs, error) {
	logrus.Info("Unserializing configurations from ", path)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		logrus.WithError(err).Warn("Error during unserialize, using default configuration")
		return Default(), err
	}
	var ret Configs
	err = yaml.Unmarshal(content, &ret)
	if err != nil {
		logrus.WithError(err).Warn("Error during unserialize, using default configuration")
		return Default(), err
	}
	fixMissingFields(&ret)
	return ret, nil
}

// Serialize creates a file in the specified path with
// corresponds to a config file reflecting the configs.
func (c Configs) Serialize(path string) error {
	logrus.Info("Serializing configurations to ", path)
	content, err := yaml.Marshal(c)
	if err != nil {
		logrus.WithError(err).Warn("Error during serialize")
		return err
	}
	err = ioutil.WriteFile(path, content, 0666)
	if err != nil {
		logrus.WithError(err).Warn("Error during serialize")
		return err
	}
	return nil
}

func fixMissingFields(c *Configs) {
	def := defaultConfig
	env := envConfig

	compareAndReplace := func(envVal string, configVal *string, defVal string) {
		if envVal != "" {
			*configVal = envVal
		} else if *configVal == "" {
			*configVal = defVal
		}
	}

	compareAndReplace(env.ProxyType, &c.ProxyType, def.ProxyType)
	compareAndReplace(env.SketchbookPath, &c.SketchbookPath, def.SketchbookPath)
	compareAndReplace(env.ArduinoDataFolder, &c.ArduinoDataFolder, def.ArduinoDataFolder)

	if env.ProxyManualConfig != nil {
		c.ProxyManualConfig = &ProxyConfigs{
			Hostname: env.ProxyManualConfig.Hostname,
			Username: env.ProxyManualConfig.Username,
			Password: env.ProxyManualConfig.Password,
		}
	} else if c.ProxyManualConfig == nil {
		if def.ProxyManualConfig != nil {
			//fmt.Println(def.ProxyManualConfig)
			c.ProxyManualConfig = &ProxyConfigs{
				Hostname: def.ProxyManualConfig.Hostname,
				Username: def.ProxyManualConfig.Username,
				Password: def.ProxyManualConfig.Password,
			}
		}

		viper.AutomaticEnv()
	}
}

var arduinoIDEDirectory *string

// Bundled returns if the CLI is bundled with the Arduino IDE.
func Bundled() bool {
	if arduinoIDEDirectory != nil {
		return *arduinoIDEDirectory != ""
	}
	empty := ""
	arduinoIDEDirectory = &empty

	logrus.Info("Checking if CLI is Bundled into the IDE")
	executable, err := os.Executable()
	if err != nil {
		logrus.WithError(err).Warn("Cannot get executable path")
		return false
	}
	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		logrus.WithError(err).Warn("Cannot get executable path (symlinks error)")
		return false
	}
	ideDir := filepath.Dir(filepath.Dir(executable))
	logrus.Info("Candidate IDE Directory:", ideDir)

	executables := []string{"arduino", "arduino.sh", "arduino.exe"}
	for _, exe := range executables {
		exePath := filepath.Join(ideDir, exe)
		_, err := os.Stat(exePath)
		if !os.IsNotExist(err) {
			arduinoIDEDirectory = &ideDir
			logrus.Info("CLI is bundled:", *arduinoIDEDirectory)
			break
		}
	}

	return *arduinoIDEDirectory != ""
}
