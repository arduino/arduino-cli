/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO AG (http://www.arduino.cc/)
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
 */

package librariesresolver

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bcmi-labs/arduino-cli/arduino/libraries"
	"github.com/bcmi-labs/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/bcmi-labs/arduino-cli/arduino/utils"
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
	cppHeaders, err := lib.SrcFolder.ReadDir()
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

	priority := int(lib.PriorityForArchitecture(arch)) // bewteen 0..255
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
