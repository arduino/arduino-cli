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
	"net/url"

	paths "github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

type yamlConfig struct {
	ProxyType         string                   `yaml:"proxy_type"`
	ProxyManualConfig *yamlProxyConfig         `yaml:"manual_configs,omitempty"`
	SketchbookPath    string                   `yaml:"sketchbook_path,omitempty"`
	ArduinoDataDir    string                   `yaml:"arduino_data,omitempty"`
	BoardsManager     *yamlBoardsManagerConfig `yaml:"board_manager"`
}

type yamlBoardsManagerConfig struct {
	AdditionalURLS []string `yaml:"additional_urls,omitempty"`
}

type yamlProxyConfig struct {
	Hostname string `yaml:"hostname"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"` // can be encrypted, see issue #71
}

// LoadFromYAML loads the configs from a yaml file.
func (config *Configuration) LoadFromYAML(path *paths.Path) error {
	logrus.Info("Unserializing configurations from ", path)
	content, err := path.ReadFile()
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

	if ret.ArduinoDataDir != "" {
		config.DataDir = paths.New(ret.ArduinoDataDir)
	}
	if ret.SketchbookPath != "" {
		config.SketchbookDir = paths.New(ret.SketchbookPath)
	}
	if ret.ProxyType != "" {
		ProxyType = ret.ProxyType
		if ret.ProxyManualConfig != nil {
			ProxyHostname = ret.ProxyManualConfig.Hostname
			ProxyUsername = ret.ProxyManualConfig.Username
			ProxyPassword = ret.ProxyManualConfig.Password
		}
	}
	if ret.BoardsManager != nil {
		for _, rawurl := range ret.BoardsManager.AdditionalURLS {
			url, err := url.Parse(rawurl)
			if err != nil {
				logrus.WithError(err).Warn("Error parsing config")
				continue
			}
			BoardManagerAdditionalUrls = append(BoardManagerAdditionalUrls, url)
		}
	}
	return nil
}

// SerializeToYAML encodes the current configuration as YAML
func (config *Configuration) SerializeToYAML() ([]byte, error) {
	c := &yamlConfig{}
	if config.SketchbookDir != nil {
		c.SketchbookPath = config.SketchbookDir.String()
	}
	if config.DataDir != nil {
		c.ArduinoDataDir = config.DataDir.String()
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
		c.BoardsManager = &yamlBoardsManagerConfig{AdditionalURLS: []string{}}
		for _, URL := range BoardManagerAdditionalUrls[1:] {
			c.BoardsManager.AdditionalURLS = append(c.BoardsManager.AdditionalURLS, URL.String())
		}
	}
	return yaml.Marshal(c)
}

// SaveToYAML the current configuration to a YAML file
func (config *Configuration) SaveToYAML(path string) error {
	content, err := config.SerializeToYAML()
	if err != nil {
		return fmt.Errorf("econding configuration to YAML: %s", err)
	}

	if err = ioutil.WriteFile(path, content, 0666); err != nil {
		return fmt.Errorf("writing configuration to %s: %s", path, err)
	}
	return nil
}
