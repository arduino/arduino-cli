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
	uuid "github.com/goadesign/goa/uuid"
	"net/http"
	"time"
)

// A file saved on the virtual filesystem (default view)
//
// Identifier: application/vnd.arduino.create.file+json; view=default
type ArduinoCreateFile struct {
	// The contents of the file, encoded in base64
	Data *string `form:"data,omitempty" json:"data,omitempty" xml:"data,omitempty"`
	// The url to use to load the file
	Href *string `form:"href,omitempty" json:"href,omitempty" xml:"href,omitempty"`
	// The mimetype of the file, from its extension
	Mimetype *string `form:"mimetype,omitempty" json:"mimetype,omitempty" xml:"mimetype,omitempty"`
	// The name of the file
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// The path of the file
	Path *string `form:"path,omitempty" json:"path,omitempty" xml:"path,omitempty"`
	// The size in bytes
	Size *int `form:"size,omitempty" json:"size,omitempty" xml:"size,omitempty"`
}

// DecodeArduinoCreateFile decodes the ArduinoCreateFile instance encoded in resp body.
func (c *Client) DecodeArduinoCreateFile(resp *http.Response) (*ArduinoCreateFile, error) {
	var decoded ArduinoCreateFile
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// A paginated list of libraries (default view)
//
// Identifier: application/vnd.arduino.create.libraries+json; view=default
type ArduinoCreateLibraries struct {
	// The list of libraries
	Libraries []*ArduinoCreateLibrary `form:"libraries" json:"libraries" xml:"libraries"`
	// Link to the following page of results. Could be empty.
	Next *string `form:"next,omitempty" json:"next,omitempty" xml:"next,omitempty"`
	// Link to the previous page of results. Could be empty.
	Prev *string `form:"prev,omitempty" json:"prev,omitempty" xml:"prev,omitempty"`
}

// Validate validates the ArduinoCreateLibraries media type instance.
func (mt *ArduinoCreateLibraries) Validate() (err error) {
	if mt.Libraries == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "libraries"))
	}
	return
}

// DecodeArduinoCreateLibraries decodes the ArduinoCreateLibraries instance encoded in resp body.
func (c *Client) DecodeArduinoCreateLibraries(resp *http.Response) (*ArduinoCreateLibraries, error) {
	var decoded ArduinoCreateLibraries
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// Library is a collection of header files containing arduino reusable code and functions. It typically contains its info in a library.properties files. The examples property contains a list of examples that use that library. (default view)
//
// Identifier: application/vnd.arduino.create.library+json; view=default
type ArduinoCreateLibrary struct {
	// The architectures supported by the library.
	Architectures []string `form:"architectures,omitempty" json:"architectures,omitempty" xml:"architectures,omitempty"`
	// A category
	Category *string `form:"category,omitempty" json:"category,omitempty" xml:"category,omitempty"`
	// A snippet of code that includes all of the library header files
	Code    *string    `form:"code,omitempty" json:"code,omitempty" xml:"code,omitempty"`
	Created *time.Time `form:"created,omitempty" json:"created,omitempty" xml:"created,omitempty"`
	// The examples contained in the library
	Examples []*ArduinoCreateSketch `form:"examples,omitempty" json:"examples,omitempty" xml:"examples,omitempty"`
	// The number of examples that it contains
	ExamplesNumber *int `form:"examples_number,omitempty" json:"examples_number,omitempty" xml:"examples_number,omitempty"`
	// The files contained in the library
	Files []*ArduinoCreateFile `form:"files,omitempty" json:"files,omitempty" xml:"files,omitempty"`
	// The url where to find the details
	Href *string `form:"href,omitempty" json:"href,omitempty" xml:"href,omitempty"`
	// The id of the library
	ID *uuid.UUID `form:"id,omitempty" json:"id,omitempty" xml:"id,omitempty"`
	// The maintainer of the library
	Maintainer *string    `form:"maintainer,omitempty" json:"maintainer,omitempty" xml:"maintainer,omitempty"`
	Modified   *time.Time `form:"modified,omitempty" json:"modified,omitempty" xml:"modified,omitempty"`
	// The name of the library, shared between many versions
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
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

// Library is a collection of header files containing arduino reusable code and functions. It typically contains its info in a library.properties files. The examples property contains a list of examples that use that library. (link view)
//
// Identifier: application/vnd.arduino.create.library+json; view=link
type ArduinoCreateLibraryLink struct {
	// The url where to find the details
	Href *string `form:"href,omitempty" json:"href,omitempty" xml:"href,omitempty"`
}

// DecodeArduinoCreateLibrary decodes the ArduinoCreateLibrary instance encoded in resp body.
func (c *Client) DecodeArduinoCreateLibrary(resp *http.Response) (*ArduinoCreateLibrary, error) {
	var decoded ArduinoCreateLibrary
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// DecodeArduinoCreateLibraryLink decodes the ArduinoCreateLibraryLink instance encoded in resp body.
func (c *Client) DecodeArduinoCreateLibraryLink(resp *http.Response) (*ArduinoCreateLibraryLink, error) {
	var decoded ArduinoCreateLibraryLink
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// A program meant to be uploaded onto a board (default view)
//
// Identifier: application/vnd.arduino.create.sketch+json; view=default
type ArduinoCreateSketch struct {
	Created *time.Time `form:"created,omitempty" json:"created,omitempty" xml:"created,omitempty"`
	// The other files contained in the sketch
	Files []*ArduinoCreateFile `form:"files,omitempty" json:"files,omitempty" xml:"files,omitempty"`
	// The folder path where the sketch is saved
	Folder *string `form:"folder,omitempty" json:"folder,omitempty" xml:"folder,omitempty"`
	// The url to use to modify or delete the sketch
	Href *string `form:"href,omitempty" json:"href,omitempty" xml:"href,omitempty"`
	// The id of the sketch
	ID *uuid.UUID `form:"id,omitempty" json:"id,omitempty" xml:"id,omitempty"`
	// The main file of the sketch
	Ino      *ArduinoCreateFile           `form:"ino,omitempty" json:"ino,omitempty" xml:"ino,omitempty"`
	Metadata *ArduinoCreateSketchMetadata `form:"metadata,omitempty" json:"metadata,omitempty" xml:"metadata,omitempty"`
	Modified *time.Time                   `form:"modified,omitempty" json:"modified,omitempty" xml:"modified,omitempty"`
	// The name of the sketch
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// The username of the owner of the sketch
	Owner *string `form:"owner,omitempty" json:"owner,omitempty" xml:"owner,omitempty"`
	// A private sketch is only visible to its owner.
	Private bool `form:"private" json:"private" xml:"private"`
	// The total size of the sketch. Only available if the sketch is fully loaded.
	Size *int `form:"size,omitempty" json:"size,omitempty" xml:"size,omitempty"`
	// A list of links to hackster tutorials.
	Tutorials []string `form:"tutorials,omitempty" json:"tutorials,omitempty" xml:"tutorials,omitempty"`
	// A list of tags. The builtin tag means that it's a builtin example.
	Types []string `form:"types,omitempty" json:"types,omitempty" xml:"types,omitempty"`
}

// DecodeArduinoCreateSketch decodes the ArduinoCreateSketch instance encoded in resp body.
func (c *Client) DecodeArduinoCreateSketch(resp *http.Response) (*ArduinoCreateSketch, error) {
	var decoded ArduinoCreateSketch
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// ArduinoCreateSketchMetadata media type (default view)
//
// Identifier: application/vnd.arduino.create.sketch.metadata+json; view=default
type ArduinoCreateSketchMetadata struct {
	CPU          *ArduinoCreateSketchMetadataCPU           `form:"cpu,omitempty" json:"cpu,omitempty" xml:"cpu,omitempty"`
	IncludedLibs []*ArduinoCreateSketchMetadataIncludedLib `form:"included_libs,omitempty" json:"included_libs,omitempty" xml:"included_libs,omitempty"`
}

// DecodeArduinoCreateSketchMetadata decodes the ArduinoCreateSketchMetadata instance encoded in resp body.
func (c *Client) DecodeArduinoCreateSketchMetadata(resp *http.Response) (*ArduinoCreateSketchMetadata, error) {
	var decoded ArduinoCreateSketchMetadata
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// ArduinoCreateSketchMetadataCpu media type (default view)
//
// Identifier: application/vnd.arduino.create.sketch.metadata.cpu+json; view=default
type ArduinoCreateSketchMetadataCPU struct {
	// The fqbn of the board
	Fqbn *string `form:"fqbn,omitempty" json:"fqbn,omitempty" xml:"fqbn,omitempty"`
	// The name of the board
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// Requires an upload via network
	Network *bool `form:"network,omitempty" json:"network,omitempty" xml:"network,omitempty"`
	// The port of the board
	Port *string `form:"port,omitempty" json:"port,omitempty" xml:"port,omitempty"`
}

// DecodeArduinoCreateSketchMetadataCPU decodes the ArduinoCreateSketchMetadataCPU instance encoded in resp body.
func (c *Client) DecodeArduinoCreateSketchMetadataCPU(resp *http.Response) (*ArduinoCreateSketchMetadataCPU, error) {
	var decoded ArduinoCreateSketchMetadataCPU
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// ArduinoCreateSketchMetadataIncluded_lib media type (default view)
//
// Identifier: application/vnd.arduino.create.sketch.metadata.included_lib+json; view=default
type ArduinoCreateSketchMetadataIncludedLib struct {
	// The name of the library
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// The version of the library
	Version *string `form:"version,omitempty" json:"version,omitempty" xml:"version,omitempty"`
}

// DecodeArduinoCreateSketchMetadataIncludedLib decodes the ArduinoCreateSketchMetadataIncludedLib instance encoded in resp body.
func (c *Client) DecodeArduinoCreateSketchMetadataIncludedLib(resp *http.Response) (*ArduinoCreateSketchMetadataIncludedLib, error) {
	var decoded ArduinoCreateSketchMetadataIncludedLib
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// A paginated list of sketches (default view)
//
// Identifier: application/vnd.arduino.create.sketches+json; view=default
type ArduinoCreateSketches struct {
	// Link to the following page of results. Could be empty.
	Next *string `form:"next,omitempty" json:"next,omitempty" xml:"next,omitempty"`
	// Link to the previous page of results. Could be empty.
	Prev *string `form:"prev,omitempty" json:"prev,omitempty" xml:"prev,omitempty"`
	// The list of sketches
	Sketches []*ArduinoCreateSketch `form:"sketches" json:"sketches" xml:"sketches"`
}

// Validate validates the ArduinoCreateSketches media type instance.
func (mt *ArduinoCreateSketches) Validate() (err error) {
	if mt.Sketches == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "sketches"))
	}
	return
}

// DecodeArduinoCreateSketches decodes the ArduinoCreateSketches instance encoded in resp body.
func (c *Client) DecodeArduinoCreateSketches(resp *http.Response) (*ArduinoCreateSketches, error) {
	var decoded ArduinoCreateSketches
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// A user registered with the create service (default view)
//
// Identifier: application/vnd.arduino.create.user+json; view=default
type ArduinoCreateUser struct {
	// When the user was activated in the create service
	Activated *time.Time `form:"activated,omitempty" json:"activated,omitempty" xml:"activated,omitempty"`
	// When the user was created in the create service
	Created *time.Time `form:"created,omitempty" json:"created,omitempty" xml:"created,omitempty"`
	Email   *string    `form:"email,omitempty" json:"email,omitempty" xml:"email,omitempty"`
	// The id of the user
	ID     *string                  `form:"id,omitempty" json:"id,omitempty" xml:"id,omitempty"`
	Limits *ArduinoCreateUserLimits `form:"limits,omitempty" json:"limits,omitempty" xml:"limits,omitempty"`
	Prefs  *ArduinoCreateUserPrefs  `form:"prefs,omitempty" json:"prefs,omitempty" xml:"prefs,omitempty"`
}

// DecodeArduinoCreateUser decodes the ArduinoCreateUser instance encoded in resp body.
func (c *Client) DecodeArduinoCreateUser(resp *http.Response) (*ArduinoCreateUser, error) {
	var decoded ArduinoCreateUser
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// The limits of the user (default view)
//
// Identifier: application/vnd.arduino.create.user.limits+json; view=default
type ArduinoCreateUserLimits struct {
	// The maximum number of compilations in 24 hours
	Compilations *int `form:"compilations,omitempty" json:"compilations,omitempty" xml:"compilations,omitempty"`
	// The total space available to the user, in kb
	Disk *int `form:"disk,omitempty" json:"disk,omitempty" xml:"disk,omitempty"`
	// The maximum number of sketches they can have
	Sketches *int `form:"sketches,omitempty" json:"sketches,omitempty" xml:"sketches,omitempty"`
}

// DecodeArduinoCreateUserLimits decodes the ArduinoCreateUserLimits instance encoded in resp body.
func (c *Client) DecodeArduinoCreateUserLimits(resp *http.Response) (*ArduinoCreateUserLimits, error) {
	var decoded ArduinoCreateUserLimits
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// The user preferences about create (default view)
//
// Identifier: application/vnd.arduino.create.user.prefs+json; view=default
type ArduinoCreateUserPrefs struct {
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

// DecodeArduinoCreateUserPrefs decodes the ArduinoCreateUserPrefs instance encoded in resp body.
func (c *Client) DecodeArduinoCreateUserPrefs(resp *http.Response) (*ArduinoCreateUserPrefs, error) {
	var decoded ArduinoCreateUserPrefs
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// The stats about the user (default view)
//
// Identifier: application/vnd.arduino.create.user.stats+json; view=default
type ArduinoCreateUserStats struct {
	// The number of compilations made in the last 24 hours
	Compilations *int `form:"compilations,omitempty" json:"compilations,omitempty" xml:"compilations,omitempty"`
	// The space used by the user's libraries, in kb
	DiskLibraries *int `form:"disk_libraries,omitempty" json:"disk_libraries,omitempty" xml:"disk_libraries,omitempty"`
	// The space used by the user's sketches, in kb
	DiskSketches *int `form:"disk_sketches,omitempty" json:"disk_sketches,omitempty" xml:"disk_sketches,omitempty"`
	// The number of libraries they own
	Libraries *int `form:"libraries,omitempty" json:"libraries,omitempty" xml:"libraries,omitempty"`
	// The number of sketches they own
	Sketches *int `form:"sketches,omitempty" json:"sketches,omitempty" xml:"sketches,omitempty"`
}

// DecodeArduinoCreateUserStats decodes the ArduinoCreateUserStats instance encoded in resp body.
func (c *Client) DecodeArduinoCreateUserStats(resp *http.Response) (*ArduinoCreateUserStats, error) {
	var decoded ArduinoCreateUserStats
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// DecodeErrorResponse decodes the ErrorResponse instance encoded in resp body.
func (c *Client) DecodeErrorResponse(resp *http.Response) (*goa.ErrorResponse, error) {
	var decoded goa.ErrorResponse
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}
