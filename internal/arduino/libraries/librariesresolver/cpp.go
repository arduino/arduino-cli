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

package librariesresolver

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/arduino/arduino-cli/internal/arduino/cores"
	"github.com/arduino/arduino-cli/internal/arduino/libraries"
	"github.com/arduino/arduino-cli/internal/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/internal/arduino/utils"
	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/schollz/closestmatch"
	"github.com/sirupsen/logrus"
)

// Cpp finds libraries made for the C++ language
type Cpp struct {
	headers map[string]libraries.List
}

var tr = i18n.Tr

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
		for _, lib := range libAlternatives {
			resolver.ScanLibrary(lib)
		}
	}
	return nil
}

// ScanIDEBuiltinLibraries reads ide-builtin librariers loaded in the LibrariesManager to find
// and cache all C++ headers for later retrieval.
func (resolver *Cpp) ScanIDEBuiltinLibraries(lm *librariesmanager.LibrariesManager) error {
	for _, libAlternatives := range lm.Libraries {
		for _, lib := range libAlternatives {
			if lib.Location == libraries.IDEBuiltIn {
				resolver.ScanLibrary(lib)
			}
		}
	}
	return nil
}

// ScanUserAndUnmanagedLibraries reads user/unmanaged librariers loaded in the LibrariesManager to find
// and cache all C++ headers for later retrieval.
func (resolver *Cpp) ScanUserAndUnmanagedLibraries(lm *librariesmanager.LibrariesManager) error {
	for _, libAlternatives := range lm.Libraries {
		for _, lib := range libAlternatives {
			if lib.Location == libraries.User || lib.Location == libraries.Unmanaged {
				resolver.ScanLibrary(lib)
			}
		}
	}
	return nil
}

// ScanPlatformLibraries reads platform-bundled libraries for a specific platform loaded in the LibrariesManager
// to find and cache all C++ headers for later retrieval.
func (resolver *Cpp) ScanPlatformLibraries(lm *librariesmanager.LibrariesManager, platform *cores.PlatformRelease) error {
	for _, libAlternatives := range lm.Libraries {
		for _, lib := range libAlternatives {
			if lib.Location != libraries.PlatformBuiltIn && lib.Location != libraries.ReferencedPlatformBuiltIn {
				continue
			}
			if lib.ContainerPlatform != platform {
				continue
			}
			resolver.ScanLibrary(lib)
		}
	}
	return nil
}

// ScanLibrary reads a library to find and cache C++ headers for later retrieval
func (resolver *Cpp) ScanLibrary(lib *libraries.Library) error {
	cppHeaders, err := lib.SourceHeaders()
	if err != nil {
		return fmt.Errorf(tr("reading lib headers: %s"), err)
	}
	for _, cppHeader := range cppHeaders {
		l := resolver.headers[cppHeader]
		l.Add(lib)
		resolver.headers[cppHeader] = l
	}
	return nil
}

// AlternativesFor returns all the libraries that provides the specified header
func (resolver *Cpp) AlternativesFor(header string) libraries.List {
	return resolver.headers[header]
}

// ResolveFor finds the most suitable library for the specified combination of
// header and architecture. If no libraries provides the requested header, nil is returned
func (resolver *Cpp) ResolveFor(header, architecture string) *libraries.Library {
	logrus.Infof("Resolving include %s for arch %s", header, architecture)
	var found libraries.List
	var foundPriority int
	for _, lib := range resolver.headers[header] {
		libPriority := ComputePriority(lib, header, architecture)
		msg := "  discarded"
		if found == nil || foundPriority < libPriority {
			found = libraries.List{}
			found.Add(lib)
			foundPriority = libPriority
			msg = "  found better lib"
		} else if foundPriority == libPriority {
			found.Add(lib)
			msg = "  found another lib with same priority"
		}
		logrus.
			WithField("lib", lib.Name).
			WithField("prio", fmt.Sprintf("%03X", libPriority)).
			Infof(msg)
	}
	if found == nil {
		return nil
	}
	if len(found) == 1 {
		return found[0]
	}

	// If more than one library qualifies use the "closestmatch" algorithm to
	// find the best matching one (instead of choosing it randomly)
	if best := findLibraryWithNameBestDistance(header, found); best != nil {
		logrus.WithField("lib", best.Name).Info("  library with the best matching name")
		return best
	}

	found.SortByName()
	logrus.WithField("lib", found[0].Name).Info("  first library in alphabetic order")
	return found[0]
}

func simplify(name string) string {
	name = utils.SanitizeName(name)
	name = strings.ToLower(name)
	return name
}

// ComputePriority returns an integer value representing the priority of the library
// for the specified header and architecture. The higher the value, the higher the
// priority.
func ComputePriority(lib *libraries.Library, header, arch string) int {
	header = strings.TrimSuffix(header, filepath.Ext(header))
	header = simplify(header)
	name := simplify(lib.Name)
	dirName := simplify(lib.DirName)

	priority := 0

	// Bonus for core-optimized libraries
	if lib.IsOptimizedForArchitecture(arch) {
		// give a slightly better bonus for libraries that have specific optimization
		// (it is more important than Location but less important than Name)
		priority += 1010
	} else if lib.IsArchitectureIndependent() {
		// standard bonus for architecture independent (vanilla) libraries
		priority += 1000
	} else {
		// the library is not architecture compatible
		priority += 0
	}

	if name == header && dirName == header {
		priority += 700
	} else if name == header || dirName == header {
		priority += 600
	} else if name == header+"-main" || dirName == header+"-main" {
		priority += 500
	} else if name == header+"-master" || dirName == header+"-master" {
		priority += 400
	} else if strings.HasPrefix(name, header) || strings.HasPrefix(dirName, header) {
		priority += 300
	} else if strings.HasSuffix(name, header) || strings.HasSuffix(dirName, header) {
		priority += 200
	} else if strings.Contains(name, header) || strings.Contains(dirName, header) {
		priority += 100
	}

	switch lib.Location {
	case libraries.IDEBuiltIn:
		priority += 0
	case libraries.ReferencedPlatformBuiltIn:
		priority++
	case libraries.PlatformBuiltIn:
		priority += 2
	case libraries.User:
		priority += 3
	case libraries.Unmanaged:
		// Bonus for libraries specified via --libraries flags, those libraries gets the highest priority
		priority += 10000
	default:
		panic(fmt.Sprintf("Invalid library location: %d", lib.Location))
	}
	return priority
}

func findLibraryWithNameBestDistance(name string, libs libraries.List) *libraries.Library {
	// Create closestmatch DB
	wordsToTest := []string{}
	for _, lib := range libs {
		wordsToTest = append(wordsToTest, simplify(lib.Name))
	}
	// Choose a set of bag sizes, more is more accurate but slower
	bagSizes := []int{2}

	// Create a closestmatch object and find the best matching name
	cm := closestmatch.New(wordsToTest, bagSizes)
	closestName := cm.Closest(name)

	// Return the closest-matching lib
	var winner *libraries.Library
	for _, lib := range libs {
		if closestName == simplify(lib.Name) {
			winner = lib
			break
		}
	}
	return winner
}
