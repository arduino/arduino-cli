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

import (
	"net/url"

	"github.com/arduino/arduino-cli/internal/arduino/resources"
)

var (
	// MainFileValidExtension is the extension that must be used for files in new sketches
	MainFileValidExtension = ".ino"

	// MainFileValidExtensions lists valid extensions for a sketch file
	MainFileValidExtensions = map[string]bool{
		MainFileValidExtension: true,
		// .pde extension is deprecated and must not be used for new sketches
		".pde": true,
	}

	// AdditionalFileValidExtensions lists any file extension the builder considers as valid
	AdditionalFileValidExtensions = map[string]bool{
		".h":    true,
		".c":    true,
		".hpp":  true,
		".hh":   true,
		".cpp":  true,
		".cxx":  true,
		".cc":   true,
		".S":    true,
		".adoc": true,
		".md":   true,
		".json": true,
		".tpp":  true,
		".ipp":  true,
	}

	// SourceFilesValidExtensions lists valid extensions for source files (no headers).
	// If a platform do not provide a compile recipe for a specific file extension, this
	// map provides the equivalent extension to use as a fallback.
	SourceFilesValidExtensions = map[string]string{
		".c":   "",
		".cpp": "",
		".cxx": ".cpp",
		".cc":  ".cpp",
		".S":   "",
	}

	// HeaderFilesValidExtensions lists valid extensions for header files
	HeaderFilesValidExtensions = map[string]bool{
		".h":   true,
		".hpp": true,
		".hh":  true,
	}

	// DefaultIndexURL is the default index url
	DefaultIndexURL = "https://downloads.arduino.cc/packages/package_index.tar.bz2"

	// LibrariesIndexURL is the URL where to get the libraries index.
	LibrariesIndexURL, _ = url.Parse("https://downloads.arduino.cc/libraries/library_index.tar.bz2")

	// LibrariesIndexResource is the IndexResource to get the libraries index.
	LibrariesIndexResource = resources.IndexResource{
		URL:                          LibrariesIndexURL,
		EnforceSignatureVerification: true,
	}
)
