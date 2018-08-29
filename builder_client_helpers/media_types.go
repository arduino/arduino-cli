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

package builderclient

import (
	"net/http"
	"time"

	"github.com/goadesign/goa"
)

// ArduinoBuilderBoard is a physical board belonging to a certain architecture in a package. The most obvious package is arduino, which contains architectures avr, sam and samd. It can contain multiple versions of the upload commands and options. If there is a default version it means that it's the only version officially supported. Of course if there is only one version it will be called default (default view)
//
// Identifier: application/vnd.arduino.builder.board+json; view=default
type ArduinoBuilderBoard struct {
	// The architecture of the board
	Architecture *string                          `form:"architecture,omitempty" json:"architecture,omitempty" xml:"architecture,omitempty"`
	Bootloader   []*ArduinoBuilderBoardBootloader `form:"bootloader,omitempty" json:"bootloader,omitempty" xml:"bootloader,omitempty"`
	Build        []*ArduinoBuilderBoardBuild      `form:"build,omitempty" json:"build,omitempty" xml:"build,omitempty"`
	// The default flavour of the board
	DefaultFlavour *string `form:"default_flavour,omitempty" json:"default_flavour,omitempty" xml:"default_flavour,omitempty"`
	// An identifier used by the tools to determine which tools to use on it
	Fqbn *string `form:"fqbn,omitempty" json:"fqbn,omitempty" xml:"fqbn,omitempty"`
	// The id of the board
	ID *string `form:"id,omitempty" json:"id,omitempty" xml:"id,omitempty"`
	// The name of the board
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// The package to which the board belongs
	Package *string `form:"package,omitempty" json:"package,omitempty" xml:"package,omitempty"`
	// A list of possible pids
	Pid    []string                     `form:"pid,omitempty" json:"pid,omitempty" xml:"pid,omitempty"`
	Upload []*ArduinoBuilderBoardUpload `form:"upload,omitempty" json:"upload,omitempty" xml:"upload,omitempty"`
	// A list of possible vids
	Vid []string `form:"vid,omitempty" json:"vid,omitempty" xml:"vid,omitempty"`
}

// DecodeArduinoBuilderBoard decodes the ArduinoBuilderBoard instance encoded in resp body.
func (c *Client) DecodeArduinoBuilderBoard(resp *http.Response) (*ArduinoBuilderBoard, error) {
	var decoded ArduinoBuilderBoard
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// ArduinoBuilderBoardBootloader contains the info used to bootload a board. (default view)
//
// Identifier: application/vnd.arduino.builder.board.bootloader; view=default
type ArduinoBuilderBoardBootloader struct {
	// The commandline used to bootload
	Commandline *string `form:"commandline,omitempty" json:"commandline,omitempty" xml:"commandline,omitempty"`
	// The flavour of the board. Usually it's default
	Flavour *string `form:"flavour,omitempty" json:"flavour,omitempty" xml:"flavour,omitempty"`
	// The signature of the commandline
	Signature *string `form:"signature,omitempty" json:"signature,omitempty" xml:"signature,omitempty"`
}

// DecodeArduinoBuilderBoardBootloader decodes the ArduinoBuilderBoardBootloader instance encoded in resp body.
func (c *Client) DecodeArduinoBuilderBoardBootloader(resp *http.Response) (*ArduinoBuilderBoardBootloader, error) {
	var decoded ArduinoBuilderBoardBootloader
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// ArduinoBuilderBoardBuild contains the info used to compile for a certain flavour of board. (default view)
//
// Identifier: application/vnd.arduino.builder.board.build; view=default
type ArduinoBuilderBoardBuild struct {
	// The flavour of the board. Usually it's default
	Flavour *string `form:"flavour,omitempty" json:"flavour,omitempty" xml:"flavour,omitempty"`
	// An identifier used by the tools to determine which tools to use on it
	Fqbn *string `form:"fqbn,omitempty" json:"fqbn,omitempty" xml:"fqbn,omitempty"`
}

// DecodeArduinoBuilderBoardBuild decodes the ArduinoBuilderBoardBuild instance encoded in resp body.
func (c *Client) DecodeArduinoBuilderBoardBuild(resp *http.Response) (*ArduinoBuilderBoardBuild, error) {
	var decoded ArduinoBuilderBoardBuild
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// ArduinoBuilderBoardUpload contains the info used to upload a certain flavour of board. (default view)
//
// Identifier: application/vnd.arduino.builder.board.upload; view=default
type ArduinoBuilderBoardUpload struct {
	// The commandline used to upload sketches
	Commandline *string `form:"commandline,omitempty" json:"commandline,omitempty" xml:"commandline,omitempty"`
	// The extension of the binary file
	Ext *string `form:"ext,omitempty" json:"ext,omitempty" xml:"ext,omitempty"`
	// Files used by the programmer
	Files ArduinoBuilderFileCollection `form:"files,omitempty" json:"files,omitempty" xml:"files,omitempty"`
	// The flavour of the board. Usually it's default
	Flavour *string `form:"flavour,omitempty" json:"flavour,omitempty" xml:"flavour,omitempty"`
	// Some options used for uploading, like the speed.
	Options map[string]string `form:"options,omitempty" json:"options,omitempty" xml:"options,omitempty"`
	// Some params used for uploading. Usually quiet and verbose.
	Params map[string]string `form:"params,omitempty" json:"params,omitempty" xml:"params,omitempty"`
	// The tool to use for uploading sketches
	Tool *string `form:"tool,omitempty" json:"tool,omitempty" xml:"tool,omitempty"`
	// The version of the tool
	ToolVersion *string `form:"tool_version,omitempty" json:"tool_version,omitempty" xml:"tool_version,omitempty"`
}

// DecodeArduinoBuilderBoardUpload decodes the ArduinoBuilderBoardUpload instance encoded in resp body.
func (c *Client) DecodeArduinoBuilderBoardUpload(resp *http.Response) (*ArduinoBuilderBoardUpload, error) {
	var decoded ArduinoBuilderBoardUpload
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// ArduinoBuilderBoardCollection is the media type for an array of ArduinoBuilderBoard (default view)
//
// Identifier: application/vnd.arduino.builder.board+json; type=collection; view=default
type ArduinoBuilderBoardCollection []*ArduinoBuilderBoard

// DecodeArduinoBuilderBoardCollection decodes the ArduinoBuilderBoardCollection instance encoded in resp body.
func (c *Client) DecodeArduinoBuilderBoardCollection(resp *http.Response) (ArduinoBuilderBoardCollection, error) {
	var decoded ArduinoBuilderBoardCollection
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return decoded, err
}

// ArduinoBuilderBoardsv2 is a paginated list of boards (default view)
//
// Identifier: application/vnd.arduino.builder.boardsv2+json; view=default
type ArduinoBuilderBoardsv2 struct {
	// The list of sketches
	Items ArduinoBuilderBoardv2Collection `form:"items" json:"items" xml:"items"`
	// Link to the following page of results. Could be empty.
	Next *string `form:"next,omitempty" json:"next,omitempty" xml:"next,omitempty"`
	// Link to the previous page of results. Could be empty.
	Prev *string `form:"prev,omitempty" json:"prev,omitempty" xml:"prev,omitempty"`
}

// Validate validates the ArduinoBuilderBoardsv2 media type instance.
func (mt *ArduinoBuilderBoardsv2) Validate() (err error) {
	if mt.Items == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "items"))
	}
	return
}

// DecodeArduinoBuilderBoardsv2 decodes the ArduinoBuilderBoardsv2 instance encoded in resp body.
func (c *Client) DecodeArduinoBuilderBoardsv2(resp *http.Response) (*ArduinoBuilderBoardsv2, error) {
	var decoded ArduinoBuilderBoardsv2
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// ArduinoBuilderBoardv2 is a physical board belonging to a certain architecture in a package. The most obvious package is arduino, which contains architectures avr, sam and samd. It can contain multiple versions of the upload commands and options. If there is a default version it means that it's the only version officially supported. Of course if there is only one version it will be called default (default view)
//
// Identifier: application/vnd.arduino.builder.boardv2+json; view=default
type ArduinoBuilderBoardv2 struct {
	// The architecture of the board
	Architecture *string                              `form:"architecture,omitempty" json:"architecture,omitempty" xml:"architecture,omitempty"`
	Build        ArduinoBuilderBoardv2BuildCollection `form:"build,omitempty" json:"build,omitempty" xml:"build,omitempty"`
	// The default flavour of the board
	DefaultFlavour *string `form:"default_flavour,omitempty" json:"default_flavour,omitempty" xml:"default_flavour,omitempty"`
	// An identifier used by the tools to determine which tools to use on it
	Fqbn *string `form:"fqbn,omitempty" json:"fqbn,omitempty" xml:"fqbn,omitempty"`
	// The url of the board
	Href *string `form:"href,omitempty" json:"href,omitempty" xml:"href,omitempty"`
	// The name of the board
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
}

// ArduinoBuilderBoardv2Full is a physical board belonging to a certain architecture in a package. The most obvious package is arduino, which contains architectures avr, sam and samd. It can contain multiple versions of the upload commands and options. If there is a default version it means that it's the only version officially supported. Of course if there is only one version it will be called default (full view)
//
// Identifier: application/vnd.arduino.builder.boardv2+json; view=full
type ArduinoBuilderBoardv2Full struct {
	// The architecture of the board
	Architecture *string                                   `form:"architecture,omitempty" json:"architecture,omitempty" xml:"architecture,omitempty"`
	Bootloader   ArduinoBuilderBoardv2BootloaderCollection `form:"bootloader,omitempty" json:"bootloader,omitempty" xml:"bootloader,omitempty"`
	Build        ArduinoBuilderBoardv2BuildCollection      `form:"build,omitempty" json:"build,omitempty" xml:"build,omitempty"`
	// The default flavour of the board
	DefaultFlavour *string `form:"default_flavour,omitempty" json:"default_flavour,omitempty" xml:"default_flavour,omitempty"`
	// An identifier used by the tools to determine which tools to use on it
	Fqbn *string `form:"fqbn,omitempty" json:"fqbn,omitempty" xml:"fqbn,omitempty"`
	// The url of the board
	Href *string `form:"href,omitempty" json:"href,omitempty" xml:"href,omitempty"`
	// The id of the board
	ID *string `form:"id,omitempty" json:"id,omitempty" xml:"id,omitempty"`
	// The name of the board
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// The package to which the board belongs
	Package *string `form:"package,omitempty" json:"package,omitempty" xml:"package,omitempty"`
	// A list of possible pids
	Pid    []string                              `form:"pid,omitempty" json:"pid,omitempty" xml:"pid,omitempty"`
	Upload ArduinoBuilderBoardv2UploadCollection `form:"upload,omitempty" json:"upload,omitempty" xml:"upload,omitempty"`
	// A list of possible vids
	Vid []string `form:"vid,omitempty" json:"vid,omitempty" xml:"vid,omitempty"`
}

// DecodeArduinoBuilderBoardv2 decodes the ArduinoBuilderBoardv2 instance encoded in resp body.
func (c *Client) DecodeArduinoBuilderBoardv2(resp *http.Response) (*ArduinoBuilderBoardv2, error) {
	var decoded ArduinoBuilderBoardv2
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// DecodeArduinoBuilderBoardv2Full decodes the ArduinoBuilderBoardv2Full instance encoded in resp body.
func (c *Client) DecodeArduinoBuilderBoardv2Full(resp *http.Response) (*ArduinoBuilderBoardv2Full, error) {
	var decoded ArduinoBuilderBoardv2Full
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// ArduinoBuilderBoardv2Bootloader contains the info used to bootload a board. (default view)
//
// Identifier: application/vnd.arduino.builder.boardv2.bootloader; view=default
type ArduinoBuilderBoardv2Bootloader struct {
	// The commandline used to bootload
	Commandline *string `form:"commandline,omitempty" json:"commandline,omitempty" xml:"commandline,omitempty"`
	// The flavour of the board. Usually it's default
	Flavour *string `form:"flavour,omitempty" json:"flavour,omitempty" xml:"flavour,omitempty"`
	// The signature of the commandline
	Signature *string `form:"signature,omitempty" json:"signature,omitempty" xml:"signature,omitempty"`
}

// DecodeArduinoBuilderBoardv2Bootloader decodes the ArduinoBuilderBoardv2Bootloader instance encoded in resp body.
func (c *Client) DecodeArduinoBuilderBoardv2Bootloader(resp *http.Response) (*ArduinoBuilderBoardv2Bootloader, error) {
	var decoded ArduinoBuilderBoardv2Bootloader
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// ArduinoBuilderBoardv2BootloaderCollection is the media type for an array of ArduinoBuilderBoardv2Bootloader (default view)
//
// Identifier: application/vnd.arduino.builder.boardv2.bootloader; type=collection; view=default
type ArduinoBuilderBoardv2BootloaderCollection []*ArduinoBuilderBoardv2Bootloader

// DecodeArduinoBuilderBoardv2BootloaderCollection decodes the ArduinoBuilderBoardv2BootloaderCollection instance encoded in resp body.
func (c *Client) DecodeArduinoBuilderBoardv2BootloaderCollection(resp *http.Response) (ArduinoBuilderBoardv2BootloaderCollection, error) {
	var decoded ArduinoBuilderBoardv2BootloaderCollection
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return decoded, err
}

// Build contains the info used to compile for a certain flavour of board. (default view)
//
// Identifier: application/vnd.arduino.builder.boardv2.build; view=default
type ArduinoBuilderBoardv2Build struct {
	// The flavour of the board. Usually it's default
	Flavour *string `form:"flavour,omitempty" json:"flavour,omitempty" xml:"flavour,omitempty"`
	// An identifier used by the tools to determine which tools to use on it
	Fqbn *string `form:"fqbn,omitempty" json:"fqbn,omitempty" xml:"fqbn,omitempty"`
}

// DecodeArduinoBuilderBoardv2Build decodes the ArduinoBuilderBoardv2Build instance encoded in resp body.
func (c *Client) DecodeArduinoBuilderBoardv2Build(resp *http.Response) (*ArduinoBuilderBoardv2Build, error) {
	var decoded ArduinoBuilderBoardv2Build
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// ArduinoBuilderBoardv2BuildCollection is the media type for an array of ArduinoBuilderBoardv2Build (default view)
//
// Identifier: application/vnd.arduino.builder.boardv2.build; type=collection; view=default
type ArduinoBuilderBoardv2BuildCollection []*ArduinoBuilderBoardv2Build

// DecodeArduinoBuilderBoardv2BuildCollection decodes the ArduinoBuilderBoardv2BuildCollection instance encoded in resp body.
func (c *Client) DecodeArduinoBuilderBoardv2BuildCollection(resp *http.Response) (ArduinoBuilderBoardv2BuildCollection, error) {
	var decoded ArduinoBuilderBoardv2BuildCollection
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return decoded, err
}

// ArduinoBuilderBoardv2Upload contains the info used to upload a certain flavour of board. (default view)
//
// Identifier: application/vnd.arduino.builder.boardv2.upload; view=default
type ArduinoBuilderBoardv2Upload struct {
	// The commandline used to upload sketches
	Commandline *string `form:"commandline,omitempty" json:"commandline,omitempty" xml:"commandline,omitempty"`
	// The extension of the binary file
	Ext *string `form:"ext,omitempty" json:"ext,omitempty" xml:"ext,omitempty"`
	// Files used by the programmer
	Files ArduinoBuilderFileCollection `form:"files,omitempty" json:"files,omitempty" xml:"files,omitempty"`
	// The flavour of the board. Usually it's default
	Flavour *string `form:"flavour,omitempty" json:"flavour,omitempty" xml:"flavour,omitempty"`
	// Some options used for uploading, like the speed.
	Options map[string]string `form:"options,omitempty" json:"options,omitempty" xml:"options,omitempty"`
	// Some params used for uploading. Usually quiet and verbose.
	Params map[string]string `form:"params,omitempty" json:"params,omitempty" xml:"params,omitempty"`
	// The tool to use for uploading sketches
	Tool *string `form:"tool,omitempty" json:"tool,omitempty" xml:"tool,omitempty"`
	// The version of the tool
	ToolVersion *string `form:"tool_version,omitempty" json:"tool_version,omitempty" xml:"tool_version,omitempty"`
}

// DecodeArduinoBuilderBoardv2Upload decodes the ArduinoBuilderBoardv2Upload instance encoded in resp body.
func (c *Client) DecodeArduinoBuilderBoardv2Upload(resp *http.Response) (*ArduinoBuilderBoardv2Upload, error) {
	var decoded ArduinoBuilderBoardv2Upload
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// ArduinoBuilderBoardv2UploadCollection is the media type for an array of ArduinoBuilderBoardv2Upload (default view)
//
// Identifier: application/vnd.arduino.builder.boardv2.upload; type=collection; view=default
type ArduinoBuilderBoardv2UploadCollection []*ArduinoBuilderBoardv2Upload

// DecodeArduinoBuilderBoardv2UploadCollection decodes the ArduinoBuilderBoardv2UploadCollection instance encoded in resp body.
func (c *Client) DecodeArduinoBuilderBoardv2UploadCollection(resp *http.Response) (ArduinoBuilderBoardv2UploadCollection, error) {
	var decoded ArduinoBuilderBoardv2UploadCollection
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return decoded, err
}

// ArduinoBuilderBoardv2Collection is the media type for an array of ArduinoBuilderBoardv2 (default view)
//
// Identifier: application/vnd.arduino.builder.boardv2+json; type=collection; view=default
type ArduinoBuilderBoardv2Collection []*ArduinoBuilderBoardv2

// ArduinoBuilderBoardv2FullCollection is the media type for an array of ArduinoBuilderBoardv2 (full view)
//
// Identifier: application/vnd.arduino.builder.boardv2+json; type=collection; view=full
type ArduinoBuilderBoardv2FullCollection []*ArduinoBuilderBoardv2Full

// DecodeArduinoBuilderBoardv2Collection decodes the ArduinoBuilderBoardv2Collection instance encoded in resp body.
func (c *Client) DecodeArduinoBuilderBoardv2Collection(resp *http.Response) (ArduinoBuilderBoardv2Collection, error) {
	var decoded ArduinoBuilderBoardv2Collection
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return decoded, err
}

// DecodeArduinoBuilderBoardv2FullCollection decodes the ArduinoBuilderBoardv2FullCollection instance encoded in resp body.
func (c *Client) DecodeArduinoBuilderBoardv2FullCollection(resp *http.Response) (ArduinoBuilderBoardv2FullCollection, error) {
	var decoded ArduinoBuilderBoardv2FullCollection
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return decoded, err
}

// ArduinoBuilderCompilationResult is the result of a compilation. It contains the output and the eventual errors. If successful it contains the generated files. (default view)
//
// Identifier: application/vnd.arduino.builder.compilation.result; view=default
type ArduinoBuilderCompilationResult struct {
	// A base64 encoded file with the extension .bin. Can be one of the artifacts generated by the compilation.
	Bin *string `form:"bin,omitempty" json:"bin,omitempty" xml:"bin,omitempty"`
	// A base64 encoded file with the extension .elf. Can be one of the artifacts generated by the compilation.
	Elf *string `form:"elf,omitempty" json:"elf,omitempty" xml:"elf,omitempty"`
	// A base64 encoded file with the extension .hex. Can be one of the artifacts generated by the compilation.
	Hex *string `form:"hex,omitempty" json:"hex,omitempty" xml:"hex,omitempty"`
	// The id of the compilation that has been saved on the database
	ID *string `form:"id,omitempty" json:"id,omitempty" xml:"id,omitempty"`
	// The stderr of the compilation. If it's present the compilation wasn't successful.
	Stderr *string `form:"stderr,omitempty" json:"stderr,omitempty" xml:"stderr,omitempty"`
	// The stdout of the compilation. If the verbose parameter was true it will be much longer.
	Stdout *string `form:"stdout,omitempty" json:"stdout,omitempty" xml:"stdout,omitempty"`
}

// DecodeArduinoBuilderCompilationResult decodes the ArduinoBuilderCompilationResult instance encoded in resp body.
func (c *Client) DecodeArduinoBuilderCompilationResult(resp *http.Response) (*ArduinoBuilderCompilationResult, error) {
	var decoded ArduinoBuilderCompilationResult
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// An ArduinoBuilderExample is a simple sketch with the purpose of demonstrating the capabilities of the language. (default view)
//
// Identifier: application/vnd.arduino.builder.example+json; view=default
type ArduinoBuilderExample struct {
	// The files contained in the example
	Files []*ArduinoBuilderFile `form:"files,omitempty" json:"files,omitempty" xml:"files,omitempty"`
	// The folder of the example. It's a way to categorize them, it doesn't necessarily translate to a folder in the filesystem.
	Folder *string `form:"folder,omitempty" json:"folder,omitempty" xml:"folder,omitempty"`
	// The url where to find the details
	Href *string `form:"href,omitempty" json:"href,omitempty" xml:"href,omitempty"`
	// The main file
	Ino *ArduinoBuilderFile `form:"ino,omitempty" json:"ino,omitempty" xml:"ino,omitempty"`
	// The name of the example
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// The path of the example, where to find it on the filesystem
	Path *string `form:"path,omitempty" json:"path,omitempty" xml:"path,omitempty"`
	// A list of tags. The builtin tag means that it's a builtin example.
	Types []string `form:"types,omitempty" json:"types,omitempty" xml:"types,omitempty"`
}

// An ArduinoBuilderExampleLink is a simple sketch with the purpose of demonstrating the capabilities of the language. (link view)
//
// Identifier: application/vnd.arduino.builder.example+json; view=link
type ArduinoBuilderExampleLink struct {
	// The url where to find the details
	Href *string `form:"href,omitempty" json:"href,omitempty" xml:"href,omitempty"`
}

// DecodeArduinoBuilderExample decodes the ArduinoBuilderExample instance encoded in resp body.
func (c *Client) DecodeArduinoBuilderExample(resp *http.Response) (*ArduinoBuilderExample, error) {
	var decoded ArduinoBuilderExample
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// DecodeArduinoBuilderExampleLink decodes the ArduinoBuilderExampleLink instance encoded in resp body.
func (c *Client) DecodeArduinoBuilderExampleLink(resp *http.Response) (*ArduinoBuilderExampleLink, error) {
	var decoded ArduinoBuilderExampleLink
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// ArduinoBuilderExampleCollection is the media type for an array of ArduinoBuilderExample (default view)
//
// Identifier: application/vnd.arduino.builder.example+json; type=collection; view=default
type ArduinoBuilderExampleCollection []*ArduinoBuilderExample

// ArduinoBuilderExampleLinkCollection is the media type for an array of ArduinoBuilderExample (link view)
//
// Identifier: application/vnd.arduino.builder.example+json; type=collection; view=link
type ArduinoBuilderExampleLinkCollection []*ArduinoBuilderExampleLink

// DecodeArduinoBuilderExampleCollection decodes the ArduinoBuilderExampleCollection instance encoded in resp body.
func (c *Client) DecodeArduinoBuilderExampleCollection(resp *http.Response) (ArduinoBuilderExampleCollection, error) {
	var decoded ArduinoBuilderExampleCollection
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return decoded, err
}

// DecodeArduinoBuilderExampleLinkCollection decodes the ArduinoBuilderExampleLinkCollection instance encoded in resp body.
func (c *Client) DecodeArduinoBuilderExampleLinkCollection(resp *http.Response) (ArduinoBuilderExampleLinkCollection, error) {
	var decoded ArduinoBuilderExampleLinkCollection
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return decoded, err
}

// ArduinoBuilderFile represents a file in the filesystem, belonging to a sketch, a library or an example (default view)
//
// Identifier: application/vnd.arduino.builder.file; view=default
type ArduinoBuilderFile struct {
	// The contents of the file, in base64
	Data *string `form:"data,omitempty" json:"data,omitempty" xml:"data,omitempty"`
	// The url where to find the details
	Href *string `form:"href,omitempty" json:"href,omitempty" xml:"href,omitempty"`
	// The last time it has been modified
	LastModified *time.Time `form:"last_modified,omitempty" json:"last_modified,omitempty" xml:"last_modified,omitempty"`
	// The mimetype of the file.
	Mimetype *string `form:"mimetype,omitempty" json:"mimetype,omitempty" xml:"mimetype,omitempty"`
	// The name and extension of the file
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// The path of the file, where to find it on the filesystem
	Path *string `form:"path,omitempty" json:"path,omitempty" xml:"path,omitempty"`
}

// DecodeArduinoBuilderFile decodes the ArduinoBuilderFile instance encoded in resp body.
func (c *Client) DecodeArduinoBuilderFile(resp *http.Response) (*ArduinoBuilderFile, error) {
	var decoded ArduinoBuilderFile
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// ArduinoBuilderFileCollection is the media type for an array of ArduinoBuilderFile (default view)
//
// Identifier: application/vnd.arduino.builder.file; type=collection; view=default
type ArduinoBuilderFileCollection []*ArduinoBuilderFile

// DecodeArduinoBuilderFileCollection decodes the ArduinoBuilderFileCollection instance encoded in resp body.
func (c *Client) DecodeArduinoBuilderFileCollection(resp *http.Response) (ArduinoBuilderFileCollection, error) {
	var decoded ArduinoBuilderFileCollection
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return decoded, err
}

// ArduinoBuilderLibrary is a collection of header files containing arduino reusable code and functions. It typically contains its info in a library.properties files. The examples property contains a list of examples that use that library. (default view)
//
// Identifier: application/vnd.arduino.builder.library+json; view=default
type ArduinoBuilderLibrary struct {
	// The architectures supported by the library.
	Architectures []string `form:"architectures,omitempty" json:"architectures,omitempty" xml:"architectures,omitempty"`
	// A category
	Category *string `form:"category,omitempty" json:"category,omitempty" xml:"category,omitempty"`
	// A snippet of code that includes all of the library header files
	Code *string `form:"code,omitempty" json:"code,omitempty" xml:"code,omitempty"`
	// The examples contained in the library
	Examples []*ArduinoBuilderExample `form:"examples,omitempty" json:"examples,omitempty" xml:"examples,omitempty"`
	// The number of examples that it contains
	ExamplesNumber *int `form:"examples_number,omitempty" json:"examples_number,omitempty" xml:"examples_number,omitempty"`
	// The files contained in the library
	Files []*ArduinoBuilderFile `form:"files,omitempty" json:"files,omitempty" xml:"files,omitempty"`
	// The url where to find the details
	Href *string `form:"href,omitempty" json:"href,omitempty" xml:"href,omitempty"`
	// The id of the library. It could be a combination of name and version, a combination of the package and architecture, or an uuid id
	ID *string `form:"id,omitempty" json:"id,omitempty" xml:"id,omitempty"`
	// The maintainer of the library
	Maintainer *string `form:"maintainer,omitempty" json:"maintainer,omitempty" xml:"maintainer,omitempty"`
	// The name of the library, shared between many versions
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// Other versions of this library
	OtherVersions []string `form:"other_versions,omitempty" json:"other_versions,omitempty" xml:"other_versions,omitempty"`
	// A short description
	Sentence *string `form:"sentence,omitempty" json:"sentence,omitempty" xml:"sentence,omitempty"`
	// A list of tags. The Arduino tag means that it's a builtin library.
	Types []string `form:"types,omitempty" json:"types,omitempty" xml:"types,omitempty"`
	// The homepage of the library
	URL *string `form:"url,omitempty" json:"url,omitempty" xml:"url,omitempty"`
	// The version of the library
	Version *string `form:"version,omitempty" json:"version,omitempty" xml:"version,omitempty"`
}

// ArduinoBuilderLibraryLink is a collection of header files containing arduino reusable code and functions. It typically contains its info in a library.properties files. The examples property contains a list of examples that use that library. (link view)
//
// Identifier: application/vnd.arduino.builder.library+json; view=link
type ArduinoBuilderLibraryLink struct {
	// The url where to find the details
	Href *string `form:"href,omitempty" json:"href,omitempty" xml:"href,omitempty"`
}

// DecodeArduinoBuilderLibrary decodes the ArduinoBuilderLibrary instance encoded in resp body.
func (c *Client) DecodeArduinoBuilderLibrary(resp *http.Response) (*ArduinoBuilderLibrary, error) {
	var decoded ArduinoBuilderLibrary
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// DecodeArduinoBuilderLibraryLink decodes the ArduinoBuilderLibraryLink instance encoded in resp body.
func (c *Client) DecodeArduinoBuilderLibraryLink(resp *http.Response) (*ArduinoBuilderLibraryLink, error) {
	var decoded ArduinoBuilderLibraryLink
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// ArduinoBuilderSlimlibrary is a partial view of a library. (default view)
//
// Identifier: application/vnd.arduino.builder.slimlibrary+json; view=default
type ArduinoBuilderSlimlibrary struct {
	// The architectures supported by the library.
	Architectures []string `form:"architectures,omitempty" json:"architectures,omitempty" xml:"architectures,omitempty"`
	// A category
	Category *string `form:"category,omitempty" json:"category,omitempty" xml:"category,omitempty"`
	// A snippet of code that includes all of the library header files
	Code *string `form:"code,omitempty" json:"code,omitempty" xml:"code,omitempty"`
	// The number of examples that it contains
	ExamplesNumber *int    `form:"examples_number,omitempty" json:"examples_number,omitempty" xml:"examples_number,omitempty"`
	Href           *string `form:"href,omitempty" json:"href,omitempty" xml:"href,omitempty"`
	// The id of the library. It could be a combination of name and version, a combination of the package and architecture, or an uuid id
	ID *string `form:"id,omitempty" json:"id,omitempty" xml:"id,omitempty"`
	// The maintainer of the library
	Maintainer *string `form:"maintainer,omitempty" json:"maintainer,omitempty" xml:"maintainer,omitempty"`
	// The name of the library, shared between many versions
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// A short description
	Sentence *string `form:"sentence,omitempty" json:"sentence,omitempty" xml:"sentence,omitempty"`
	// A list of tags. The Arduino tag means that it's a builtin library.
	Types []string `form:"types,omitempty" json:"types,omitempty" xml:"types,omitempty"`
	// The homepage of the library
	URL *string `form:"url,omitempty" json:"url,omitempty" xml:"url,omitempty"`
	// The version of the library
	Version *string `form:"version,omitempty" json:"version,omitempty" xml:"version,omitempty"`
}

// DecodeArduinoBuilderSlimlibrary decodes the ArduinoBuilderSlimlibrary instance encoded in resp body.
func (c *Client) DecodeArduinoBuilderSlimlibrary(resp *http.Response) (*ArduinoBuilderSlimlibrary, error) {
	var decoded ArduinoBuilderSlimlibrary
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return &decoded, err
}

// ArduinoBuilderSlimlibraryCollection is the media type for an array of ArduinoBuilderSlimlibrary (default view)
//
// Identifier: application/vnd.arduino.builder.slimlibrary+json; type=collection; view=default
type ArduinoBuilderSlimlibraryCollection []*ArduinoBuilderSlimlibrary

// DecodeArduinoBuilderSlimlibraryCollection decodes the ArduinoBuilderSlimlibraryCollection instance encoded in resp body.
func (c *Client) DecodeArduinoBuilderSlimlibraryCollection(resp *http.Response) (ArduinoBuilderSlimlibraryCollection, error) {
	var decoded ArduinoBuilderSlimlibraryCollection
	err := c.Decoder.Decode(&decoded, resp.Body, resp.Header.Get("Content-Type"))
	return decoded, err
}
