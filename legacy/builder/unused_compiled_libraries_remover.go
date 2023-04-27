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

package builder

import (
	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
)

type UnusedCompiledLibrariesRemover struct{}

func (s *UnusedCompiledLibrariesRemover) Run(ctx *types.Context) error {
	librariesBuildPath := ctx.LibrariesBuildPath

	if librariesBuildPath.NotExist() {
		return nil
	}

	libraryNames := toLibraryNames(ctx.ImportedLibraries)

	files, err := librariesBuildPath.ReadDir()
	if err != nil {
		return errors.WithStack(err)
	}
	for _, file := range files {
		if file.IsDir() {
			if !slices.Contains(libraryNames, file.Base()) {
				if err := file.RemoveAll(); err != nil {
					return errors.WithStack(err)
				}
			}
		}
	}

	return nil
}

func toLibraryNames(libraries []*libraries.Library) []string {
	libraryNames := []string{}
	for _, library := range libraries {
		libraryNames = append(libraryNames, library.Name)
	}
	return libraryNames
}
