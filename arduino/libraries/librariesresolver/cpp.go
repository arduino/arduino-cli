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

package librariesresolver

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/arduino/utils"
	"github.com/sirupsen/logrus"
)

// Cpp finds libraries made for the C++ language
type Cpp struct {
	headers map[string]libraries.List
}

// NewCppResolver creates a new Cpp resolver
func NewCppResolver() *Cpp {
	return &Cpp{
		headers: map[string]libraries.List{},
	}
}

// ScanFromLibrariesManager reads all librariers loaded in the LibrariesManager to find
// and cache all C++ headers for later retrieval
func (resolver *Cpp) ScanFromLibrariesManager(lm *librariesmanager.LibrariesManager) error {
	for _, libAlternatives := range lm.Libraries {
		for _, lib := range libAlternatives.Alternatives {
			resolver.ScanLibrary(lib)
		}
	}
	return nil
}

// ScanLibrary reads a library to find and cache C++ headers for later retrieval
func (resolver *Cpp) ScanLibrary(lib *libraries.Library) error {
	cppHeaders, err := lib.SourceDir.ReadDir()
	if err != nil {
		return fmt.Errorf("reading lib src dir: %s", err)
	}
	cppHeaders.FilterSuffix(".h", ".hpp", ".hh")
	for _, cppHeaderPath := range cppHeaders {
		cppHeader := cppHeaderPath.Base()
		l := resolver.headers[cppHeader]
		l.Add(lib)
		resolver.headers[cppHeader] = l
	}
	return nil
}

// AlternativesFor returns all the libraries that provides the specified header
func (resolver *Cpp) AlternativesFor(header string) libraries.List {
	fmt.Printf("Alternatives for %s: %s\n", header, resolver.headers[header])
	return resolver.headers[header]
}

// ResolveFor finds the most suitable library for the specified combination of
// header and architecture. If no libraries provides the requested header, nil is returned
func (resolver *Cpp) ResolveFor(header, architecture string) *libraries.Library {
	logrus.Infof("Resolving include %s for arch %s", header, architecture)
	var found *libraries.Library
	var foundPriority int
	for _, lib := range resolver.headers[header] {
		libPriority := computePriority(lib, header, architecture)
		msg := "  discarded"
		if found == nil || foundPriority < libPriority {
			found = lib
			foundPriority = libPriority
			msg = "  found better lib"
		}
		logrus.
			WithField("lib", lib.Name).
			WithField("prio", fmt.Sprintf("%03X", libPriority)).
			Infof(msg)
	}
	return found
}

func computePriority(lib *libraries.Library, header, arch string) int {
	simplify := func(name string) string {
		name = utils.SanitizeName(name)
		name = strings.ToLower(name)
		return name
	}

	header = strings.TrimSuffix(header, filepath.Ext(header))
	header = simplify(header)
	name := simplify(lib.Name)

	priority := int(lib.PriorityForArchitecture(arch)) // between 0..255
	if name == header {
		priority += 0x500
	} else if name == header+"-master" {
		priority += 0x400
	} else if strings.HasPrefix(name, header) {
		priority += 0x300
	} else if strings.HasSuffix(name, header) {
		priority += 0x200
	} else if strings.Contains(name, header) {
		priority += 0x100
	}
	return priority
}
