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

package lib

import (
	"context"
	"fmt"
	"strings"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/pkg/errors"
)

type installedLib struct {
	Library   *libraries.Library
	Available *librariesindex.Release
}

// LibraryList FIXMEDOC
func LibraryList(ctx context.Context, req *rpc.LibraryListReq) (*rpc.LibraryListResp, error) {
	pm := commands.GetPackageManager(req.GetInstance().GetId())
	if pm == nil {
		return nil, errors.New("invalid instance")
	}

	lm := commands.GetLibraryManager(req.GetInstance().GetId())
	if lm == nil {
		return nil, errors.New("invalid instance")
	}

	nameFilter := strings.ToLower(req.GetName())

	instaledLib := []*rpc.InstalledLibrary{}
	res := listLibraries(lm, req.GetUpdatable(), req.GetAll())
	if len(res) > 0 {
		if f := req.GetFqbn(); f != "" {
			fqbn, err := cores.ParseFQBN(req.GetFqbn())
			if err != nil {
				return nil, fmt.Errorf("parsing fqbn: %s", err)
			}
			_, boardPlatform, _, _, refBoardPlatform, err := pm.ResolveFQBN(fqbn)
			if err != nil {
				return nil, fmt.Errorf("loading board data: %s", err)
			}

			filteredRes := map[string]*installedLib{}
			for _, lib := range res {
				if cp := lib.Library.ContainerPlatform; cp != nil {
					if cp != boardPlatform && cp != refBoardPlatform {
						// Filter all libraries from extraneous platforms
						continue
					}
				}
				if latest, has := filteredRes[lib.Library.Name]; has {
					if latest.Library.LocationPriorityFor(boardPlatform, refBoardPlatform) >= lib.Library.LocationPriorityFor(boardPlatform, refBoardPlatform) {
						continue
					}
				}
				filteredRes[lib.Library.Name] = lib
			}

			res = []*installedLib{}
			for _, lib := range filteredRes {
				res = append(res, lib)
			}
		}

		for _, lib := range res {
			if nameFilter != "" && strings.ToLower(lib.Library.Name) != nameFilter {
				continue
			}
			libtmp, err := GetOutputLibrary(lib.Library)
			if err != nil {
				return nil, err
			}
			release := GetOutputRelease(lib.Available)
			instaledLib = append(instaledLib, &rpc.InstalledLibrary{
				Library: libtmp,
				Release: release,
			})
		}

		return &rpc.LibraryListResp{InstalledLibrary: instaledLib}, nil
	}
	return &rpc.LibraryListResp{}, nil
}

// listLibraries returns the list of installed libraries. If updatable is true it
// returns only the libraries that may be updated.
func listLibraries(lm *librariesmanager.LibrariesManager, updatable bool, all bool) []*installedLib {
	res := []*installedLib{}
	for _, libAlternatives := range lm.Libraries {
		for _, lib := range libAlternatives.Alternatives {
			if !all {
				if lib.Location != libraries.User {
					continue
				}
			}
			available := lm.Index.FindLibraryUpdate(lib)
			if updatable && available == nil {
				continue
			}
			res = append(res, &installedLib{
				Library:   lib,
				Available: available,
			})
		}
	}
	return res
}

// GetOutputLibrary FIXMEDOC
func GetOutputLibrary(lib *libraries.Library) (*rpc.Library, error) {
	insdir := ""
	if lib.InstallDir != nil {
		insdir = lib.InstallDir.String()
	}
	srcdir := ""
	if lib.SourceDir != nil {
		srcdir = lib.SourceDir.String()
	}
	utldir := ""
	if lib.UtilityDir != nil {
		utldir = lib.UtilityDir.String()
	}
	cntplat := ""
	if lib.ContainerPlatform != nil {
		cntplat = lib.ContainerPlatform.String()
	}

	return &rpc.Library{
		Name:              lib.Name,
		Author:            lib.Author,
		Maintainer:        lib.Maintainer,
		Sentence:          lib.Sentence,
		Paragraph:         lib.Paragraph,
		Website:           lib.Website,
		Category:          lib.Category,
		Architectures:     lib.Architectures,
		Types:             lib.Types,
		InstallDir:        insdir,
		SourceDir:         srcdir,
		UtilityDir:        utldir,
		Location:          lib.Location.ToRPCLibraryLocation(),
		ContainerPlatform: cntplat,
		Layout:            lib.Layout.ToRPCLibraryLayout(),
		RealName:          lib.RealName,
		DotALinkage:       lib.DotALinkage,
		Precompiled:       lib.Precompiled,
		LdFlags:           lib.LDflags,
		IsLegacy:          lib.IsLegacy,
		Version:           lib.Version.String(),
		License:           lib.License,
		Examples:          lib.Examples.AsStrings(),
		ProvidesIncludes:  lib.DeclaredHeaders(),
	}, nil
}

// GetOutputRelease FIXMEDOC
func GetOutputRelease(lib *librariesindex.Release) *rpc.LibraryRelease { //
	if lib != nil {
		return &rpc.LibraryRelease{
			Author:        lib.Author,
			Version:       lib.Version.String(),
			Maintainer:    lib.Maintainer,
			Sentence:      lib.Sentence,
			Paragraph:     lib.Paragraph,
			Website:       lib.Website,
			Category:      lib.Category,
			Architectures: lib.Architectures,
			Types:         lib.Types,
		}
	}
	return &rpc.LibraryRelease{}
}
