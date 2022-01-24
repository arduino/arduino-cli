package arguments

import (
	"context"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/board"
	"github.com/arduino/arduino-cli/commands/core"
	"github.com/arduino/arduino-cli/commands/lib"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// GetInstalledBoards is an helper function useful to autocomplete.
// It returns a list of fqbn
// it's taken from cli/board/listall.go
func GetInstalledBoards() []string {
	instance.Init()
	inst := instance.Get()

	list, _ := board.ListAll(context.Background(), &rpc.BoardListAllRequest{
		Instance:            inst.ToRPC(),
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

// GetInstalledProtocols is an helper function useful to autocomplete.
// It returns a list of protocols available based on the installed boards
func GetInstalledProtocols() []string {
	instance.Init()
	inst := instance.Get()
	pm := commands.GetPackageManager(inst.ID())
	boards := pm.InstalledBoards()

	installedProtocols := make(map[string]struct{})
	for _, board := range boards {
		for _, protocol := range board.Properties.SubTree("upload.tool").FirstLevelKeys() {
			if protocol == "default" {
				// default is used as fallback when trying to upload to a board
				// using a protocol not defined for it, it's useless showing it
				// in autocompletion
				continue
			}
			installedProtocols[protocol] = struct{}{}
		}
	}
	res := make([]string, len(installedProtocols))
	i := 0
	for k := range installedProtocols {
		res[i] = k
		i++
	}
	return res
}

// GetInstalledProgrammers is an helper function useful to autocomplete.
// It returns a list of programmers available based on the installed boards
func GetInstalledProgrammers() []string {
	instance.Init()
	inst := instance.Get()
	pm := commands.GetPackageManager(inst.ID())

	// we need the list of the available fqbn in order to get the list of the programmers
	list, _ := board.ListAll(context.Background(), &rpc.BoardListAllRequest{
		Instance:            inst.ToRPC(),
		SearchArgs:          nil,
		IncludeHiddenBoards: false,
	})

	installedProgrammers := make(map[string]string)
	for _, board := range list.Boards {
		fqbn, _ := cores.ParseFQBN(board.Fqbn)
		_, boardPlatform, _, _, _, _ := pm.ResolveFQBN(fqbn)
		for programmerID, programmer := range boardPlatform.Programmers {
			installedProgrammers[programmerID] = programmer.Name
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
	instance.Init()
	inst := instance.Get()

	platforms, _ := core.GetPlatforms(&rpc.PlatformListRequest{
		Instance:      inst.ToRPC(),
		UpdatableOnly: false,
		All:           false,
	})
	var res []string
	// transform the data structure for the completion
	for _, i := range platforms {
		res = append(res, i.Id+"\t"+i.Name)
	}
	return res
}

// GetInstallableCores is an helper function useful to autocomplete.
// It returns a list of cores which can be installed/downloaded
func GetInstallableCores() []string {
	instance.Init()
	inst := instance.Get()

	platforms, _ := core.PlatformSearch(&rpc.PlatformSearchRequest{
		Instance:    inst.ToRPC(),
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
	instance.Init()
	inst := instance.Get()

	libs, _ := lib.LibraryList(context.Background(), &rpc.LibraryListRequest{
		Instance:  inst.ToRPC(),
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
	instance.Init()
	inst := instance.Get()

	libs, _ := lib.LibrarySearch(context.Background(), &rpc.LibrarySearchRequest{
		Instance: inst.ToRPC(),
		Query:    "", // if no query is specified all the libs are returned
	})
	var res []string
	// transform the data structure for the completion
	for _, i := range libs.Libraries {
		res = append(res, i.Name+"\t"+i.Latest.Sentence)
	}
	return res
}

// GetConnectedBoards is an helper function useful to autocomplete.
// It returns a list of boards which are currently connected
// Obviously it does not suggests network ports because of the timeout
func GetConnectedBoards() []string {
	instance.Init()
	inst := instance.Get()

	list, _ := board.List(&rpc.BoardListRequest{
		Instance: inst.ToRPC(),
	})
	var res []string
	// transform the data structure for the completion
	for _, i := range list {
		res = append(res, i.Port.Address)
	}
	return res
}
