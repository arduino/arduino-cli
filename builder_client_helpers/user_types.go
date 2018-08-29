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
	"github.com/goadesign/goa"
)

// A compilation is made up of a sketch (or a path to a sketch) and an fqbn.
// Eventual libraries are automatically determined and linked.
// NOTE: swagger will force you to define the files inside the sketch because of a bug. But they are not mandatory.
type compilation struct {
	// The fully qualified board name
	Fqbn *string `form:"fqbn,omitempty" json:"fqbn,omitempty" xml:"fqbn,omitempty"`
	// The path of the sketch or example to compile. Mandatory if sketch parameter is missing.
	Path *string `form:"path,omitempty" json:"path,omitempty" xml:"path,omitempty"`
	// The full sketch to compile. Mandatory if path parameter is missing.
	Sketch *sketch `form:"sketch,omitempty" json:"sketch,omitempty" xml:"sketch,omitempty"`
	// An option stating if the output should be verbose. Defaults to false.
	Verbose *bool `form:"verbose,omitempty" json:"verbose,omitempty" xml:"verbose,omitempty"`
}

// Finalize sets the default values for compilation type instance.
func (ut *compilation) Finalize() {
	var defaultVerbose = false
	if ut.Verbose == nil {
		ut.Verbose = &defaultVerbose
	}
}

// Validate validates the compilation type instance.
func (ut *compilation) Validate() (err error) {
	if ut.Fqbn == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "fqbn"))
	}
	if ut.Sketch != nil {
		if err2 := ut.Sketch.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// Publicize creates Compilation from compilation
func (ut *compilation) Publicize() *Compilation {
	var pub Compilation
	if ut.Fqbn != nil {
		pub.Fqbn = *ut.Fqbn
	}
	if ut.Path != nil {
		pub.Path = ut.Path
	}
	if ut.Sketch != nil {
		pub.Sketch = ut.Sketch.Publicize()
	}
	if ut.Verbose != nil {
		pub.Verbose = *ut.Verbose
	}
	return &pub
}

// A Compilation is made up of a sketch (or a path to a sketch) and an fqbn.
// Eventual libraries are automatically determined and linked.
// NOTE: swagger will force you to define the files inside the sketch because of a bug. But they are not mandatory.
type Compilation struct {
	// The fully qualified board name
	Fqbn string `form:"fqbn" json:"fqbn" xml:"fqbn"`
	// The path of the sketch or example to compile. Mandatory if sketch parameter is missing.
	Path *string `form:"path,omitempty" json:"path,omitempty" xml:"path,omitempty"`
	// The full sketch to compile. Mandatory if path parameter is missing.
	Sketch *Sketch `form:"sketch,omitempty" json:"sketch,omitempty" xml:"sketch,omitempty"`
	// An option stating if the output should be verbose. Defaults to false.
	Verbose bool `form:"verbose" json:"verbose" xml:"verbose"`
}

// Validate validates the Compilation type instance.
func (ut *Compilation) Validate() (err error) {
	if ut.Fqbn == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "fqbn"))
	}
	if ut.Sketch != nil {
		if err2 := ut.Sketch.Validate(); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// An example is a simple sketch with the purpose of demonstrating the capabilities of the language.
type example struct {
	// Other files contained in the example
	Files []*filemeta `form:"files,omitempty" json:"files,omitempty" xml:"files,omitempty"`
	// The folder of the example. It's a way to categorize them, it doesn't necessarily translate to a folder in the filesystem.
	Folder *string `form:"folder,omitempty" json:"folder,omitempty" xml:"folder,omitempty"`
	// The url where to find the details
	Href *string `form:"href,omitempty" json:"href,omitempty" xml:"href,omitempty"`
	// The main file
	Ino *filemeta `form:"ino,omitempty" json:"ino,omitempty" xml:"ino,omitempty"`
	// The name of the example
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// The path of the example, where to find it on the filesystem
	Path *string `form:"path,omitempty" json:"path,omitempty" xml:"path,omitempty"`
	// A list of tags. The builtin tag means that it's a builtin example.
	Types []string `form:"types,omitempty" json:"types,omitempty" xml:"types,omitempty"`
}

// Validate validates the example type instance.
func (ut *example) Validate() (err error) {
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

// Publicize creates Example from example
func (ut *example) Publicize() *Example {
	var pub Example
	if ut.Files != nil {
		pub.Files = make([]*Filemeta, len(ut.Files))
		for i2, elem2 := range ut.Files {
			pub.Files[i2] = elem2.Publicize()
		}
	}
	if ut.Folder != nil {
		pub.Folder = ut.Folder
	}
	if ut.Href != nil {
		pub.Href = ut.Href
	}
	if ut.Ino != nil {
		pub.Ino = ut.Ino.Publicize()
	}
	if ut.Name != nil {
		pub.Name = ut.Name
	}
	if ut.Path != nil {
		pub.Path = ut.Path
	}
	if ut.Types != nil {
		pub.Types = ut.Types
	}
	return &pub
}

// An example is a simple sketch with the purpose of demonstrating the capabilities of the language.
type Example struct {
	// Other files contained in the example
	Files []*Filemeta `form:"files,omitempty" json:"files,omitempty" xml:"files,omitempty"`
	// The folder of the example. It's a way to categorize them, it doesn't necessarily translate to a folder in the filesystem.
	Folder *string `form:"folder,omitempty" json:"folder,omitempty" xml:"folder,omitempty"`
	// The url where to find the details
	Href *string `form:"href,omitempty" json:"href,omitempty" xml:"href,omitempty"`
	// The main file
	Ino *Filemeta `form:"ino,omitempty" json:"ino,omitempty" xml:"ino,omitempty"`
	// The name of the example
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// The path of the example, where to find it on the filesystem
	Path *string `form:"path,omitempty" json:"path,omitempty" xml:"path,omitempty"`
	// A list of tags. The builtin tag means that it's a builtin example.
	Types []string `form:"types,omitempty" json:"types,omitempty" xml:"types,omitempty"`
}

// Validate validates the Example type instance.
func (ut *Example) Validate() (err error) {
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

// FileFull represent a file in the filesystem, belonging to a sketch, a library or an example. Must contain a data property with the content of the file.
type filefull struct {
	// The contents of the file, in base64
	Data *string `form:"data,omitempty" json:"data,omitempty" xml:"data,omitempty"`
	// The name and extension of the file
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
}

// Validate validates the filefull type instance.
func (ut *filefull) Validate() (err error) {
	if ut.Name == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "name"))
	}
	if ut.Data == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "data"))
	}
	return
}

// Publicize creates Filefull from filefull
func (ut *filefull) Publicize() *Filefull {
	var pub Filefull
	if ut.Data != nil {
		pub.Data = *ut.Data
	}
	if ut.Name != nil {
		pub.Name = *ut.Name
	}
	return &pub
}

// FileFull represent a file in the filesystem, belonging to a sketch, a library or an example. Must contain a data property with the content of the file.
type Filefull struct {
	// The contents of the file, in base64
	Data string `form:"data" json:"data" xml:"data"`
	// The name and extension of the file
	Name string `form:"name" json:"name" xml:"name"`
}

// Validate validates the Filefull type instance.
func (ut *Filefull) Validate() (err error) {
	if ut.Name == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "name"))
	}
	if ut.Data == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "data"))
	}
	return
}

// FileMeta represent a file in the filesystem, belonging to a sketch, a library or an example. Can contain a data property with the content of the file.
type filemeta struct {
	// The contents of the file, in base64
	Data *string `form:"data,omitempty" json:"data,omitempty" xml:"data,omitempty"`
	// The name and extension of the file
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
}

// Validate validates the filemeta type instance.
func (ut *filemeta) Validate() (err error) {
	if ut.Name == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "name"))
	}
	return
}

// Publicize creates Filemeta from filemeta
func (ut *filemeta) Publicize() *Filemeta {
	var pub Filemeta
	if ut.Data != nil {
		pub.Data = ut.Data
	}
	if ut.Name != nil {
		pub.Name = *ut.Name
	}
	return &pub
}

// FileMeta represent a file in the filesystem, belonging to a sketch, a library or an example.
// Can contain a data property with the content of the file.
type Filemeta struct {
	// The contents of the file, in base64
	Data *string `form:"data,omitempty" json:"data,omitempty" xml:"data,omitempty"`
	// The name and extension of the file
	Name string `form:"name" json:"name" xml:"name"`
}

// Validate validates the Filemeta type instance.
func (ut *Filemeta) Validate() (err error) {
	if ut.Name == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "name"))
	}
	return
}

// Library is a collection of header files containing arduino reusable code and functions.
// It typically contains its info in a library.properties files.
// The examples property contains a list of examples that use that library.
type library struct {
	// The architectures supported by the library.
	Architectures []string `form:"architectures,omitempty" json:"architectures,omitempty" xml:"architectures,omitempty"`
	// A category
	Category *string `form:"category,omitempty" json:"category,omitempty" xml:"category,omitempty"`
	// A snippet of code that includes all of the library header files
	Code *string `form:"code,omitempty" json:"code,omitempty" xml:"code,omitempty"`
	// The examples contained in the library
	Examples []*example `form:"examples,omitempty" json:"examples,omitempty" xml:"examples,omitempty"`
	// The number of examples that it contains
	ExamplesNumber *int `form:"examples_number,omitempty" json:"examples_number,omitempty" xml:"examples_number,omitempty"`
	// The files contained in the library
	Files []*filemeta `form:"files,omitempty" json:"files,omitempty" xml:"files,omitempty"`
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

// Validate validates the library type instance.
func (ut *library) Validate() (err error) {
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
		pub.Examples = make([]*Example, len(ut.Examples))
		for i2, elem2 := range ut.Examples {
			pub.Examples[i2] = elem2.Publicize()
		}
	}
	if ut.ExamplesNumber != nil {
		pub.ExamplesNumber = ut.ExamplesNumber
	}
	if ut.Files != nil {
		pub.Files = make([]*Filemeta, len(ut.Files))
		for i2, elem2 := range ut.Files {
			pub.Files[i2] = elem2.Publicize()
		}
	}
	if ut.ID != nil {
		pub.ID = ut.ID
	}
	if ut.Maintainer != nil {
		pub.Maintainer = ut.Maintainer
	}
	if ut.Name != nil {
		pub.Name = ut.Name
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

// Library is a collection of header files containing arduino reusable code and functions.
// It typically contains its info in a library.properties files.
// The examples property contains a list of examples that use that library.
type Library struct {
	// The architectures supported by the library.
	Architectures []string `form:"architectures,omitempty" json:"architectures,omitempty" xml:"architectures,omitempty"`
	// A category
	Category *string `form:"category,omitempty" json:"category,omitempty" xml:"category,omitempty"`
	// A snippet of code that includes all of the library header files
	Code *string `form:"code,omitempty" json:"code,omitempty" xml:"code,omitempty"`
	// The examples contained in the library
	Examples []*Example `form:"examples,omitempty" json:"examples,omitempty" xml:"examples,omitempty"`
	// The number of examples that it contains
	ExamplesNumber *int `form:"examples_number,omitempty" json:"examples_number,omitempty" xml:"examples_number,omitempty"`
	// The files contained in the library
	Files []*Filemeta `form:"files,omitempty" json:"files,omitempty" xml:"files,omitempty"`
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

// Validate validates the Library type instance.
func (ut *Library) Validate() (err error) {
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

// pinnedLib user type.
type pinnedLib struct {
	// The slugified name of the library
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// The version of the library. Can be latest
	Version *string `form:"version,omitempty" json:"version,omitempty" xml:"version,omitempty"`
}

// Publicize creates PinnedLib from pinnedLib
func (ut *pinnedLib) Publicize() *PinnedLib {
	var pub PinnedLib
	if ut.Name != nil {
		pub.Name = ut.Name
	}
	if ut.Version != nil {
		pub.Version = ut.Version
	}
	return &pub
}

// PinnedLib user type.
type PinnedLib struct {
	// The slugified name of the library
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// The version of the library. Can be latest
	Version *string `form:"version,omitempty" json:"version,omitempty" xml:"version,omitempty"`
}

// A sketch is a program intended to run on an arduino board.
// It's composed by a main .ino file and optional other files.
// You should upload only .ino and .h files.
type sketch struct {
	// Other files contained in the example
	Files []*filefull `form:"files,omitempty" json:"files,omitempty" xml:"files,omitempty"`
	// The main file
	Ino      *filefull `form:"ino,omitempty" json:"ino,omitempty" xml:"ino,omitempty"`
	Metadata *struct {
		IncludedLibs []*pinnedLib `form:"included_libs,omitempty" json:"included_libs,omitempty" xml:"included_libs,omitempty"`
	} `form:"metadata,omitempty" json:"metadata,omitempty" xml:"metadata,omitempty"`
	// The name of the sketch
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
}

// Validate validates the sketch type instance.
func (ut *sketch) Validate() (err error) {
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
		pub.Files = make([]*Filefull, len(ut.Files))
		for i2, elem2 := range ut.Files {
			pub.Files[i2] = elem2.Publicize()
		}
	}
	if ut.Ino != nil {
		pub.Ino = ut.Ino.Publicize()
	}
	if ut.Metadata != nil {
		pub.Metadata = &struct {
			IncludedLibs []*PinnedLib `form:"included_libs,omitempty" json:"included_libs,omitempty" xml:"included_libs,omitempty"`
		}{}
		if ut.Metadata.IncludedLibs != nil {
			pub.Metadata.IncludedLibs = make([]*PinnedLib, len(ut.Metadata.IncludedLibs))
			for i3, elem3 := range ut.Metadata.IncludedLibs {
				pub.Metadata.IncludedLibs[i3] = elem3.Publicize()
			}
		}
	}
	if ut.Name != nil {
		pub.Name = ut.Name
	}
	return &pub
}

// A sketch is a program intended to run on an arduino board.
// It's composed by a main .ino file and optional other files. You should upload only .ino and .h files.
type Sketch struct {
	// Other files contained in the example
	Files []*Filefull `form:"files,omitempty" json:"files,omitempty" xml:"files,omitempty"`
	// The main file
	Ino      *Filefull `form:"ino,omitempty" json:"ino,omitempty" xml:"ino,omitempty"`
	Metadata *struct {
		IncludedLibs []*PinnedLib `form:"included_libs,omitempty" json:"included_libs,omitempty" xml:"included_libs,omitempty"`
	} `form:"metadata,omitempty" json:"metadata,omitempty" xml:"metadata,omitempty"`
	// The name of the sketch
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
}

// Validate validates the Sketch type instance.
func (ut *Sketch) Validate() (err error) {
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
