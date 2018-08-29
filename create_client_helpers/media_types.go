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
	"net/http"
	"time"

	"github.com/goadesign/goa"
	uuid "github.com/goadesign/goa/uuid"
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

// DecodeErrorResponse decodes the ErrorResponse instance encoded in resp body.
func (c *Client) DecodeErrorResponse(resp *http.Response) (*goa.ErrorResponse, error) {
	var decoded goa.ErrorResponse
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}
