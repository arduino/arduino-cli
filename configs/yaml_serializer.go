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
 * Copyright 2017-2018 ARDUINO AG (http://www.arduino.cc/)
 */

package configs

import (
	"fmt"
	"io/ioutil"

	"github.com/bcmi-labs/arduino-cli/pathutils"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

type yamlConfig struct {
	ProxyType         string           `yaml:"proxy_type"`
	ProxyManualConfig *yamlProxyConfig `yaml:"manual_configs,omitempty"`
	SketchbookPath    string           `yaml:"sketchbook_path,omitempty"`
	ArduinoDataFolder string           `yaml:"arduino_data,omitempty"`
	BoardsManager     struct {
		AdditionalURLS []string `yaml:"additional_urls,omitempty"`
	} `yaml:"board_manager"`
}

type yamlProxyConfig struct {
	Hostname string `yaml:"hostname"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"` // can be encrypted, see issue #71
}

// LoadFromYAML loads the configs from a yaml file.
func LoadFromYAML(path string) error {
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

// SerializeToYAML encodes the current configuration as YAML
func SerializeToYAML() ([]byte, error) {
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
	if len(BoardManagerAdditionalUrls) > 1 {
		c.BoardsManager.AdditionalURLS = []string{}
		for _, URL := range BoardManagerAdditionalUrls[1:] {
			c.BoardsManager.AdditionalURLS = append(c.BoardsManager.AdditionalURLS, URL.String())
		}
	}
	return yaml.Marshal(c)
}

// SaveToYAML the current configuration to a YAML file
func SaveToYAML(path string) error {
	content, err := SerializeToYAML()
	if err != nil {
		return fmt.Errorf("econding configuration to YAML: %s", err)
	}

	if err = ioutil.WriteFile(path, content, 0666); err != nil {
		return fmt.Errorf("writing configuration to %s: %s", path, err)
	}
	return nil
}
