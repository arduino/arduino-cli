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
 * Copyright 2017 BCMI LABS SA (http://www.arduino.cc/)
 */

// Package configs contains all CLI configurations handling.
//
// It is done via a YAML file which can be in a custom location,
// but is defaulted to "$EXECUTABLE_DIR/cli-config.yaml"
package configs

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	properties "github.com/arduino/go-properties-map"
	"github.com/bcmi-labs/arduino-cli/common"

	"gopkg.in/yaml.v2"
)

// FileLocation represents the default location of the config file (same directory as executable).
var FileLocation = getFileLocation()

// BundledInIDE tells if the CLI is paired with the Java Arduino IDE.
var BundledInIDE bool

// ArduinoIDEFolder represents the path of the IDE directory, set only if BundledInIDE = true.
var ArduinoIDEFolder string

func getFileLocation() string {
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
	ProxyType         string        `yaml:"proxy_type,required"`
	ProxyManualConfig *ProxyConfigs `yaml:"manual_configs,omitempty"`
	SketchbookPath    string        `yaml:"sketchbook_path,omitempty"`
	ArduinoDataFolder string        `yaml:"arduino_data,omitempty"`
}

//ProxyConfigs represents a possible manual proxy configuration.
type ProxyConfigs struct {
	Hostname string `yaml:"hostname,required"`
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
	defArduinoData, _ := common.GetDefaultArduinoFolder()
	defSketchbook, _ := common.GetDefaultArduinoHomeFolder()
	if checkIfBundled() != nil {
		BundledInIDE = false
	}

	defaultConfig = Configs{
		ProxyType:         "auto",
		SketchbookPath:    defSketchbook,
		ArduinoDataFolder: defArduinoData,
	}
}

// Default returns a copy of the default configuration.
func Default() Configs {
	return defaultConfig
}

// UnserializeFromIDEPreferences loads the config from an IDE preferences.txt file, by Updating the specified otherConfigs.
func UnserializeFromIDEPreferences(otherConfigs *Configs) error {
	props, err := properties.Load(filepath.Join(otherConfigs.ArduinoDataFolder, "preferences.txt"))
	if err != nil {
		return err
	}
	err = proxyConfigsFromIDEPrefs(otherConfigs, props)
	if err != nil {
		return err
	}
	sketchbookPath, exists := props.SubTree("sketchbook")["path"]
	if exists {
		otherConfigs.SketchbookPath = sketchbookPath
	}
	return nil
}

func proxyConfigsFromIDEPrefs(otherConfigs *Configs, props properties.Map) error {
	proxy := props.SubTree("proxy")
	typo := proxy["type"]
	switch typo {
	case "auto":
		//automatic proxy
		viper.Set("proxy.type", "auto")
		break
	case "manual":
		//manual proxy configuration
		manualConfig := proxy.SubTree("manual")
		hostname, exists := manualConfig["hostname"]
		if !exists {
			return errors.New("Proxy ostname not found in preferences.txt")
		}
		username := manualConfig["username"]
		password := manualConfig["password"]

		otherConfigs.ProxyType = "manual"
		otherConfigs.ProxyManualConfig = &ProxyConfigs{
			Hostname: hostname,
			Username: username,
			Password: password,
		}
		viper.Set("proxy", *otherConfigs.ProxyManualConfig)
		break
	case "none":
		//No proxy
		break
	default:
		return errors.New("Unsupported config type")
	}
	return nil
}

// Unserialize loads the configs from a yaml file.
// Returns the default configuration if there is an
// error.
func Unserialize(path string) (Configs, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return Default(), err
	}
	var ret Configs
	err = yaml.Unmarshal(content, &ret)
	if err != nil {
		return Default(), err
	}
	fixMissingFields(&ret)
	return ret, nil
}

// Serialize creates a file in the specified path with
// corresponds to a config file reflecting the configs.
func (c Configs) Serialize(path string) error {
	content, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, content, 0666)
	if err != nil {
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
		c.ProxyManualConfig = &ProxyConfigs{
			Hostname: def.ProxyManualConfig.Hostname,
			Username: def.ProxyManualConfig.Username,
			Password: def.ProxyManualConfig.Password,
		}
		viper.AutomaticEnv()
		viper.Get("asd")
	}
}

// checkIfBundled checks if the CLI is bundled with the Arduino IDE.
func checkIfBundled() error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		return err
	}
	execParent := filepath.SplitList(filepath.Dir(executable))
	ArduinoIDEFolder := filepath.Join(execParent[0 : len(execParent)-1]...)

	BundledInIDE = false
	executables := []string{"arduino", "arduino.sh", "arduino.exe"}
	for _, exe := range executables {
		exePath := filepath.Join(ArduinoIDEFolder, exe)
		_, err := os.Stat(exePath)
		if !os.IsNotExist(err) {
			BundledInIDE = true
			break
		}
	}
	return nil
}
