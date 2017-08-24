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

package configs

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

// Configs represents the possible configurations for the CLI.
type Configs struct {
	HTTPProxy      string `yaml:"HTTP_Proxy,omitempty"`
	SketchBookPath string `yaml:"HTTP_Proxy,omitempty"`
	LibrariesPath  string `yaml:"HTTP_Proxy,omitempty"`
	PackagesPath   string `yaml:"HTTP_Proxy,omitempty"`
}

// defaultConfig represents the default configuration.
var defaultConfig = Configs{
	HTTPProxy:      os.Getenv("HTTP_PROXY"),
	SketchBookPath: "",
	LibrariesPath:  "",
	PackagesPath:   "",
}

var envConfig = Configs{
	HTTPProxy:      os.Getenv("HTTP_PROXY"),
	SketchBookPath: os.Getenv("SKETCHBOOK_FOLDER"),
	LibrariesPath:  os.Getenv("LIBS_FOLDER"),
	PackagesPath:   os.Getenv("PACKAGES_FOLDER"),
}

// Default returns a copy of the default configuration.
func Default() Configs {
	return defaultConfig
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
	//env := envConfig

	if c.HTTPProxy == "" {
		c.HTTPProxy = def.HTTPProxy
	}
	if c.SketchBookPath == "" {
		c.SketchBookPath = def.SketchBookPath
	}
	if c.LibrariesPath == "" {
		c.LibrariesPath = def.LibrariesPath
	}
	if c.PackagesPath == "" {
		c.PackagesPath = def.PackagesPath
	}
}
