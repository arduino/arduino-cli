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

	"github.com/bcmi-labs/arduino-cli/pathutils"

	"github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"
)

// ConfigFilePath represents the default location of the config file (same directory as executable).
var ConfigFilePath = getDefaultConfigFilePath()

// getDefaultConfigFilePath returns the default path for .cli-config.yml,
// this is the directory where the arduino-cli executable resides.
func getDefaultConfigFilePath() string {
	fileLocation, err := os.Executable()
	if err != nil {
		fileLocation = "."
	}
	fileLocation = filepath.Dir(fileLocation)
	fileLocation = filepath.Join(fileLocation, ".cli-config.yml")
	return fileLocation
}

type yamlConfig struct {
	ProxyType         string           `yaml:"proxy_type"`
	ProxyManualConfig *yamlProxyConfig `yaml:"manual_configs,omitempty"`
	SketchbookPath    string           `yaml:"sketchbook_path,omitempty"`
	ArduinoDataFolder string           `yaml:"arduino_data,omitempty"`
}

type yamlProxyConfig struct {
	Hostname string `yaml:"hostname"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"` // can be encrypted, see issue #71
}

// ProxyType is the type of proxy configured
var ProxyType = "auto"

// ProxyHostname is the proxy hostname
var ProxyHostname string

// ProxyUsername is the proxy user
var ProxyUsername string

// ProxyPassword is the proxy password
var ProxyPassword string

// LoadFromEnv read configurations from the environment variables
func LoadFromEnv() {
	if p, has := os.LookupEnv("PROXY_TYPE"); has {
		ProxyType = p
	}
	if dir, has := os.LookupEnv("SKETCHBOOK_DIR"); has {
		SketchbookFolder = pathutils.NewConstPath(dir)
		ArduinoHomeFolder = SketchbookFolder
	}
	if dir, has := os.LookupEnv("ARDUINO_DATA_DIR"); has {
		ArduinoDataFolder = pathutils.NewConstPath(dir)
	}
}

// Unserialize loads the configs from a yaml file.
func Unserialize(path string) error {
	logrus.Info("Unserializing configurations from ", path)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		logrus.WithError(err).Warn("Error reading config, using default configuration")
		return err
	}
	var ret yamlConfig
	err = yaml.Unmarshal(content, &ret)
	if err != nil {
		logrus.WithError(err).Warn("Error parsing config, using default configuration")
		return err
	}

	if ret.ArduinoDataFolder != "" {
		ArduinoDataFolder = pathutils.NewConstPath(ret.ArduinoDataFolder)
	}
	if ret.SketchbookPath != "" {
		SketchbookFolder = pathutils.NewConstPath(ret.SketchbookPath)
	}
	if ret.ProxyType != "" {
		ProxyType = ret.ProxyType
		if ret.ProxyManualConfig != nil {
			ProxyHostname = ret.ProxyManualConfig.Hostname
			ProxyUsername = ret.ProxyManualConfig.Username
			ProxyPassword = ret.ProxyManualConfig.Password
		}
	}
	return nil
}

// Serialize creates a file in the specified path with
// corresponds to a config file reflecting the configs.
func Serialize(path string) error {
	logrus.Info("Serializing configurations to ", path)
	c := &yamlConfig{}
	if dir, err := SketchbookFolder.Get(); err == nil {
		c.SketchbookPath = dir
	}
	if dir, err := ArduinoDataFolder.Get(); err == nil {
		c.ArduinoDataFolder = dir
	}
	c.ProxyType = ProxyType
	if ProxyType == "manual" {
		c.ProxyManualConfig = &yamlProxyConfig{
			Hostname: ProxyHostname,
			Username: ProxyUsername,
			Password: ProxyPassword,
		}
	}
	content, err := yaml.Marshal(c)
	if err != nil {
		logrus.WithError(err).Warn("Error encoding config")
		return err
	}
	err = ioutil.WriteFile(path, content, 0666)
	if err != nil {
		logrus.WithError(err).Warn("Error writing config")
		return err
	}
	return nil
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
	ideDir := filepath.Dir(executable)
	logrus.Info("Candidate IDE Directory: ", ideDir)

	executables := []string{"arduino", "arduino.sh", "arduino.exe"}
	for _, exe := range executables {
		exePath := filepath.Join(ideDir, exe)
		_, err := os.Stat(exePath)
		if !os.IsNotExist(err) {
			arduinoIDEDirectory = &ideDir
			break
		}
	}

	return *arduinoIDEDirectory != ""
}
