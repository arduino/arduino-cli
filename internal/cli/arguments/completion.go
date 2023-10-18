// This file is part of arduino-cli.
//
// Copyright 2023 ARDUINO SA (http://www.arduino.cc/)
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

package arguments

import (
	"context"

	"github.com/arduino/arduino-cli/commands/board"
	"github.com/arduino/arduino-cli/commands/core"
	"github.com/arduino/arduino-cli/commands/lib"
	"github.com/arduino/arduino-cli/commands/upload"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// GetInstalledBoards is an helper function useful to autocomplete.
// It returns a list of fqbn
// it's taken from cli/board/listall.go
func GetInstalledBoards() []string {
	inst := instance.CreateAndInit()

	list, _ := board.ListAll(context.Background(), &rpc.BoardListAllRequest{
		Instance:            inst,
		SearchArgs:          nil,
		IncludeHiddenBoards: false,
	})
	var res []string
	// transform the data structure for the completion
	for _, i := range list.Boards {
		res = append(res, i.Fqbn+"\t"+i.Name)
	}
	return res
}

// GetInstalledProgrammers is an helper function useful to autocomplete.
// It returns a list of programmers available based on the installed boards
func GetInstalledProgrammers() []string {
	inst := instance.CreateAndInit()

	// we need the list of the available fqbn in order to get the list of the programmers
	listAllReq := &rpc.BoardListAllRequest{
		Instance:            inst,
		SearchArgs:          nil,
		IncludeHiddenBoards: false,
	}
	list, _ := board.ListAll(context.Background(), listAllReq)

	installedProgrammers := make(map[string]string)
	for _, board := range list.Boards {
		programmers, _ := upload.ListProgrammersAvailableForUpload(context.Background(), &rpc.ListProgrammersAvailableForUploadRequest{
			Instance: inst,
			Fqbn:     board.Fqbn,
		})
		for _, programmer := range programmers.GetProgrammers() {
			installedProgrammers[programmer.GetId()] = programmer.GetName()
		}
	}

	res := make([]string, len(installedProgrammers))
	i := 0
	for programmerID := range installedProgrammers {
		res[i] = programmerID + "\t" + installedProgrammers[programmerID]
		i++
	}
	return res
}

// GetUninstallableCores is an helper function useful to autocomplete.
// It returns a list of cores which can be uninstalled
func GetUninstallableCores() []string {
	inst := instance.CreateAndInit()

	platforms, _ := core.PlatformList(&rpc.PlatformListRequest{
		Instance:      inst,
		UpdatableOnly: false,
		All:           false,
	})
	var res []string
	// transform the data structure for the completion
	for _, i := range platforms.InstalledPlatforms {
		res = append(res, i.Id+"\t"+i.Name)
	}
	return res
}

// GetInstallableCores is an helper function useful to autocomplete.
// It returns a list of cores which can be installed/downloaded
func GetInstallableCores() []string {
	inst := instance.CreateAndInit()

	platforms, _ := core.PlatformSearch(&rpc.PlatformSearchRequest{
		Instance:    inst,
		SearchArgs:  "",
		AllVersions: false,
	})
	var res []string
	// transform the data structure for the completion
	for _, i := range platforms.SearchOutput {
		res = append(res, i.Id+"\t"+i.Name)
	}
	return res
}

// GetInstalledLibraries is an helper function useful to autocomplete.
// It returns a list of libs which are currently installed, including the builtin ones
func GetInstalledLibraries() []string {
	return getLibraries(true)
}

// GetUninstallableLibraries is an helper function useful to autocomplete.
// It returns a list of libs which can be uninstalled
func GetUninstallableLibraries() []string {
	return getLibraries(false)
}

func getLibraries(all bool) []string {
	inst := instance.CreateAndInit()
	libs, _ := lib.LibraryList(context.Background(), &rpc.LibraryListRequest{
		Instance:  inst,
		All:       all,
		Updatable: false,
		Name:      "",
		Fqbn:      "",
	})
	var res []string
	// transform the data structure for the completion
	for _, i := range libs.InstalledLibraries {
		res = append(res, i.Library.Name+"\t"+i.Library.Sentence)
	}
	return res
}

// GetInstallableLibs is an helper function useful to autocomplete.
// It returns a list of libs which can be installed/downloaded
func GetInstallableLibs() []string {
	inst := instance.CreateAndInit()

	libs, _ := lib.LibrarySearch(context.Background(), &rpc.LibrarySearchRequest{
		Instance:   inst,
		SearchArgs: "", // if no query is specified all the libs are returned
	})
	var res []string
	// transform the data structure for the completion
	for _, i := range libs.Libraries {
		res = append(res, i.Name+"\t"+i.Latest.Sentence)
	}
	return res
}

// GetAvailablePorts is an helper function useful to autocomplete.
// It returns a list of upload port of the boards which are currently connected.
// It will not suggests network ports because the timeout is not set.
func GetAvailablePorts() []*rpc.Port {
	inst := instance.CreateAndInit()

	list, _, _ := board.List(&rpc.BoardListRequest{
		Instance: inst,
	})
	var res []*rpc.Port
	// transform the data structure for the completion
	for _, i := range list {
		res = append(res, i.Port)
	}
	return res
}
