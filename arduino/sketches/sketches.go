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

package sketches

import (
	"path/filepath"

	"github.com/arduino/go-paths-helper"

	"github.com/bcmi-labs/arduino-cli/configs"
	"github.com/bcmi-labs/arduino-modules/sketches"
)

// Sketch is a sketch for Arduino
type Sketch struct {
	Name          string
	Path          string
	BoardMetadata *BoardMetadata `json:"board"`
}

// BoardMetadata represents the board metadata for the sketch
type BoardMetadata struct {
	Fqbn string `json:"fqbn,required"`
	Name string `json:"name,required"`
}

// NewSketchFromCurrentSketchbook loads a sketch from the sketchbook
func NewSketchFromCurrentSketchbook(name string) (*sketches.Sketch, error) {
	sketchbookLocation, err := configs.ArduinoHomeFolder.Get()
	if err != nil {
		return nil, err
	}
	sketch := sketches.Sketch{
		FullPath: filepath.Join(sketchbookLocation, name),
		Name:     name,
	}
	sketch.ImportMetadata()
	return &sketch, nil
}

// NewSketchFromPath loads a sketch from the specified path
func NewSketchFromPath(path *paths.Path) (*sketches.Sketch, error) {
	sketch := sketches.Sketch{
		FullPath: path.String(),
		Name:     path.Base(),
	}
	sketch.ImportMetadata()
	return &sketch, nil
}
