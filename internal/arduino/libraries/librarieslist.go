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

package libraries

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sort"

	"github.com/arduino/go-paths-helper"
	semver "go.bug.st/relaxed-semver"
)

// List is a list of Libraries
type List []*Library

// Contains check if a lib is contained in the list
func (list *List) Contains(lib *Library) bool {
	for _, l := range *list {
		if l == lib {
			return true
		}
	}
	return false
}

// Add appends all libraries passed as parameter in the list
func (list *List) Add(libs ...*Library) {
	for _, lib := range libs {
		*list = append(*list, lib)
	}
}

func (list *List) binaryMagicNumber() uint32 {
	return 0xAD000001
}

func (list *List) UnmarshalBinary(in io.Reader, prefix *paths.Path) error {
	var magic uint32
	if err := binary.Read(in, binary.NativeEndian, &magic); err != nil {
		return err
	}
	if magic != list.binaryMagicNumber() {
		return errors.New("invalid cache version")
	}
	var n int32
	if err := binary.Read(in, binary.NativeEndian, &n); err != nil {
		return err
	}
	res := make([]*Library, n)
	for i := range res {
		var lib Library
		if err := lib.UnmarshalBinary(in, prefix); err != nil {
			return err
		}
		res[i] = &lib
	}
	*list = res
	return nil
}

func (list *List) MarshalBinary(out io.Writer, prefix *paths.Path) error {
	if err := binary.Write(out, binary.NativeEndian, list.binaryMagicNumber()); err != nil {
		return err
	}
	if err := binary.Write(out, binary.NativeEndian, int32(len(*list))); err != nil {
		return err
	}
	for _, lib := range *list {
		if err := lib.MarshalBinary(out, prefix); err != nil {
			return fmt.Errorf("could not encode lib data of %s: %w", lib.InstallDir.String(), err)
		}
	}
	return nil
}

// Remove removes the given library from the list
func (list *List) Remove(libraryToRemove *Library) {
	for i, lib := range *list {
		if lib == libraryToRemove {
			*list = append((*list)[:i], (*list)[i+1:]...)
			return
		}
	}
}

// FindByName returns the first library in the list that match
// the specified name or nil if not found
func (list *List) FindByName(name string) *Library {
	for _, lib := range *list {
		if lib.Name == name {
			return lib
		}
	}
	return nil
}

// FilterByVersionAndInstallLocation returns the libraries matching the provided version and install location. If version
// is nil all version are matched.
func (list *List) FilterByVersionAndInstallLocation(version *semver.Version, installLocation LibraryLocation) List {
	var found List
	for _, lib := range *list {
		if lib.Location != installLocation {
			continue
		}
		if version != nil && !lib.Version.Equal(version) {
			continue
		}
		found.Add(lib)
	}
	return found
}

// SortByName sorts the libraries by name
func (list *List) SortByName() {
	sort.Slice(*list, func(i, j int) bool {
		a, b := (*list)[i], (*list)[j]
		return a.Name < b.Name
	})
}

/*
// HasHigherPriority returns true if library x has higher priority compared to library
// y for the given header and architecture.
func HasHigherPriority(libX, libY *Library, header string, arch string) bool {
	//return computePriority(libX, header, arch) > computePriority(libY, header, arch)
	header = strings.TrimSuffix(header, filepath.Ext(header))

	simplify := func(name string) string {
		name = utils.SanitizeName(name)
		name = strings.ToLower(name)
		return name
	}
	header = simplify(header)
	nameX := simplify(libX.Name)
	nameY := simplify(libY.Name)

	compareLocations := func() bool {
		// XXX: priority inversion case.
		if libX.Location < libY.Location {
			return true
		}
		return false
	}

	checks := []func(name, header string) bool{
		func(name, header string) bool { return name == header },
		func(name, header string) bool { return name == header+"-master" },
		strings.HasPrefix,
		strings.HasSuffix,
		strings.Contains,
	}
	// Run all checks to sort priorities based on library name
	// If both library match the same name check, then fallback to
	// compare locations
	for _, check := range checks {
		x := check(nameX, header)
		y := check(nameY, header)
		if x && y {
			return compareLocations()
		}
		if x {
			return true
		}
		if y {
			return false
		}
	}

	return compareLocations()
}
*/
