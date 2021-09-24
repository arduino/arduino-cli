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

package globals

var (
	empty struct{}

	// MainFileValidExtension is the extension that must be used for files in new sketches
	MainFileValidExtension string = ".ino"

	// MainFileValidExtensions lists valid extensions for a sketch file
	MainFileValidExtensions = map[string]struct{}{
		MainFileValidExtension: empty,
		// .pde extension is deprecated and must not be used for new sketches
		".pde": empty,
	}

	// AdditionalFileValidExtensions lists any file extension the builder considers as valid
	AdditionalFileValidExtensions = map[string]struct{}{
		".h":    empty,
		".c":    empty,
		".hpp":  empty,
		".hh":   empty,
		".cpp":  empty,
		".S":    empty,
		".adoc": empty,
		".md":   empty,
		".json": empty,
		".tpp":  empty,
		".ipp":  empty,
	}

	// SourceFilesValidExtensions lists valid extensions for source files (no headers)
	SourceFilesValidExtensions = map[string]struct{}{
		".c":   empty,
		".cpp": empty,
		".S":   empty,
	}

	// HeaderFilesValidExtensions lists valid extensions for header files
	HeaderFilesValidExtensions = map[string]struct{}{
		".h":   empty,
		".hpp": empty,
		".hh":  empty,
	}
)
