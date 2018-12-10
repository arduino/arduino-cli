/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package sketches

import (
	"github.com/arduino/go-paths-helper"
)

// SketchBook is a sketchbook
type SketchBook struct {
	Path *paths.Path
}

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

// NewSketchBook returns a new SketchBook object
func NewSketchBook(path *paths.Path) *SketchBook {
	return &SketchBook{
		Path: path,
	}
}

// NewSketch loads a sketch from the sketchbook
func (sketchbook *SketchBook) NewSketch(name string) (*sketches.Sketch, error) {
	sketch := sketches.Sketch{
		FullPath: sketchbook.Path.Join(name).String(),
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
