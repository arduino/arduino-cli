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

package sketches

import (
	"encoding/json"
	"fmt"

	"github.com/arduino/arduino-cli/arduino/builder"
	"github.com/arduino/go-paths-helper"
	"github.com/pkg/errors"
)

// Sketch is a sketch for Arduino
type Sketch struct {
	Name     string
	FullPath *paths.Path
	Metadata *Metadata
}

// Metadata is the kind of data associated to a project such as the connected board
type Metadata struct {
	CPU BoardMetadata `json:"cpu,omitempty" gorethink:"cpu"`
}

// BoardMetadata represents the board metadata for the sketch
type BoardMetadata struct {
	Fqbn string `json:"fqbn,required"`
	Name string `json:"name,omitempty"`
	Port string `json:"port,omitepty"`
}

// NewSketchFromPath loads a sketch from the specified path
func NewSketchFromPath(path *paths.Path) (*Sketch, error) {
	path, err := path.Abs()
	if err != nil {
		return nil, errors.Errorf("getting sketch path: %s", err)
	}
	if !path.IsDir() {
		path = path.Parent()
	}
	sketchFile := path.Join(path.Base() + ".ino")
	if !sketchFile.Exist() {
		return nil, errors.Errorf("no valid sketch found in %s: missing %s", path, sketchFile.Base())
	}
	sketch := &Sketch{
		FullPath: path,
		Name:     path.Base(),
		Metadata: &Metadata{},
	}
	sketch.ImportMetadata()
	return sketch, nil
}

// ImportMetadata imports metadata into the sketch from a sketch.json file in the root
// path of the sketch.
func (s *Sketch) ImportMetadata() error {
	sketchJSON := s.FullPath.Join("sketch.json")
	content, err := sketchJSON.ReadFile()
	if err != nil {
		return fmt.Errorf("reading sketch metadata %s: %s", sketchJSON, err)
	}
	var meta Metadata
	err = json.Unmarshal(content, &meta)
	if err != nil {
		if s.Metadata == nil {
			s.Metadata = new(Metadata)
		}
		return fmt.Errorf("encoding sketch metadata: %s", err)
	}
	s.Metadata = &meta
	return nil
}

// ExportMetadata writes sketch metadata into a sketch.json file in the root path of
// the sketch
func (s *Sketch) ExportMetadata() error {
	d, err := json.MarshalIndent(&s.Metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("decoding sketch metadata: %s", err)
	}

	sketchJSON := s.FullPath.Join("sketch.json")
	if err := sketchJSON.WriteFile(d); err != nil {
		return fmt.Errorf("writing sketch metadata %s: %s", sketchJSON, err)
	}
	return nil
}

// BuildPath returns this Sketch build path in the temp directory of the system.
// Returns an error if the Sketch's FullPath is not set
func (s *Sketch) BuildPath() (*paths.Path, error) {
	if s.FullPath == nil {
		return nil, fmt.Errorf("sketch path is empty")
	}
	return builder.GenBuildPath(s.FullPath), nil
}
