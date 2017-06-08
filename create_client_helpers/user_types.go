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

package createClient

import (
	"github.com/goadesign/goa"
)

// A file saved on the virtual filesystem
type file struct {
	// The contents of the file, encoded in base64
	Data *string `form:"data,omitempty" json:"data,omitempty" xml:"data,omitempty"`
	// The name of the file
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
}

// Validate validates the file type instance.
func (ut *file) Validate() (err error) {
	if ut.Name == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "name"))
	}
	return
}

// Publicize creates File from file
func (ut *file) Publicize() *File {
	var pub File
	if ut.Data != nil {
		pub.Data = ut.Data
	}
	if ut.Name != nil {
		pub.Name = *ut.Name
	}
	return &pub
}

// A file saved on the virtual filesystem
type File struct {
	// The contents of the file, encoded in base64
	Data *string `form:"data,omitempty" json:"data,omitempty" xml:"data,omitempty"`
	// The name of the file
	Name string `form:"name" json:"name" xml:"name"`
}

// Validate validates the File type instance.
func (ut *File) Validate() (err error) {
	if ut.Name == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "name"))
	}
	return
}

// Library is a collection of header files containing arduino reusable code and functions. It typically contains its info in a library.properties files. The examples property contains a list of examples that use that library.
type library struct {
	// The architectures supported by the library.
	Architectures []string `form:"architectures,omitempty" json:"architectures,omitempty" xml:"architectures,omitempty"`
	// A category
	Category *string `form:"category,omitempty" json:"category,omitempty" xml:"category,omitempty"`
	// A snippet of code that includes all of the library header files
	Code *string `form:"code,omitempty" json:"code,omitempty" xml:"code,omitempty"`
	// The examples contained in the library
	Examples []*sketch `form:"examples,omitempty" json:"examples,omitempty" xml:"examples,omitempty"`
	// The files contained in the library
	Files []*file `form:"files,omitempty" json:"files,omitempty" xml:"files,omitempty"`
	// The maintainer of the library
	Maintainer *string `form:"maintainer,omitempty" json:"maintainer,omitempty" xml:"maintainer,omitempty"`
	// The name of the library, shared between many versions
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// A private sketch is only visible to its owner.
	Private *bool `form:"private,omitempty" json:"private,omitempty" xml:"private,omitempty"`
	// A short description
	Sentence *string `form:"sentence,omitempty" json:"sentence,omitempty" xml:"sentence,omitempty"`
	// A list of tags. The Arduino tag means that it's a builtin library.
	Types []string `form:"types,omitempty" json:"types,omitempty" xml:"types,omitempty"`
	// The homepage of the library
	URL *string `form:"url,omitempty" json:"url,omitempty" xml:"url,omitempty"`
	// The version of the library
	Version *string `form:"version,omitempty" json:"version,omitempty" xml:"version,omitempty"`
}

// Finalize sets the default values for library type instance.
func (ut *library) Finalize() {
	for _, e := range ut.Examples {
		var defaultPrivate = false
		if e.Private == nil {
			e.Private = &defaultPrivate
		}
	}
	var defaultPrivate = false
	if ut.Private == nil {
		ut.Private = &defaultPrivate
	}
}

// Validate validates the library type instance.
func (ut *library) Validate() (err error) {
	if ut.Name == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "name"))
	}
	if ut.Private == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "private"))
	}
	for _, e := range ut.Examples {
		if e != nil {
			if err2 := e.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	}
	for _, e := range ut.Files {
		if e != nil {
			if err2 := e.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	}
	return
}

// Publicize creates Library from library
func (ut *library) Publicize() *Library {
	var pub Library
	if ut.Architectures != nil {
		pub.Architectures = ut.Architectures
	}
	if ut.Category != nil {
		pub.Category = ut.Category
	}
	if ut.Code != nil {
		pub.Code = ut.Code
	}
	if ut.Examples != nil {
		pub.Examples = make([]*Sketch, len(ut.Examples))
		for i2, elem2 := range ut.Examples {
			pub.Examples[i2] = elem2.Publicize()
		}
	}
	if ut.Files != nil {
		pub.Files = make([]*File, len(ut.Files))
		for i2, elem2 := range ut.Files {
			pub.Files[i2] = elem2.Publicize()
		}
	}
	if ut.Maintainer != nil {
		pub.Maintainer = ut.Maintainer
	}
	if ut.Name != nil {
		pub.Name = *ut.Name
	}
	if ut.Private != nil {
		pub.Private = *ut.Private
	}
	if ut.Sentence != nil {
		pub.Sentence = ut.Sentence
	}
	if ut.Types != nil {
		pub.Types = ut.Types
	}
	if ut.URL != nil {
		pub.URL = ut.URL
	}
	if ut.Version != nil {
		pub.Version = ut.Version
	}
	return &pub
}

// Library is a collection of header files containing arduino reusable code and functions. It typically contains its info in a library.properties files. The examples property contains a list of examples that use that library.
type Library struct {
	// The architectures supported by the library.
	Architectures []string `form:"architectures,omitempty" json:"architectures,omitempty" xml:"architectures,omitempty"`
	// A category
	Category *string `form:"category,omitempty" json:"category,omitempty" xml:"category,omitempty"`
	// A snippet of code that includes all of the library header files
	Code *string `form:"code,omitempty" json:"code,omitempty" xml:"code,omitempty"`
	// The examples contained in the library
	Examples []*Sketch `form:"examples,omitempty" json:"examples,omitempty" xml:"examples,omitempty"`
	// The files contained in the library
	Files []*File `form:"files,omitempty" json:"files,omitempty" xml:"files,omitempty"`
	// The maintainer of the library
	Maintainer *string `form:"maintainer,omitempty" json:"maintainer,omitempty" xml:"maintainer,omitempty"`
	// The name of the library, shared between many versions
	Name string `form:"name" json:"name" xml:"name"`
	// A private sketch is only visible to its owner.
	Private bool `form:"private" json:"private" xml:"private"`
	// A short description
	Sentence *string `form:"sentence,omitempty" json:"sentence,omitempty" xml:"sentence,omitempty"`
	// A list of tags. The Arduino tag means that it's a builtin library.
	Types []string `form:"types,omitempty" json:"types,omitempty" xml:"types,omitempty"`
	// The homepage of the library
	URL *string `form:"url,omitempty" json:"url,omitempty" xml:"url,omitempty"`
	// The version of the library
	Version *string `form:"version,omitempty" json:"version,omitempty" xml:"version,omitempty"`
}

// Validate validates the Library type instance.
func (ut *Library) Validate() (err error) {
	if ut.Name == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "name"))
	}

	for _, e := range ut.Examples {
		if e != nil {
			if err2 := e.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	}
	for _, e := range ut.Files {
		if e != nil {
			if err2 := e.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	}
	return
}

// A program meant to be uploaded onto a board
type sketch struct {
	// The other files contained in the sketch
	Files []*file `form:"files,omitempty" json:"files,omitempty" xml:"files,omitempty"`
	// The folder path where the sketch is saved
	Folder *string `form:"folder,omitempty" json:"folder,omitempty" xml:"folder,omitempty"`
	// The main file of the sketch
	Ino      *file           `form:"ino,omitempty" json:"ino,omitempty" xml:"ino,omitempty"`
	Metadata *sketchMetadata `form:"metadata,omitempty" json:"metadata,omitempty" xml:"metadata,omitempty"`
	// The name of the sketch
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// The username of the owner of the sketch
	Owner *string `form:"owner,omitempty" json:"owner,omitempty" xml:"owner,omitempty"`
	// A private sketch is only visible to its owner.
	Private *bool `form:"private,omitempty" json:"private,omitempty" xml:"private,omitempty"`
	// A list of links to hackster tutorials.
	Tutorials []string `form:"tutorials,omitempty" json:"tutorials,omitempty" xml:"tutorials,omitempty"`
	// A list of tags. The builtin tag means that it's a builtin example.
	Types []string `form:"types,omitempty" json:"types,omitempty" xml:"types,omitempty"`
}

// Finalize sets the default values for sketch type instance.
func (ut *sketch) Finalize() {
	var defaultPrivate = false
	if ut.Private == nil {
		ut.Private = &defaultPrivate
	}
}

// Validate validates the sketch type instance.
func (ut *sketch) Validate() (err error) {
	if ut.Name == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "name"))
	}
	if ut.Ino == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "ino"))
	}
	for _, e := range ut.Files {
		if e != nil {
			if err2 := e.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	}
	if ut.Ino != nil {
		if err2 := ut.Ino.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// Publicize creates Sketch from sketch
func (ut *sketch) Publicize() *Sketch {
	var pub Sketch
	if ut.Files != nil {
		pub.Files = make([]*File, len(ut.Files))
		for i2, elem2 := range ut.Files {
			pub.Files[i2] = elem2.Publicize()
		}
	}
	if ut.Folder != nil {
		pub.Folder = ut.Folder
	}
	if ut.Ino != nil {
		pub.Ino = ut.Ino.Publicize()
	}
	if ut.Metadata != nil {
		pub.Metadata = ut.Metadata.Publicize()
	}
	if ut.Name != nil {
		pub.Name = *ut.Name
	}
	if ut.Owner != nil {
		pub.Owner = ut.Owner
	}
	if ut.Private != nil {
		pub.Private = *ut.Private
	}
	if ut.Tutorials != nil {
		pub.Tutorials = ut.Tutorials
	}
	if ut.Types != nil {
		pub.Types = ut.Types
	}
	return &pub
}

// A program meant to be uploaded onto a board
type Sketch struct {
	// The other files contained in the sketch
	Files []*File `form:"files,omitempty" json:"files,omitempty" xml:"files,omitempty"`
	// The folder path where the sketch is saved
	Folder *string `form:"folder,omitempty" json:"folder,omitempty" xml:"folder,omitempty"`
	// The main file of the sketch
	Ino      *File           `form:"ino" json:"ino" xml:"ino"`
	Metadata *SketchMetadata `form:"metadata,omitempty" json:"metadata,omitempty" xml:"metadata,omitempty"`
	// The name of the sketch
	Name string `form:"name" json:"name" xml:"name"`
	// The username of the owner of the sketch
	Owner *string `form:"owner,omitempty" json:"owner,omitempty" xml:"owner,omitempty"`
	// A private sketch is only visible to its owner.
	Private bool `form:"private" json:"private" xml:"private"`
	// A list of links to hackster tutorials.
	Tutorials []string `form:"tutorials,omitempty" json:"tutorials,omitempty" xml:"tutorials,omitempty"`
	// A list of tags. The builtin tag means that it's a builtin example.
	Types []string `form:"types,omitempty" json:"types,omitempty" xml:"types,omitempty"`
}

// Validate validates the Sketch type instance.
func (ut *Sketch) Validate() (err error) {
	if ut.Name == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "name"))
	}
	if ut.Ino == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "ino"))
	}
	for _, e := range ut.Files {
		if e != nil {
			if err2 := e.Validate(); err2 != nil {
				err = goa.MergeErrors(err, err2)
			}
		}
	}
	if ut.Ino != nil {
		if err2 := ut.Ino.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// sketchMetadata user type.
type sketchMetadata struct {
	CPU          *sketchMetadataCPU   `form:"cpu,omitempty" json:"cpu,omitempty" xml:"cpu,omitempty"`
	IncludedLibs []*sketchMetadataLib `form:"included_libs,omitempty" json:"included_libs,omitempty" xml:"included_libs,omitempty"`
}

// Publicize creates SketchMetadata from sketchMetadata
func (ut *sketchMetadata) Publicize() *SketchMetadata {
	var pub SketchMetadata
	if ut.CPU != nil {
		pub.CPU = ut.CPU.Publicize()
	}
	if ut.IncludedLibs != nil {
		pub.IncludedLibs = make([]*SketchMetadataLib, len(ut.IncludedLibs))
		for i2, elem2 := range ut.IncludedLibs {
			pub.IncludedLibs[i2] = elem2.Publicize()
		}
	}
	return &pub
}

// SketchMetadata user type.
type SketchMetadata struct {
	CPU          *SketchMetadataCPU   `form:"cpu,omitempty" json:"cpu,omitempty" xml:"cpu,omitempty"`
	IncludedLibs []*SketchMetadataLib `form:"included_libs,omitempty" json:"included_libs,omitempty" xml:"included_libs,omitempty"`
}

// The board associated with the sketch
type sketchMetadataCPU struct {
	// The fqbn of the board
	Fqbn *string `form:"fqbn,omitempty" json:"fqbn,omitempty" xml:"fqbn,omitempty"`
	// The name of the board
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// Requires an upload via network
	Network *bool `form:"network,omitempty" json:"network,omitempty" xml:"network,omitempty"`
	// The port of the board
	Port *string `form:"port,omitempty" json:"port,omitempty" xml:"port,omitempty"`
}

// Publicize creates SketchMetadataCPU from sketchMetadataCPU
func (ut *sketchMetadataCPU) Publicize() *SketchMetadataCPU {
	var pub SketchMetadataCPU
	if ut.Fqbn != nil {
		pub.Fqbn = ut.Fqbn
	}
	if ut.Name != nil {
		pub.Name = ut.Name
	}
	if ut.Network != nil {
		pub.Network = ut.Network
	}
	if ut.Port != nil {
		pub.Port = ut.Port
	}
	return &pub
}

// The board associated with the sketch
type SketchMetadataCPU struct {
	// The fqbn of the board
	Fqbn *string `form:"fqbn,omitempty" json:"fqbn,omitempty" xml:"fqbn,omitempty"`
	// The name of the board
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// Requires an upload via network
	Network *bool `form:"network,omitempty" json:"network,omitempty" xml:"network,omitempty"`
	// The port of the board
	Port *string `form:"port,omitempty" json:"port,omitempty" xml:"port,omitempty"`
}

// A library associated with the sketch
type sketchMetadataLib struct {
	// The name of the library
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// The version of the library
	Version *string `form:"version,omitempty" json:"version,omitempty" xml:"version,omitempty"`
}

// Publicize creates SketchMetadataLib from sketchMetadataLib
func (ut *sketchMetadataLib) Publicize() *SketchMetadataLib {
	var pub SketchMetadataLib
	if ut.Name != nil {
		pub.Name = ut.Name
	}
	if ut.Version != nil {
		pub.Version = ut.Version
	}
	return &pub
}

// A library associated with the sketch
type SketchMetadataLib struct {
	// The name of the library
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// The version of the library
	Version *string `form:"version,omitempty" json:"version,omitempty" xml:"version,omitempty"`
}

// The fields that can be edited about a user
type user struct {
	// Sets the activated field to the current time
	ActivateNow *bool `form:"activate_now,omitempty" json:"activate_now,omitempty" xml:"activate_now,omitempty"`
	// Set the email of the user at the first save
	Email *string    `form:"email,omitempty" json:"email,omitempty" xml:"email,omitempty"`
	Prefs *userPrefs `form:"prefs,omitempty" json:"prefs,omitempty" xml:"prefs,omitempty"`
}

// Finalize sets the default values for user type instance.
func (ut *user) Finalize() {
	var defaultActivateNow = false
	if ut.ActivateNow == nil {
		ut.ActivateNow = &defaultActivateNow
	}
}

// Publicize creates User from user
func (ut *user) Publicize() *User {
	var pub User
	if ut.ActivateNow != nil {
		pub.ActivateNow = *ut.ActivateNow
	}
	if ut.Email != nil {
		pub.Email = ut.Email
	}
	if ut.Prefs != nil {
		pub.Prefs = ut.Prefs.Publicize()
	}
	return &pub
}

// The fields that can be edited about a user
type User struct {
	// Sets the activated field to the current time
	ActivateNow bool `form:"activate_now" json:"activate_now" xml:"activate_now"`
	// Set the email of the user at the first save
	Email *string    `form:"email,omitempty" json:"email,omitempty" xml:"email,omitempty"`
	Prefs *UserPrefs `form:"prefs,omitempty" json:"prefs,omitempty" xml:"prefs,omitempty"`
}

// The user preferences about create
type userPrefs struct {
	// Save every few seconds
	Autosave *bool `form:"autosave,omitempty" json:"autosave,omitempty" xml:"autosave,omitempty"`
	// The size of the text
	FontSize *int `form:"font_size,omitempty" json:"font_size,omitempty" xml:"font_size,omitempty"`
	// Hide the verbose panel
	HidePanel *bool `form:"hide_panel,omitempty" json:"hide_panel,omitempty" xml:"hide_panel,omitempty"`
	// Save when compiling
	SaveOnBuild *bool `form:"save_on_build,omitempty" json:"save_on_build,omitempty" xml:"save_on_build,omitempty"`
	// Show all files
	ShowAllContent *bool `form:"show_all_content,omitempty" json:"show_all_content,omitempty" xml:"show_all_content,omitempty"`
	// The editor theme
	Skin *string `form:"skin,omitempty" json:"skin,omitempty" xml:"skin,omitempty"`
	// Show verbose output
	Verbose *bool `form:"verbose,omitempty" json:"verbose,omitempty" xml:"verbose,omitempty"`
	// DEPRECATED. Use hide_panel instead
	VerboseAlwaysVisible *bool `form:"verbose_always_visible,omitempty" json:"verbose_always_visible,omitempty" xml:"verbose_always_visible,omitempty"`
}

// Publicize creates UserPrefs from userPrefs
func (ut *userPrefs) Publicize() *UserPrefs {
	var pub UserPrefs
	if ut.Autosave != nil {
		pub.Autosave = ut.Autosave
	}
	if ut.FontSize != nil {
		pub.FontSize = ut.FontSize
	}
	if ut.HidePanel != nil {
		pub.HidePanel = ut.HidePanel
	}
	if ut.SaveOnBuild != nil {
		pub.SaveOnBuild = ut.SaveOnBuild
	}
	if ut.ShowAllContent != nil {
		pub.ShowAllContent = ut.ShowAllContent
	}
	if ut.Skin != nil {
		pub.Skin = ut.Skin
	}
	if ut.Verbose != nil {
		pub.Verbose = ut.Verbose
	}
	if ut.VerboseAlwaysVisible != nil {
		pub.VerboseAlwaysVisible = ut.VerboseAlwaysVisible
	}
	return &pub
}

// The user preferences about create
type UserPrefs struct {
	// Save every few seconds
	Autosave *bool `form:"autosave,omitempty" json:"autosave,omitempty" xml:"autosave,omitempty"`
	// The size of the text
	FontSize *int `form:"font_size,omitempty" json:"font_size,omitempty" xml:"font_size,omitempty"`
	// Hide the verbose panel
	HidePanel *bool `form:"hide_panel,omitempty" json:"hide_panel,omitempty" xml:"hide_panel,omitempty"`
	// Save when compiling
	SaveOnBuild *bool `form:"save_on_build,omitempty" json:"save_on_build,omitempty" xml:"save_on_build,omitempty"`
	// Show all files
	ShowAllContent *bool `form:"show_all_content,omitempty" json:"show_all_content,omitempty" xml:"show_all_content,omitempty"`
	// The editor theme
	Skin *string `form:"skin,omitempty" json:"skin,omitempty" xml:"skin,omitempty"`
	// Show verbose output
	Verbose *bool `form:"verbose,omitempty" json:"verbose,omitempty" xml:"verbose,omitempty"`
	// DEPRECATED. Use hide_panel instead
	VerboseAlwaysVisible *bool `form:"verbose_always_visible,omitempty" json:"verbose_always_visible,omitempty" xml:"verbose_always_visible,omitempty"`
}
