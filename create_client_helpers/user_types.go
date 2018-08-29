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

package createclient

import (
	"encoding/base64"
	"io/ioutil"
	"path/filepath"

	"github.com/bcmi-labs/arduino-modules/sketches"
)

// File is a file saved on the virtual filesystem.
type File struct {
	// The contents of the file, encoded in base64.
	Data *string `form:"data,omitempty" json:"data,omitempty" xml:"data,omitempty"`
	// The name of the file.
	Name string `form:"name" json:"name" xml:"name"`
}

// Sketch is a program meant to be uploaded onto a board.
type Sketch struct {
	// The name of the sketch.
	Name string `form:"name" json:"name" xml:"name"`
	// The other files contained in the sketch.
	Files []*File `form:"files,omitempty" json:"files,omitempty" xml:"files,omitempty"`
	// The folder path where the sketch is saved.
	Folder *string `form:"folder,omitempty" json:"folder,omitempty" xml:"folder,omitempty"`
	// The main file of the sketch.
	Ino      *File           `form:"ino" json:"ino" xml:"ino"`
	Metadata *SketchMetadata `form:"metadata,omitempty" json:"metadata,omitempty" xml:"metadata,omitempty"`
	// The username of the owner of the sketch.
	Owner *string `form:"owner,omitempty" json:"owner,omitempty" xml:"owner,omitempty"`
	// A private sketch is only visible to its owner.
	Private bool `form:"private" json:"private" xml:"private"`
	// A list of links to hackster tutorials.
	Tutorials []string `form:"tutorials,omitempty" json:"tutorials,omitempty" xml:"tutorials,omitempty"`
	// A list of tags. The builtin tag means that it's a builtin example.
	Types []string `form:"types,omitempty" json:"types,omitempty" xml:"types,omitempty"`
}

//ConvertFrom converts from a local sketch to an Arduino Create sketch.
func ConvertFrom(sketch sketches.Sketch) *Sketch {
	_, inoPath := filepath.Split(sketch.Ino.Path)
	content, err := ioutil.ReadFile(filepath.Join(sketch.FullPath, inoPath))
	if err != nil {
		return nil
	}

	ino := base64.StdEncoding.EncodeToString(content)
	ret := Sketch{
		Name:   sketch.Name,
		Folder: &sketch.Path,
		Ino: &File{
			Data: &ino,
			Name: sketch.Ino.Name,
		},
		Private:   sketch.Private,
		Tutorials: sketch.Tutorials,
		Types:     sketch.Types,
		Metadata:  ConvertMetadataFrom(sketch.Metadata),
	}
	for _, f := range sketch.Files {
		if f.Name == "sketch.json" { // Skipping sketch.json file, since it is Metadata of the sketch.
			continue
		}
		_, filePath := filepath.Split(f.Path)
		content, err := ioutil.ReadFile(filepath.Join(sketch.FullPath, filePath))
		if err != nil {
			return nil
		}

		data := base64.StdEncoding.EncodeToString(content)
		ret.Files = append(ret.Files, &File{
			Data: &data,
			Name: f.Name,
		})
	}
	return &ret
}

// SketchMetadata user type.
type SketchMetadata struct {
	CPU          *SketchMetadataCPU   `form:"cpu,omitempty" json:"cpu,omitempty" xml:"cpu,omitempty"`
	IncludedLibs []*SketchMetadataLib `form:"included_libs,omitempty" json:"included_libs,omitempty" xml:"included_libs,omitempty"`
}

// ConvertMetadataFrom creates SketchMetadata object from sketches.Metadata.
func ConvertMetadataFrom(metadata *sketches.Metadata) *SketchMetadata {
	if metadata == nil {
		return nil
	}
	network := metadata.CPU.Type == "network"
	ret := SketchMetadata{
		CPU: &SketchMetadataCPU{
			Fqbn:    &metadata.CPU.Fqbn,
			Name:    &metadata.CPU.Name,
			Network: &network,
			Port:    &metadata.CPU.Port,
		},
	}
	ret.IncludedLibs = make([]*SketchMetadataLib, len(metadata.IncludedLibs))
	for i, lib := range metadata.IncludedLibs {
		ret.IncludedLibs[i] = &SketchMetadataLib{
			Name:    &lib.Name,
			Version: &lib.Version,
		}
	}

	return &ret
}

// SketchMetadataCPU is the board associated with the sketch.
type SketchMetadataCPU struct {
	// The fqbn of the board.
	Fqbn *string `form:"fqbn,omitempty" json:"fqbn,omitempty" xml:"fqbn,omitempty"`
	// The name of the board.
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// Requires an upload via network.
	Network *bool `form:"network,omitempty" json:"network,omitempty" xml:"network,omitempty"`
	// The port of the board.
	Port *string `form:"port,omitempty" json:"port,omitempty" xml:"port,omitempty"`
}

// SketchMetadataLib is a library associated with the sketch.
type SketchMetadataLib struct {
	// The name of the library.
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// The version of the library.
	Version *string `form:"version,omitempty" json:"version,omitempty" xml:"version,omitempty"`
}
